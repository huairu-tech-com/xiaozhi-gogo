package hub

import (
	"context"

	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/repo"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/types"

	"github.com/go-playground/validator/v10"
	"github.com/hertz-contrib/websocket"
	"github.com/pkg/errors"
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

	msgHandlers map[MessageType]ClientMessageHandler
	ctx         context.Context
	cancel      context.CancelFunc
}

func newSession(ctx context.Context) *Session {
	s := &Session{
		msgHandlers:     make(map[MessageType]ClientMessageHandler),
		deviceAudioMode: AudioModeNone,
	}

	s.buildState()

	s.msgHandlers[MessageTypeRawAudio] = s.handleAudio
	s.msgHandlers[MessageTypeHello] = s.handleHello
	s.msgHandlers[MessageTypeListenStart] = s.handleListenStart
	s.msgHandlers[MessageTypeListenStop] = s.handleListenStop
	s.msgHandlers[MessageTypeListenDetect] = s.handleListenDetect
	s.msgHandlers[MessageTypeTTSStart] = s.handleTTSStart
	s.msgHandlers[MessageTypeTTSStop] = s.handleTTSStop
	s.msgHandlers[MessageTypeTTSSentenceStart] = s.handleTTSSentenceStart

	s.ctx, s.cancel = context.WithCancel(ctx)

	return s
}

func (s *Session) isValid() bool {
	return validator.New().Struct(s) == nil
}

// load device object from repository
func (s *Session) populateDevice() error {
	var err error
	s.device, err = s.hub.repo.FindDevice(repo.WhereCondition{})
	return err
}

func (s *Session) loop() error {
	var (
		err      error
		mt       int
		rawBytes []byte
	)

	for {
		if s.ctx.Err() != nil {
			return s.ctx.Err()
		}

		mt, rawBytes, err = s.conn.ReadMessage()
		if err != nil {
			log.Error().Err(err).Msgf("Failed to read message from device %s: %v", s.deviceId, err)
			s.cancel()
			continue
		}

		var messagePayloadType MessageType = MessageTypeNone
		if mt == websocket.TextMessage {
			var meta *MetaMessage
			meta, err = MessageFromBytes[MetaMessage](rawBytes)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to parse message from device %s: %v, content is %s", s.deviceId, err, string(rawBytes))
				s.cancel()
				continue
			}
			messagePayloadType = meta.MessageType()
		}

		if mt == websocket.BinaryMessage {
			messagePayloadType = MessageTypeRawAudio
		}

		handler, ok := s.msgHandlers[messagePayloadType]
		if !ok {
			log.Error().Msgf("No handler found for message type %s from device %s", messagePayloadType, s.deviceId)
			s.cancel()
			continue
		}

		if err = handler(rawBytes); err != nil {
			log.Error().Err(err).Msgf("Failed to handle message type %s from device %s: %v", messagePayloadType, s.deviceId, err)
			s.cancel()
			continue
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

func (s *Session) buildState() error {
	if s.deviceAudioMode != AudioModeNone {
		return errors.New("session state can not be built when audio mode is set")
	}

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

		s.state.Callbacks[KindPair{kSessionStateIdle, kSessionStateConnecting}] = []TransitionCallback{
			logTransition,
		}

		s.state.Callbacks[KindPair{kSessionStateConnecting, kSessionStateListening}] = []TransitionCallback{
			logTransition,
		}
	}

	if s.deviceAudioMode == AudioModeManual {
		s.state = newSessionState(s, kSessionStateIdle)
		s.state.ValidTransitions[kSessionStateIdle] = []SessionStateKind{kSessionStateConnecting, kSessionStateSpeaking, kSessionStateListening}
		s.state.ValidTransitions[kSessionStateConnecting] = []SessionStateKind{kSessionStateListening, kSessionStateIdle}
		s.state.ValidTransitions[kSessionStateListening] = []SessionStateKind{kSessionStateIdle}
		s.state.ValidTransitions[kSessionStateSpeaking] = []SessionStateKind{kSessionStateIdle}

		s.state.Callbacks[KindPair{kSessionStateIdle, kSessionStateConnecting}] = []TransitionCallback{
			logTransition,
		}
	}
	return nil
}

func (s *Session) close() {
	if s.cancel != nil && s.ctx.Err() != nil {
		s.cancel()
	}
	if s.conn != nil {
		s.conn.Close()
	}
}
