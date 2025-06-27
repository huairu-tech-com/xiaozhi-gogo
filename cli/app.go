package cli

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/huairu-tech-com/xiaozhi-gogo/hub"
	"github.com/huairu-tech-com/xiaozhi-gogo/webui"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/rs/zerolog/log"
)

type runnable any
type shutdownable any

func runServers(ctx context.Context, cfg *Config) error {
	log.Info().Msgf("准备启动 WS 服务 %s", cfg.Addr)
	log.Info().Msgf("准备启动 WebUI 服务 %s", cfg.WebUIAddr)
	log.Info().Msg("准备启动 OTA 服务")

	hertzForDevice := server.Default(
		server.WithHostPorts(cfg.Addr),
	)
	deviceHubSrv := hub.New()
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
		log.Error().Err(ctx.Err()).Msg("Context取消，准备停止服务")
		break
	case err := <-errCh:
		log.Error().Err(err).Msgf("服务启动失败%+v", err)
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
					log.Error().Err(err).Msg("服务停止失败")
				}
			}
		}(srv)
	}

	wg.Wait()

	return nil
}
