package src

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/huairu-tech-com/xiaozhi-gogo/config"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/repo"
	"github.com/huairu-tech-com/xiaozhi-gogo/utils"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cornelk/hashmap"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Hub struct {
	cfgOta *config.OtaConfig
	cfgAsr *config.AsrConfig // ASR configuration
	cfgLlm *config.LlmConfig // LLM configuration, if needed
	cfgTts *config.TtsConfig // TTS configuration, if needed

	repo       repo.Respository
	sessionMap *hashmap.Map[string, *Session]
}

func New(cfgOta *config.OtaConfig,
	cfgAsr *config.AsrConfig,
	cfgLlm *config.LlmConfig,
	cfgTts *config.TtsConfig,
) (*Hub, error) {
	h := &Hub{
		cfgOta:     cfgOta,
		cfgAsr:     cfgAsr,
		cfgLlm:     cfgLlm,
		cfgTts:     cfgTts,
		repo:       repo.NewInMemoryRepository(),
		sessionMap: hashmap.New[string, *Session](),
	}

	if cfgOta == nil {
		return nil, errors.New("ota configuration cannot be nil")
	}

	if cfgAsr == nil {
		return nil, errors.New("asr configuration cannot be nil")
	}

	if cfgAsr.Doubao == nil {
		return nil, errors.New("doubao ASR configuration cannot be nil")
	}

	return h, nil
}

func (h *Hub) Run(ctx context.Context) error {
	time.Sleep(100000 * time.Second) // Simulate long-running process
	// 启动 Hub 的逻辑
	return nil
}

func (h *Hub) Shutdown(ctx context.Context) error {
	// 停止 Hub 的逻辑
	return nil
}

func (h *Hub) Hook(srv *server.Hertz) {
	srv.Use(LoggerMiddleware())

	srv.GET("/health", utils.HealthCheck())
	srv.POST("/xiaozhi/ota/", otaHandler(h))

	// https: //github.com/cloudwego/hertz/issues/121
	srv.GET("/xiaozhi/ws/", wsHandler(h))

}

func LoggerMiddleware() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		start := time.Now()

		defer func() {
			stop := time.Now()
			log.Info().
				Str("remote_ip", ctx.ClientIP()).
				Str("method", string(ctx.Method())).
				Str("path", string(ctx.Path())).
				Str("user_agent", string(ctx.UserAgent())).
				Int("status", ctx.Response.StatusCode()).
				Dur("latency", stop.Sub(start)).
				Str("latency_human", stop.Sub(start).String()).
				Msg("request processed")
		}()

		ctx.Next(c)
	}
}
