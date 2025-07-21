package src

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/asr"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/llm"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/repo"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/tts"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/types"

	"github.com/go-playground/validator/v10"
	"github.com/hertz-contrib/websocket"
	"github.com/rs/zerolog/log"
)

type ClientMessageHandler func([]byte) error

type Session struct {
	conn *websocket.Conn
	hub  *Hub

	deviceId        string `validate:"required"`
	clientId        string `validate:"required"`
	bearerToken     string
	sessionId       string `validate:"required"`
	protocolVersion string

	device *types.Device

	// populated in hello message
	deviceVersion     int32
	deviceAudioParams HelloAudioParams
	deviceSupportMCP  bool

	// populate in listen start message
	deviceAudioMode AudioMode

	state *SessionState

	asrProcessor *AsrProcessor
	llmProcessor *LlmProcessor
	ttsProcessor *TtsProcessor

	msgHandlers map[MessageType]ClientMessageHandler
	ctx         context.Context
	cancel      context.CancelFunc
}

func newSession(ctx context.Context) *Session {
	s := &Session{
		msgHandlers: make(map[MessageType]ClientMessageHandler),
	}

	s.resetState(AudioModeNone)

	s.msgHandlers[MessageTypeRawAudio] = s.handleAudio
	s.msgHandlers[MessageTypeHello] = s.handleHello
	s.msgHandlers[MessageTypeListenStart] = s.handleListenStart
	s.msgHandlers[MessageTypeListenStop] = s.handleListenStop
	s.msgHandlers[MessageTypeListenDetect] = s.handleListenDetect

	s.ctx, s.cancel = context.WithCancel(ctx)

	return s
}

func (s *Session) isValid() bool {
	return validator.New().Struct(s) == nil
}

// load device object from repository
func (s *Session) populateDevice() error {
	var err error
	s.device, err = s.hub.repo.FindDevice(repo.WhereCondition{
		"device_id": s.deviceId,
	})
	return err
}

func (s *Session) loop() error {
	var (
		err      error
		mt       int
		rawBytes []byte
	)

	defer func() {
		if (s.ctx.Err() != nil || err != nil) && s.cancel != nil {
			s.cancel()
		}
	}()

	asrResponseCh := make(chan *asr.AsrResponse, 10) // buffered channel for ASR responses
	s.asrProcessor, err = NewAsrProcessor(s.ctx, s.hub.cfgAsr, asrResponseCh)
	if err != nil {
		return err
	}
	llmResponseCh := make(chan *llm.LLMResponse, 10) // buffered channel for LLM responses
	s.llmProcessor = NewLlmProcessor(s.ctx, s.hub.cfgLlm.Deepseek)
	ttsResponseCh := make(chan *tts.TTSResponse, 10) // buffered channel for TTS responses
	s.ttsProcessor = NewTtsProcessor(s.ctx, s.hub.cfgTts.CosyVoice)

	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		case r := <-asrResponseCh:
			if r.IsFinish && len(r.Text) != 0 {
				if err := s.cmdSTT(r.Text); err != nil {
					log.Error().Err(err).Msgf("Failed to send STT command for device %s: %v", s.deviceId, err)
					return err
				}

				go func() {
					resp, err := s.llmProcessor.Push(r.Text)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to ask conversation for device %s: %v", s.deviceId, err)
						llmResponseCh <- &llm.LLMResponse{
							Question: r.Text,
							Answer:   "",
							Err:      err,
						}
						return
					}

					llmResponseCh <- &llm.LLMResponse{
						Question: r.Text,
						Answer:   resp,
						Err:      nil,
					}
				}()

				if err := s.cmdEmotion("thinking"); err != nil {
					log.Error().Err(err).Msgf("Failed to send emotion command for device %s: %v", s.deviceId, err)
					return err
				}
			}

		case r := <-llmResponseCh:
			fmt.Printf("LLM response received: %s", r.Answer)
			if err := s.cmdTTSSentenceStart(r.Answer); err != nil {
				return err
			}

			go func() {
				resp, err := s.ttsProcessor.Push(r.Answer)
				if err != nil {
					ttsResponseCh <- &tts.TTSResponse{
						Text:  r.Answer,
						Audio: nil,
						Err:   err,
					}
				}

				opusData := make([]byte, 0)
				for _, opus := range resp {
					opusData = append(opusData, opus...)
				}

				os.WriteFile("/tmp/opus.opus", opusData, 0644)

				for i, opus := range resp {
					ttsResponseCh <- &tts.TTSResponse{
						IsEnd:   i == len(resp)-1,
						IsStart: i == 0,
						Text:    r.Answer,
						Audio:   opus,
						Err:     nil,
					}
				}
			}()

		case r, ok := <-ttsResponseCh:
			if !ok {
				return nil
			}

			if r.Err != nil {
				return fmt.Errorf("failed to process TTS response: %w", r.Err)
			}

			if r.IsStart {
				if err := s.cmdTTSStart(); err != nil {
					return err
				}

				if err := s.cmdEmotion("happy"); err != nil {
					return err
				}
			}

			println(len(r.Audio), r.IsEnd, r.Text)
			if err := s.cmdAudio(r.Audio); err != nil {
				return err
			}

			if r.IsEnd {
				if err := s.cmdTTSStop(); err != nil {
					return err
				}
			}

		default:
			s.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 50))
			mt, rawBytes, err = s.conn.ReadMessage()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					continue // ignore timeout errors
				}

				log.Error().Err(err).Msgf("Failed to read message from device %s: %v", s.deviceId, err)
				return err
			}

			var messagePayloadType MessageType = MessageTypeNone
			if mt == websocket.TextMessage {
				log.Debug().Msgf("XZ -> Server[T] %s: %s", s.deviceId, string(rawBytes))

				var meta *MetaMessage
				meta, err = MessageFromBytes[MetaMessage](rawBytes)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to parse message from device %s: %v, content is %s", s.deviceId, err, string(rawBytes))
					return err
				}
				messagePayloadType = meta.MessageType()
			}

			if mt == websocket.BinaryMessage {
				// log.Debug().Msgf("XZ -> Server[B] %s len(message) is %d", s.deviceId, len(rawBytes))

				messagePayloadType = MessageTypeRawAudio
			}

			handler, ok := s.msgHandlers[messagePayloadType]
			if !ok {
				log.Error().Msgf("No handler found for message type %s from device %s", messagePayloadType, s.deviceId)
				return fmt.Errorf("no handler found for message type %s from device %s", messagePayloadType, s.deviceId)
			}

			if err = handler(rawBytes); err != nil {
				log.Error().Err(err).Msgf("Failed to handle message type %s from device %s: %v", messagePayloadType, s.deviceId, err)
				return fmt.Errorf("failed to handle message type %s from device %s: %v", messagePayloadType, s.deviceId, err)
			}
		}
	}
}

// TODO
func (s *Session) isAuthenticated() bool {
	return true
}

// TODO
func (s *Session) isAuthorized() bool {
	return true
}

func (s *Session) isSessionIdMatch(sessionId string) bool {
	if len(sessionId) == 0 {
		return false
	}

	return s.sessionId == sessionId
}

func (s *Session) resetState(newState AudioMode) error {
	s.deviceAudioMode = newState
	if s.deviceAudioMode == AudioModeNone {
		s.state = newSessionState(s, kSessionStateIdle)
	}

	// https://github.com/78/xiaozhi-esp32/blob/main/docs/websocket.md
	if s.deviceAudioMode == AudioModeAuto {
		s.state = newSessionState(s, kSessionStateListening)
		s.state.ValidTransitions[kSessionStateIdle] = []SessionStateKind{kSessionStateConnecting}
		s.state.ValidTransitions[kSessionStateConnecting] = []SessionStateKind{kSessionStateListening, kSessionStateIdle}
		s.state.ValidTransitions[kSessionStateListening] = []SessionStateKind{kSessionStateSpeaking, kSessionStateIdle}
		s.state.ValidTransitions[kSessionStateSpeaking] = []SessionStateKind{kSessionStateListening, kSessionStateIdle}

		s.state.OnEnterCallbacks[kSessionStateIdle] = []TransitionCallback{
			logTransition,
		}
	}

	if s.deviceAudioMode == AudioModeManual {
		s.state = newSessionState(s, kSessionStateIdle)
		s.state.ValidTransitions[kSessionStateIdle] = []SessionStateKind{kSessionStateConnecting, kSessionStateSpeaking, kSessionStateListening}
		s.state.ValidTransitions[kSessionStateConnecting] = []SessionStateKind{kSessionStateListening, kSessionStateIdle}
		s.state.ValidTransitions[kSessionStateListening] = []SessionStateKind{kSessionStateIdle}
		s.state.ValidTransitions[kSessionStateSpeaking] = []SessionStateKind{kSessionStateIdle}
	}
	return nil
}

func (s *Session) Close() {
	if s.cancel != nil && s.ctx.Err() != nil {
		s.cancel()
	}

	if s.conn != nil {
		s.conn.Close()
	}

	if s.asrProcessor != nil {
		s.asrProcessor.Close()
	}
}

func (s *Session) String() string {
	var sb strings.Builder

	sb.WriteString("Session{")
	sb.WriteString("deviceId: " + s.deviceId + ", ")
	sb.WriteString("clientId: " + s.clientId + ", ")
	sb.WriteString("sessionId: " + s.sessionId + ", ")
	sb.WriteString("protocolVersion: " + s.protocolVersion + ", ")
	sb.WriteString("}")

	return sb.String()
}
