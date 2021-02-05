package main

import (
	"context"
	"fmt"
	"github.com/kubemq-io/file-uploader/api"
	"github.com/kubemq-io/file-uploader/config"
	"github.com/kubemq-io/file-uploader/pkg/file_creator"
	"github.com/kubemq-io/file-uploader/pkg/logger"
	"github.com/kubemq-io/file-uploader/senders/kubemq"
	"github.com/kubemq-io/file-uploader/source"
	"github.com/kubemq-io/file-uploader/types"
	"github.com/spf13/pflag"
	"os"
	"os/signal"
	"syscall"
)

var log = logger.NewLogger("file-uploader")
var (
	generate = pflag.BoolP("generate", "g", false, "set generate files before start")
	size     = pflag.IntP("size", "s", 1000000, "set file size")
	items    = pflag.IntP("items", "i", 10, "set how many items to generate")
)

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
	if *generate {
		g := &file_creator.GeneratorRequest{
			Dir:   cfg.Source.Root,
			Size:  *size,
			Items: *items,
		}
		if err := g.Do(); err != nil {
			return err
		}
	}

	var senders []types.Sender
	for i := 0; i < cfg.Source.ConcurrentSenders; i++ {
		client := kubemq.NewClient()
		err := client.Init(ctx, cfg)
		if err != nil {
			return err
		}
		senders = append(senders, client)
	}
	sourceService := source.NewSourceService(cfg)
	sourceService.Start(ctx, senders)
	apiServer, err := api.Start(ctx, cfg, sourceService)
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
			if sourceService != nil {
				sourceService.Stop()
			}
			sourceService := source.NewSourceService(cfg)
			sourceService.Start(ctx, senders)
			apiServer, err = api.Start(ctx, newConfig, sourceService)
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
	pflag.Parse()
	log.Info("starting file-uploader")
	if err := run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
