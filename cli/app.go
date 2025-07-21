package cli

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/huairu-tech-com/xiaozhi-gogo/config"
	"github.com/huairu-tech-com/xiaozhi-gogo/src"
	"github.com/huairu-tech-com/xiaozhi-gogo/webui"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/rs/zerolog/log"
)

type runnable any
type shutdownable any

func runServers(ctx context.Context, cfg *config.Config) error {
	log.Info().Msgf("launching WS service: %s", cfg.Addr)
	log.Info().Msgf("launching WebUI %s", cfg.WebUIAddr)
	log.Info().Msg("launching OTA service")

	hertzForDevice := server.Default(
		server.WithHostPorts(cfg.Addr),
	)
	deviceHubSrv, err := src.New(cfg.Ota, cfg.Asr, cfg.Llm)
	if err != nil {
		return err
	}

	deviceHubSrv.Hook(hertzForDevice)

	hertzForInternal := server.Default(
		server.WithHostPorts(cfg.WebUIAddr),
	)
	webUISrv := webui.New()
	webUISrv.Hook(hertzForInternal)

	errCh := make(chan error, 1)
	go func() {
		errCh <- deviceHubSrv.Run(ctx)
	}()

	go func() {
		errCh <- webUISrv.Run(ctx)
	}()

	go func() {
		hertzForDevice.SetCustomSignalWaiter(func(e chan error) error {
			<-ctx.Done() // if someone send SIGINT or SIGTERM, this will be pass
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		})

		errCh <- hertzForDevice.Run()
	}()

	go func() {
		hertzForInternal.SetCustomSignalWaiter(func(e chan error) error {
			<-ctx.Done() // if someone send SIGINT or SIGTERM, this will be pass
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		})

		errCh <- hertzForInternal.Run()
	}()

	select {
	case <-ctx.Done():
		log.Error().Err(ctx.Err()).Msg("context cancelled, shutting down services")
		break
	case err := <-errCh:
		log.Error().Err(err).Msgf("%+v", err)
		break
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var wg sync.WaitGroup
	shutdownables := []shutdownable{
		deviceHubSrv,
		webUISrv,
		hertzForInternal,
		hertzForDevice,
	}

	wg.Add(len(shutdownables))
	for _, srv := range shutdownables {
		go func(srv shutdownable) {
			defer wg.Done()
			if s, ok := srv.(interface{ Shutdown(context.Context) error }); ok {
				if err := s.Shutdown(shutdownCtx); err != nil {
					log.Error().Err(err).Msg("shutdown service failed")
				}
			}
		}(srv)
	}

	wg.Wait()

	return nil
}
