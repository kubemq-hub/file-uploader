package main

import (
	"context"
	"fmt"
	"github.com/kubemq-io/file-uploader/config"
	"github.com/kubemq-io/file-uploader/pkg/logger"
	"github.com/kubemq-io/file-uploader/server"
	"os"
	"os/signal"
	"syscall"
)

var log = logger.NewLogger("file-uploader")

func run() error {
	var gracefulShutdown = make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGTERM)
	signal.Notify(gracefulShutdown, syscall.SIGINT)
	signal.Notify(gracefulShutdown, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	configCh := make(chan *config.Config)
	cfg, err := config.Load(configCh)
	if err != nil {
		return err
	}
	apiServer, err := server.Start(ctx, cfg)
	if err != nil {
		return err
	}
	for {
		select {
		case newConfig := <-configCh:
			if apiServer != nil {
				err = apiServer.Stop()
				if err != nil {
					return fmt.Errorf("error on shutdown api server: %s", err.Error())
				}
			}
			apiServer, err = server.Start(ctx, newConfig)
			if err != nil {
				return fmt.Errorf("error on start api server: %s", err.Error())
			}
		case <-gracefulShutdown:
			_ = apiServer.Stop()
			return nil
		}
	}
}

func main() {
	log.Info("starting file-uploader")
	if err := run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
