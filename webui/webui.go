package webui

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app/server"

	"github.com/huairu-tech-com/xiaozhi-gogo/utils"
)

type WebUI struct {
}

func New() *WebUI {
	return &WebUI{}
}

func (w *WebUI) Run(ctx context.Context) error {
	return nil
}

func (w *WebUI) Shutdown(ctx context.Context) error {
	return nil
}

func (w *WebUI) Hook(h *server.Hertz) {
	h.GET("/health", utils.HealthCheck())
	configGroup := h.Group("/webui/config")

	handleConfig(w, configGroup)

	apiGroup := h.Group("/webui/api")
	handleInternalAPI(w, apiGroup)
}
