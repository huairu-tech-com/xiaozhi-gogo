package webui

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/rs/zerolog/log"
)

func handleConfig(w *WebUI, group *route.RouterGroup) {
	group.GET("/get", getConfig(w))
	group.POST("/validate", validateConfig(w))
	group.POST("/override", overrideConfig(w))
}

func getConfig(w *WebUI) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		log.Info().Msg("Fetching configuration")
	}
}

func validateConfig(w *WebUI) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		log.Info().Msg("Validating configuration")
	}
}

func overrideConfig(w *WebUI) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		log.Info().Msg("Overriding configuration")
	}
}
