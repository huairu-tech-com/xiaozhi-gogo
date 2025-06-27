package utils

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

func HealthCheck() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}
}

func BadRequest(ctx *app.RequestContext, message string) {
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusBadRequest, map[string]string{"error": message})
}

func InternalServerError(ctx *app.RequestContext, message string) {
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusInternalServerError, map[string]string{"error": message})

}
