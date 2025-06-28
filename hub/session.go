package hub

import (
	"context"

	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/repo"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/types"

	"github.com/go-playground/validator/v10"
	"github.com/hertz-contrib/websocket"
)

var sessionContextKey = struct{}{} // Unique key for storing SessionContext in context.Context

type sessionContext struct {
	Hub           *Hub
	ParentContext context.Context
	DeviceId      string `validate:"required"`
	SessionId     string `validate:"required"`
	ClientId      string `validate:"required"`
	Device        *types.Device
}

func (sc *sessionContext) IsValid() bool {
	return validator.New().Struct(sc) == nil
}

func FromContext(ctx context.Context) *sessionContext {
	if ctx == nil {
		panic("context is nil")
	}

	sc, ok := ctx.Value(sessionContextKey).(*sessionContext)
	if !ok || sc == nil {
		panic("session context not found in context")
	}

	return sc
}

// session represents a session for a device in the hub.
type Session struct {
	conn   *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc
}

func NewSession(ctx context.Context, conn *websocket.Conn) *Session {
	s := &Session{
		conn: conn,
	}

	c := FromContext(ctx)
	s.ctx, s.cancel = context.WithCancel(c.ParentContext)

	return s
}

func (s *Session) populate() error {
	var err error

	c := FromContext(s.ctx)
	c.Device, err = c.Hub.repo.FindDevice(repo.WhereCondition{})
	return err
}

func (s *Session) loop() error {
	c := FromContext(s.ctx)

	for {
		if c.ParentContext.Err() != nil {
			return c.ParentContext.Err()
		}

		mt, message, err := s.conn.ReadMessage()
		if err != nil {
			break
		}
		err = s.conn.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}

	return nil
}
