package api

import (
	"context"
	"fmt"
	"github.com/kubemq-io/file-uploader/config"
	"github.com/kubemq-io/file-uploader/pkg/logger"
	"github.com/kubemq-io/file-uploader/source"
	"github.com/kubemq-io/infra-services-usage/pkg/api"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"time"
)

type Server struct {
	echoWebServer *echo.Echo
	cfg           *config.Config
	logger        *logger.Logger
	sourceService *source.Service
}

func Start(ctx context.Context, cfg *config.Config, sourceService *source.Service) (*Server, error) {
	s := &Server{
		cfg:           cfg,
		echoWebServer: echo.New(),
		logger:        logger.NewLogger("file-uploader-api"),
		sourceService: sourceService,
	}
	s.echoWebServer.Use(middleware.Recover())
	s.echoWebServer.Use(s.loggingMiddleware())
	s.echoWebServer.Use(middleware.CORS())
	s.echoWebServer.HideBanner = true

	s.echoWebServer.GET("/health", func(c echo.Context) error {
		return c.String(200, "ok")
	})
	s.echoWebServer.GET("/ready", func(c echo.Context) error {
		return c.String(200, "ready")
	})
	s.echoWebServer.GET("/status", func(c echo.Context) error {
		return c.JSONPretty(200, s.sourceService.Status(), "\t")
	})
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.echoWebServer.Start(fmt.Sprintf("0.0.0.0:%d", cfg.ApiPort))
	}()
	select {
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
		return s, nil
	case <-time.After(1 * time.Second):
		return s, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("error strarting api server, %w", ctx.Err())
	}
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.echoWebServer.Shutdown(ctx)
}

func (s *Server) loggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			s.logger.Debugw("start api call", "method", c.Request().RequestURI, "sender", c.Request().RemoteAddr)
			err = next(c)
			val := c.Get("result")
			res, ok := val.(*api.Response)
			if ok {
				if !res.Ok {
					s.logger.Errorw("api call ended with error", "method", c.Request().RequestURI, "sender", c.Request().RemoteAddr, "error", res.Error)
				} else {
					s.logger.Debugw("api call ended successfully", "method", c.Request().RequestURI, "sender", c.Request().RemoteAddr, "result", res.Data)
				}

			}
			return err
		}
	}
}
