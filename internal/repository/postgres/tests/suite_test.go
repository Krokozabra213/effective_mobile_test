//go:build integration

package tests

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	repository "github.com/Krokozabra213/effective_mobile/internal/repository/postgres"
	pgxclient "github.com/Krokozabra213/effective_mobile/pkg/pgx-client"
	migrations "github.com/Krokozabra213/effective_mobile/sql/goose"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testRepo *repository.PostgresRepository

const (
	dbname   = "test_db"
	username = "test"
	password = "test"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Запускаем PostgreSQL
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		fmt.Printf("failed to start container: %v\n", err)
		os.Exit(1)
	}

	host, err := container.Host(ctx)
	if err != nil {
		fmt.Printf("failed to get host: %v\n", err)
		os.Exit(1)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		fmt.Printf("failed to get port: %v\n", err)
		os.Exit(1)
	}

	cfg := pgxclient.NewPGXConfig(host, port.Port(), username, password, dbname, "disable", 5*time.Second,
		1*time.Hour, 10*time.Minute, 10, 2)

	client, err := pgxclient.New(context.Background(), cfg)
	if err != nil {
		fmt.Printf("failed to connect pgx: %v\n", err)
		os.Exit(1)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Printf("failed to get connection string: %v\n", err)
		os.Exit(1)
	}

	// Запускаем goose миграции
	if err := runMigrations(ctx, connStr); err != nil {
		fmt.Printf("failed to migrate: %v\n", err)
		os.Exit(1)
	}

	testRepo = repository.NewRepository(client)

	code := m.Run()

	client.Close()
	container.Terminate(ctx)
	os.Exit(code)
}

func runMigrations(ctx context.Context, connStr string) error {
	// Goose работает с database/sql
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	// Используем embed.FS с миграциями
	goose.SetBaseFS(migrations.Files)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}

func cleanup(t *testing.T) {
	t.Helper()
	_, err := testRepo.DB.Exec(context.Background(), "TRUNCATE TABLE subscriptions RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
}
