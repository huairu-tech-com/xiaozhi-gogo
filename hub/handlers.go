package hub

import (
	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/hertz-contrib/websocket"

	"github.com/pkg/errors"
)

var (
	ErrSessionIdMismatch = errors.New("session ID mismatch")
)

func (s *Session) handleHello(raw []byte) error {
	msg, err := MessageFromBytes[Hello](raw)
	if err != nil {
		return err
	}

	s.deviceVersion = msg.Version
	s.deviceAudioParams.SampleRate = msg.AudioParams.SampleRate
	s.deviceAudioParams.Format = msg.AudioParams.Format
	s.deviceAudioParams.Channels = msg.AudioParams.Channels
	s.deviceAudioParams.FrameDuration = msg.AudioParams.FrameDuration
	s.deviceSupportMCP = msg.Features.MCP

	// generate a new session ID
	s.sessionId = uuid.New().String()

	resp := HelloResponse{
		Type:        MessageTypeHello,
		SessionId:   s.sessionId,
		Transport:   TransportTypeWebsocket,
		AudioParams: msg.AudioParams,
	}

	w, err := s.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer w.Close()

	return sonic.ConfigDefault.NewEncoder(w).Encode(resp)
}

func (s *Session) handleListenStart(raw []byte) error {
	msg, err := MessageFromBytes[ListenStart](raw)
	if err != nil {
		return err
	}

	if !s.isSessionIdMatch(msg.SessionId) {
		return ErrSessionIdMismatch
	}
	s.deviceAudioMode = msg.Mode
	if err := s.buildState(); err != nil {
		return err
	}

	return nil
}

func (s *Session) handleListenStop(raw []byte) error {
	return nil
}

func (s *Session) handleListenDetect(raw []byte) error {
	msg, err := MessageFromBytes[ListenDetect](raw)
	if err != nil {
		return err
	}

	if s.isSessionIdMatch(msg.SessionId) {
		return ErrSessionIdMismatch
	}

	// TODO

	return nil
}

func (s *Session) handleTTSStart(raw []byte) error {
	return nil
}

func (s *Session) handleTTSStop(raw []byte) error {
	return nil
}

func (s *Session) handleTTSSentenceStart(raw []byte) error {
	return nil
}

func (s *Session) handleAudio(opusData []byte) error {
	return nil
}

func (s *Session) handleAbort(raw []byte) error {
	msg, err := MessageFromBytes[Abort](raw)
	if err != nil {
		return err
	}

	if !s.isSessionIdMatch(msg.SessionId) {
		return ErrSessionIdMismatch
	}

	return nil
}

func (s *Session) handleIotDescribe(raw []byte) error {
	return nil
}

func (s *Session) handleIotStates(raw []byte) error {
	return nil
}

func (s *Session) handleLlm(raw []byte) error {
	return nil
}
