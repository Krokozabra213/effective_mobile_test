package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Krokozabra213/effective_mobile/internal/business"
	"github.com/Krokozabra213/effective_mobile/internal/config"
	handler "github.com/Krokozabra213/effective_mobile/internal/delivery/http"
	"github.com/Krokozabra213/effective_mobile/internal/repository/postgres"
	httpserver "github.com/Krokozabra213/effective_mobile/internal/server/http"
	"github.com/Krokozabra213/effective_mobile/pkg/logger"
	pgxclient "github.com/Krokozabra213/effective_mobile/pkg/pgx-client"
)

const (
	configFile      = "configs/main.yml"
	envFile         = ".env"
	shutdownTimeout = 5 * time.Second
)

// @title           Subscription API
// @version         1.0
// @description     API для управления подписками пользователей

// @host      localhost:8080
// @BasePath  /
func main() {
	if err := run(); err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Config
	cfg, err := config.Init(configFile, envFile)
	if err != nil {
		return err
	}

	// Logger
	log := logger.Init(cfg.App.Environment)
	log.Info("initialized config", "config", cfg.LogValue())
	log.Info("starting application")

	// Database
	pgxConf := pgxclient.NewPGXConfig(cfg.PG.Host, cfg.PG.Port, cfg.PG.User, cfg.PG.Password, cfg.PG.DBName, cfg.PG.SSLMode,
		cfg.PG.ConnectTimeout, cfg.PG.MaxConnLifeTime, cfg.PG.MaxConnIdleTime, cfg.PG.MaxConns, cfg.PG.MinConns)

	dbClient, err := pgxclient.New(context.Background(), pgxConf)
	if err != nil {
		return err
	}
	defer dbClient.Close()
	log.Info("connected to postgres")

	// Dependencies
	repo := postgres.NewRepository(dbClient)
	biz := business.New(log, repo)

	// Router
	mux := http.NewServeMux()

	// Handler
	handler.New(mux, biz)

	// Server
	srv := httpserver.NewServer(cfg, mux)

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		log.Info("server started", "address", srv.Addr())
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	// Wait for shutdown signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-quit:
		log.Info("received shutdown signal", "signal", sig)
	case err := <-errCh:
		log.Error("server error", "error", err)
		return err
	}

	// Graceful shutdown
	if err := srv.ShutDown(shutdownTimeout); err != nil {
		log.Error("server shutdown error", "error", err)
		return err
	}

	return nil
}
