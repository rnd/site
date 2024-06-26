package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/rnd/site/config"
	"github.com/rnd/site/logging"
	"github.com/rnd/site/server"
)

var version = "local"

func main() {
	logger := logging.New()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal(err.Error())
	}
	cfg.Server.Version = version

	srv := server.New(cfg.Server)

	idleConnsClosed := make(chan struct{})
	// watch for os.Interrupt signal and gracefully shutdown
	// the server.
	go func() {
		const shutdownTimeout = 10 * time.Second

		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		ctx, cancel := context.WithTimeout(
			context.Background(), shutdownTimeout,
		)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Fatal("failed to shutdown server", slog.Any("err", err))
		}
		close(idleConnsClosed)
	}()

	if err := srv.Run(); err != nil {
		logger.Fatal("server run error", slog.Any("err", err))
	}

	<-idleConnsClosed
	logger.Info("server exited gracefully")
}
