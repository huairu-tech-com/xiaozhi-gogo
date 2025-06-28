package hub

import (
	"context"

	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/repo"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/types"
	"github.com/rs/zerolog/log"

	"github.com/go-playground/validator/v10"
	"github.com/hertz-contrib/websocket"
)

type ClientMessageHandler func([]byte) error

type Session struct {
	conn *websocket.Conn
	hub  *Hub

	deviceId  string `validate:"required"`
	sessionId string `validate:"required"`
	clientId  string `validate:"required"`

	device *types.Device

	deviceVersion     int32
	deviceAudioParams HelloAudioParams

	msgHandlers map[MessageType]ClientMessageHandler
	ctx         context.Context
	cancel      context.CancelFunc
}

func newSession(ctx context.Context) *Session {
	s := &Session{
		msgHandlers: make(map[MessageType]ClientMessageHandler),
	}

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

		if mt != websocket.TextMessage {
			continue // Only handle text messages
		}

		var meta *MetaMessage
		meta, err = MessageFromBytes[MetaMessage](rawBytes)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to parse message from device %s: %v, content is %s", s.deviceId, err, string(rawBytes))
			s.cancel()
			continue
		}

		handler, ok := s.msgHandlers[meta.MessageType()]
		if !ok {
			log.Error().Msgf("No handler found for message type %s from device %s", meta.Type, s.deviceId)
			s.cancel()
			continue
		}

		if err = handler(rawBytes); err != nil {
			log.Error().Err(err).Msgf("Failed to handle message type %s from device %s: %v", meta.Type, s.deviceId, err)
			s.cancel()
			continue
		}
	}
}

func (s *Session) close() {
	if s.cancel != nil && s.ctx.Err() != nil {
		s.cancel()
	}
	if s.conn != nil {
		s.conn.Close()
	}
}
