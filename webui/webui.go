package webui

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
)

type WebUI struct {
}

func New() *WebUI {
	return &WebUI{}
}

func (w *WebUI) Run(ctx context.Context) error {
	time.Sleep(1 * time.Second * 100) // 模拟一些清理工作
	return nil
}

func (w *WebUI) Shutdown(ctx context.Context) error {
	return nil
}

func (w *WebUI) Hook(h *server.Hertz) {
}
