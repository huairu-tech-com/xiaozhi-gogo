package hub

import (
	"encoding/binary"
	"time"

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
	if err := s.buildState(msg.Mode); err != nil {
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

func (s *Session) handleAudio(opusData []byte) error {
	var bp3 BinaryProtocol3
	if len(opusData) < (1 + 1 + 2) {
		return errors.New("opus data too short")
	}

	bp3.Type = uint8(opusData[0])
	bp3.Reserved = uint8(opusData[1])
	bp3.PayloadSize = uint16(binary.BigEndian.Uint16(opusData[2:4]))

	if len(opusData) < int(bp3.PayloadSize+4) {
		return errors.Errorf("opus data too short, expected %d bytes, got %d bytes", bp3.PayloadSize+4, len(opusData))
	}
	bp3.Payload = opusData[4 : 4+bp3.PayloadSize]

	s.audioProcessor.PushOpus(bp3.Payload)
	for !s.audioProcessor.IsEmpty() {
		audioFrame, seqNo, isLast, err := s.audioProcessor.PopPCMWithVoice()
		if err != nil {
			return err
		}

		if err := s.asrSrv.SendAudio(audioFrame, seqNo, isLast, 50*time.Millisecond); err != nil {
			return err
		}
	}

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
