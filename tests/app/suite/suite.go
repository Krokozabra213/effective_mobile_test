package suite

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/Krokozabra213/effective_mobile/internal/config"
	pgxclient "github.com/Krokozabra213/effective_mobile/pkg/pgx-client"
)

const (
	ctxTimeout = 30 * time.Second
)

type APISuite struct {
	*testing.T
	Config     *config.Config
	DB         *pgxclient.Client
	HTTPClient *Client
}

func New(t *testing.T) (context.Context, *APISuite) {
	t.Helper()
	configFile := filepath.Join("..", "..", "configs", "main.yml")
	envFile := filepath.Join("..", "..", ".env")

	cfg, err := config.Init(configFile, envFile)
	if err != nil {
		t.Fatalf("config init err: %v", err)
	}

	httpLocalhost := "localhost"
	pgxConf := pgxclient.NewPGXConfig(httpLocalhost, cfg.PG.Port, cfg.PG.User, cfg.PG.Password, cfg.PG.DBName, cfg.PG.SSLMode,
		cfg.PG.ConnectTimeout, cfg.PG.MaxConnLifeTime, cfg.PG.MaxConnIdleTime, cfg.PG.MaxConns, cfg.PG.MinConns)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	db, err := pgxclient.New(ctx, pgxConf)
	if err != nil {
		t.Fatalf("db connect err: %v", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), ctxTimeout)

	t.Cleanup(func() {
		t.Helper()
		cancel()
		db.Close()
	})

	httpAddress := fmt.Sprintf("http://%s:%s", cfg.HTTP.Host, cfg.HTTP.Port)
	client := NewClient(httpAddress, nil)

	return ctx, &APISuite{
		T:          t,
		Config:     cfg,
		DB:         db,
		HTTPClient: client,
	}
}

func (s *APISuite) CleanupTestData() error {
	_, err := s.DB.Exec(context.Background(), "TRUNCATE TABLE subscriptions RESTART IDENTITY CASCADE")
	return err
}
