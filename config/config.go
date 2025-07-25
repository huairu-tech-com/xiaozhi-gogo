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

type DoubalAsrConfig struct {
	ApiKey    string `yaml:"api_key"`    // API key for Doubao ASR
	AccessKey string `yaml:"access_key"` // Access key for Doubao ASR
}

type AsrConfig struct {
	Doubao *DoubalAsrConfig `yaml:"doubao"` // Doubao ASR configuration
}

type DeepseekConfig struct {
	Model   string `yaml:"model"`    // DeepSeek model name, e.g., "deepseek-chat-3.5"
	BaseUrl string `yaml:"base_url"` // Base URL for DeepSeek API, e.g., "https://api.deepseek.com/v1/chat/completions"
	ApiKey  string `yaml:"api_key"`  // API key for DeepSeek, optional
}

type LlmConfig struct {
	Deepseek *DeepseekConfig `yaml:"deepseek"` // DeepSeek LLM configuration
}

type CosyVoiceConfig struct {
	BaseUrl string `yaml:"base_url"` // Base URL for CosyVoice TTS API, e.g., "https://api.cosyvoice.com/v1/tts"
	Voice   string `yaml:"voice"`    // Voice ID for CosyVoice TTS, e.g., "cosy-voice-1"
	ApiKey  string `yaml:"api_key"`  // API key for CosyVoice TTS
}

type TtsConfig struct {
	CosyVoice *CosyVoiceConfig `yaml:"cosy_voice"` // CosyVoice TTS configuration
}

type Config struct {
	Addr          string     `yaml:"addr"`        // endpoint of both WS and HTTP, publicly accessible
	WebUIAddr     string     `yaml:"web_ui_addr"` // web UI address
	Log           *LogConfig `yaml:"log"`         // log
	Ota           *OtaConfig `yaml:"ota"`         // OTA configuration
	Asr           *AsrConfig `yaml:"asr"`         // ASR configuration
	Llm           *LlmConfig `yaml:"llm"`         // LLM configuration, if needed
	Tts           *TtsConfig `yaml:"tts"`         // TTS configuration, if needed
	EnableProfile bool       `yaml:"enable_profile"`
}

func DefaultConfig() *Config {
	return &Config{
		Addr:      "0.0.0.0:3457",
		WebUIAddr: "localhost:3456",
		Log: &LogConfig{
			Level:   "info",
			LogPath: "logs/app.log",
		},
		Asr: &AsrConfig{
			Doubao: &DoubalAsrConfig{},
		},
		Llm: &LlmConfig{
			Deepseek: &DeepseekConfig{
				BaseUrl: "https://api.deepseek.com/v1/chat/completions",
				Model:   "deepseek-chat-3.5",
				ApiKey:  "",
			},
		},
		Tts: &TtsConfig{
			CosyVoice: &CosyVoiceConfig{
				BaseUrl: "https://api.siliconflow.cn",
				Voice:   "benjimin",
				ApiKey:  "",
			},
		},
		Ota: &OtaConfig{
			WsEndpoint:      "ws://192.168.1.7:3457/xiaozhi/ws/",
			WsToken:         "xiaozhi-gogo",
			FirmwareUrl:     "http://192.168.1.7:3457/firmware/latest",
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
