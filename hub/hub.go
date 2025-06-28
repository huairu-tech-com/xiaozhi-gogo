package hub

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/huairu-tech-com/xiaozhi-gogo/config"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/repo"
	"github.com/huairu-tech-com/xiaozhi-gogo/utils"

	"github.com/cornelk/hashmap"
)

type Hub struct {
	cfgOta     *config.OtaConfig
	repo       repo.Respository
	sessionMap *hashmap.Map[string, *Session]
}

func New(cfgOta *config.OtaConfig) *Hub {
	return &Hub{
		cfgOta:     cfgOta,
		repo:       repo.NewInMemoryRepository(),
		sessionMap: hashmap.New[string, *Session](),
	}
}

func (h *Hub) Run(ctx context.Context) error {
	// 启动 Hub 的逻辑
	return nil
}

func (h *Hub) Shutdown(ctx context.Context) error {
	// 停止 Hub 的逻辑
	return nil
}

func (h *Hub) Hook(srv *server.Hertz) {
	srv.GET("/health", utils.HealthCheck())
	srv.POST("/xiaozhi/ota/", otaHandler(h))

	// https: //github.com/cloudwego/hertz/issues/121
	srv.POST("/xiaozhi/ws/", wsHandler(h))

}
