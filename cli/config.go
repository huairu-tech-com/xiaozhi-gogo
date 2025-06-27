package cli

type LogConfig struct {
	Level   string `yaml:"level"`
	LogPath string `yaml:"log_path"`
}

type Config struct {
	// endpoint of both WS and HTTP
	Addr string `yaml:"addr"`
	// Web UI addr
	WebUIAddr string `yaml:"web_ui_addr"`

	Log LogConfig `yaml:"log"`

	EnableProfile bool `yaml:"enable_profile"`
}

func DefaultConfig() *Config {
	return &Config{
		Addr:      "0.0.0.0:3456",
		WebUIAddr: "0.0.0.0:3457",
		Log: LogConfig{
			Level:   "info",
			LogPath: "logs/app.log",
		},
		EnableProfile: false,
	}
}
