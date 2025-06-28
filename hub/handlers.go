package hub

import (
	"github.com/bytedance/sonic"
	"github.com/hertz-contrib/websocket"
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

	resp := HelloResponse{
		Type:        MessageTypeHello,
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
	return nil
}

func (s *Session) handleListenStop(raw []byte) error {
	return nil
}

func (s *Session) handleListenDetect(raw []byte) error {
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

func (s *Session) handleAbort(raw []byte) error {
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
