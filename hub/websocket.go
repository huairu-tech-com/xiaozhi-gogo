package hub

import (
	"context"
	"time"

	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/asr/doubao"
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
		s := newSession(ctx)

		defer func() {
			h.sessionMap.Del(rctx.Request.Header.Get("Device-Id"))
			if s.asrSrv != nil {
				s.asrSrv.Close()
			}
			conn.Close()
			s.close()
		}()

		s.hub = h
		s.conn = conn
		s.deviceId = rctx.Request.Header.Get("Device-Id")
		s.clientId = rctx.Request.Header.Get("Client-Id")
		s.protocolVersion = rctx.Request.Header.Get("Protocol-Version")
		authorization := rctx.Request.Header.Get("Authorization")
		if authorization != "" && len(authorization) > 7 {
			s.bearerToken = authorization[7:]
		}

		log.Info().Msgf("New session created: DeviceId=%s, ClientId=%s, SessionId=%s",
			s.deviceId, s.clientId, s.sessionId)

		if !s.isValid() {
			log.Error().Msgf("Invalid session parameters: DeviceId=%s, ClientId=%s, SessionId=%s",
				s.deviceId, s.clientId, s.sessionId)
			return
		}

		if !s.isAuthenticated() {
			log.Error().Msgf("Session not authenticated: DeviceId=%s, ClientId=%s, SessionId=%s",
				s.deviceId, s.clientId, s.sessionId)
			return
		}

		if !s.isAuthorized() {
			log.Error().Msgf("Session not authorized: DeviceId=%s, ClientId=%s, SessionId=%s",
				s.deviceId, s.clientId, s.sessionId)
			return
		}

		asrConfig := doubao.DefaultConfig()
		asrConfig.ApiKey = h.cfgAsr.Doubao.ApiKey
		asrConfig.AccessKey = h.cfgAsr.Doubao.AccessKey
		var err error
		s.asrSrv, err = doubao.DefaultDialer(ctx, asrConfig)
		if err != nil {
			return
		}

		// if err := s.populateDevice(); err != nil {
		// 	log.Error().Err(err).Msgf("Failed to populate session context err: %+v", err)
		// 	return
		// }
		h.sessionMap.Set(rctx.Request.Header.Get("Device-Id"), s)

		if err := s.loop(); err != nil {
			log.Error().Err(err).Msgf("Session loop error: %+v", err)
		}
	}
}
