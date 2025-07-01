package hub

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/repo"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/types"
	"github.com/huairu-tech-com/xiaozhi-gogo/utils"
	"github.com/rs/zerolog/log"

	"github.com/cloudwego/hertz/pkg/app"
)

const (
	DeviceIdHeader       = "Device-Id"       // Header key for Device ID
	ClientIdHeader       = "Client-Id"       // Header key for Client ID
	AcceptLanguageHeader = "Accept-Language" // Header key for Accept-Language
	UserAgentHeader      = "User-Agent"      // Header key for User-Agent
)

const (
	WebSocketProtocolVersion = 3
)

type MQTT struct {
	Endpoint     string `json:"endpoint"`
	ClientId     string `json:"client_id"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	PublishTopic string `json:"publish_topic"`
}

type Websocket struct {
	URL    string `json:"url"`
	Token  string `json:"token"`
	Verson int32  `json:"version"` // Version of the WebSocket protocol
}

type ServerTime struct {
	Timestamp      int64  `json:"timestamp"`       // Server time in milliseconds
	Timezone       string `json:"timezone"`        // Timezone information
	TimezoneOffset int32  `json:"timezone_offset"` // Offset in seconds from UTC
}

type Firmware struct {
	Version string `json:"version"` // Firmware version
	Url     string `json:"url"`     // URL to download the firmware
}

type OtaResponse struct {
	MQTT       *MQTT      `json:"mqtt,omitempty"` // MQTT configuration
	Websocket  Websocket  `json:"websocket"`      // WebSocket configuration
	ServerTime ServerTime `json:"server_time"`    // Server time information
	Firmware   Firmware   `json:"firmware"`       // Firmware information
}

func otaHandler(h *Hub) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var device types.Device
		if err := ctx.Bind(&device); err != nil {
			utils.BadRequest(ctx, "Invalid request body")
			return
		}

		device.DeviceId = ctx.Request.Header.Get("Device-Id")
		device.ClientId = ctx.Request.Header.Get("Client-Id")
		device.AcceptLanguage = ctx.Request.Header.Get("Accept-Language")
		device.UserAgent = ctx.Request.Header.Get("User-Agent")

		log.Debug().Str("device_id", device.DeviceId).Str("client_id", device.ClientId).Str("user_agent",
			device.UserAgent).Str("accept_language", device.AcceptLanguage).Msg("OTA Request")

		if err := ctx.Validate(&device); err != nil {
			utils.BadRequest(ctx, "Validation failed: "+err.Error())
			return
		}

		where := repo.WhereCondition{
			"device_id": device.DeviceId,
		}

		var (
			err error
			d   *types.Device
		)
		d, err = h.repo.FindDevice(where)
		if err != nil && !repo.IsNotExists(err) {
			utils.InternalServerError(ctx, "Failed to find device: "+err.Error())
			return
		}

		if d == nil {
			err = h.repo.CreateDevice(&device)
		} else {
			err = h.repo.UpdateDevice(d, &device)
		}

		if err != nil {
			utils.InternalServerError(ctx, "Failed to save device: "+err.Error())
			return
		}

		var response OtaResponse
		response.Websocket = Websocket{
			URL:    h.cfgOta.WsEndpoint,
			Token:  h.cfgOta.WsToken,
			Verson: WebSocketProtocolVersion,
		}
		response.ServerTime = ServerTime{
			Timestamp:      time.Now().Unix(),
			Timezone:       h.cfgOta.Timezone,
			TimezoneOffset: h.cfgOta.TimezoneOffset,
		}

		// TODO should dynamically calculate the firmware version and URL
		response.Firmware = Firmware{
			Version: h.cfgOta.FirmwareVersion,
			Url:     h.cfgOta.FirmwareUrl,
		}

		ctx.Header("Content-Type", "application/json")

		json.NewEncoder(os.Stdout).Encode(response)
		ctx.JSON(200, response)
	}
}
