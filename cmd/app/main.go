package app

import (
	"context"
	"log/slog"
	"os"

	"github.com/Krokozabra213/effective_mobile/internal/config"
	"github.com/Krokozabra213/effective_mobile/pkg/logger"
	pgxclient "github.com/Krokozabra213/effective_mobile/pkg/pgx-client"
)

const (
	configFile = "configs/main.yml"
	envFile    = ".env"
)

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
	pgxConf := pgxclient.NewPGXConfig(cfg.PG.Host, cfg.PG.Port, cfg.PG.User, cfg.PG.Password,
		cfg.PG.DBName, cfg.PG.SSLMode, cfg.PG.ConnectTimeout, cfg.PG.MaxConnLifeTime,
		cfg.PG.MaxConnIdleTime, cfg.PG.MaxConns, cfg.PG.MinConns,
	)

	dbClient, err := pgxclient.New(context.Background(), pgxConf)
	if err != nil {
		return err
	}
	defer dbClient.Close()
	log.Info("connected to postgres")

	return nil
}
