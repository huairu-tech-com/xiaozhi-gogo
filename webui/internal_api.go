package webui

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
)

func handleInternalAPI(w *WebUI, group *route.RouterGroup) {
	// Define your API endpoints here
	group.GET("/devices", func(ctx context.Context, c *app.RequestContext) {
		// Handle endpoint1
	})
}
