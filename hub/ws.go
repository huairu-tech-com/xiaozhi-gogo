package hub

import (
	"context"
	"time"

	"github.com/huairu-tech-com/xiaozhi-gogo/utils"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"
	"github.com/rs/zerolog/log"
)

func wsHandler(h *Hub) app.HandlerFunc {
	var upgrader = websocket.HertzUpgrader{
		HandshakeTimeout: time.Second * 10, // Set a timeout for the handshake
	}

	return func(c context.Context, ctx *app.RequestContext) {
		if err := upgrader.Upgrade(ctx, wsProtocolHandler(c, ctx, h)); err != nil {
			utils.InternalServerError(ctx, "failed to upgrade connection: "+err.Error())
		}
	}
}

func wsProtocolHandler(ctx context.Context, rctx *app.RequestContext, h *Hub) websocket.HertzHandler {
	return func(conn *websocket.Conn) {
		defer func() {
			h.sessionMap.Del(rctx.Request.Header.Get("Device-Id"))
			conn.Close()
		}()

		sc := &sessionContext{
			ParentContext: ctx,
			DeviceId:      rctx.Request.Header.Get("Device-Id"),
			ClientId:      rctx.Request.Header.Get("Client-Id"),
			SessionId:     rctx.Request.Header.Get("Session-Id"),
		}

		if sc.IsValid() {
		}

		valueCtx := context.WithValue(ctx, sessionContextKey, sc)
		s := NewSession(valueCtx, conn)
		if err := s.populate(); err != nil {
			log.Error().Err(err).Msgf("Failed to populate session context err: %+v", err)
			goto end
		}
		h.sessionMap.Set(rctx.Request.Header.Get("Device-Id"), s)

		if err := s.loop(); err != nil {
			log.Error().Err(err).Msgf("Session loop error: %+v", err)
		}

	end:
	}
}
