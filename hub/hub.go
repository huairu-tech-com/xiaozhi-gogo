package hub

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/repo"
	"github.com/huairu-tech-com/xiaozhi-gogo/utils"
)

type Hub struct {
	websocketUrl    string
	websocketToken  string
	timezone        string
	timezoneOffset  int32
	firmwareVersion string
	firmwareUrl     string

	repo repo.Respository
}

func New() *Hub {
	return &Hub{}
}

func (h *Hub) Run(ctx context.Context) error {
	// 启动 Hub 的逻辑
	time.Sleep(1 * time.Second * 100) // 模拟一些清理工作
	return nil
}

func (h *Hub) Shutdown(ctx context.Context) error {
	// 停止 Hub 的逻辑
	return nil
}

func (h *Hub) Hook(srv *server.Hertz) {
	srv.GET("/health", utils.HealthCheck())
	srv.POST("/xiaozhi/ota/", otaHandler(h))
}
