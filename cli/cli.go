package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/huairu-tech-com/xiaozhi-gogo/config"

	"github.com/go-yaml/yaml"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	ExitCodeOK   = 0
	ExitCodeFail = 1
)

var (
	configPath = flag.String("config-path", "config.yaml", "where config file is located, default is config.yaml")
	dump       = flag.Bool("dump", false, "output default config and exit, useful for generating a new config file")
)

func Run() int {
	flag.Parse()

	if *dump {
		enc := yaml.NewEncoder(os.Stdout)
		defer enc.Close()
		enc.Encode(config.DefaultConfig())

		return ExitCodeOK
	}

	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "config path %s does not exists", *configPath)
		return ExitCodeFail
	}

	configFile, err := os.Open(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open config file failed with errro %v", err)
	}

	cfg := config.DefaultConfig()
	if err := yaml.NewDecoder(configFile).Decode(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "decode config file failed with error %v", err)
		return ExitCodeFail
	}

	fmt.Println(strings.Repeat("=", 50))
	fmt.Fprintf(os.Stdout, "config file loaded %s\n", *configPath)
	fmt.Fprintf(os.Stdout, "%+v\n", cfg)
	fmt.Println(strings.Repeat("=", 50))

	fmt.Fprintf(os.Stderr, "log setup failed %v", err)
	if err := setupLogger(cfg.Log); err != nil {
		return ExitCodeFail
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt,
		syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	if err := runServers(ctx, cfg); err != nil {
		log.Error().Err(err).Msg("run server failed")
		return ExitCodeFail
	}

	return ExitCodeOK
}

func setupLogger(logCfg *config.LogConfig) error {
	loglevel := zerolog.InfoLevel
	switch strings.ToLower(logCfg.Level) {
	case "debug":
		loglevel = zerolog.DebugLevel
		break
	case "info":
		loglevel = zerolog.InfoLevel
		break
	case "warn", "warning":
		loglevel = zerolog.WarnLevel
		break
	case "error":
		loglevel = zerolog.ErrorLevel
		break
	case "fatal":
		loglevel = zerolog.FatalLevel
		break
	default:
		loglevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(loglevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logFile, err := os.OpenFile(logCfg.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file %v", err)
		return err
	}
	log.Logger = zerolog.New(io.MultiWriter(os.Stdout, logFile)).With().Logger()
	log.Info().Msgf("log level set to %s", logCfg.Level)
	log.Info().Msgf("log path set to %s", logCfg.LogPath)

	return nil
}
