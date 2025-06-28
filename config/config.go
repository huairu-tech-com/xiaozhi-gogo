package config

import (
	"bytes"

	"github.com/go-yaml/yaml"
)

type LogConfig struct {
	Level   string `yaml:"level"`
	LogPath string `yaml:"log_path"`
}

type OtaConfig struct {
	WsEndpoint      string `yaml:"ws_endpoint"`      // WebSocket endpoint for OTA ws:// or wss:// , fully qualified URL
	WsToken         string `yaml:"ws_token"`         // WebSocket token for authentication, optional
	FirmwareUrl     string `yaml:"firmware_url"`     // URL to download the firmware, fully qualified URL
	FirmwareVersion string `yaml:"firmware_version"` // Firmware version, e.g., "1.0.0", should be latest version
	Timezone        string `yaml:"timezone"`         // Timezone, e.g., "Asia/Shanghai"
	TimezoneOffset  int32  `yaml:"timezone_offset"`  // Timezone offset in seconds, e.g., 28800 for Asia/Shanghai
}

type Config struct {
	Addr          string     `yaml:"addr"`        // endpoint of both WS and HTTP, publicly accessible
	WebUIAddr     string     `yaml:"web_ui_addr"` // web UI address
	Log           *LogConfig `yaml:"log"`         // log
	Ota           *OtaConfig `yaml:"ota"`         // OTA configuration
	EnableProfile bool       `yaml:"enable_profile"`
}

func DefaultConfig() *Config {
	return &Config{
		Addr:      "0.0.0.0:3456",
		WebUIAddr: "localhost:3457",
		Log: &LogConfig{
			Level:   "info",
			LogPath: "logs/app.log",
		},
		Ota: &OtaConfig{
			WsEndpoint:      "ws://localhost:3456/xiaozhi/ota/",
			WsToken:         "xiaozhi-gogo",
			FirmwareUrl:     "http://localhost:3456/firmware/latest",
			FirmwareVersion: "1.0.0",
			Timezone:        "Asia/Shanghai",
			TimezoneOffset:  28800, // Asia/Shanghai is UTC+8
		},
		EnableProfile: false,
	}
}

func (cfg *Config) String() string {
	var buf bytes.Buffer

	yaml.NewEncoder(&buf).Encode(cfg)
	return buf.String()
}
