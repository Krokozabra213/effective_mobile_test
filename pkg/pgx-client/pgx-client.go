package pgxclient

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host            string
	Port            string
	User            string
	Password        string
	Database        string
	SSLMode         string
	ConnectTimeout  time.Duration
	MaxConns        int32
	MinConns        int32
	MaxConnLifeTime time.Duration
	MaxConnIdleTime time.Duration
}

func NewPGXConfig(host, port, user, password, database, sslMode string, connectTimeout, maxConnLifeTime, maxConnIdleTime time.Duration,
	maxConns, minConns int,
) Config {
	return Config{
		Host:            host,
		Port:            port,
		User:            user,
		Password:        password,
		Database:        database,
		SSLMode:         sslMode,
		ConnectTimeout:  connectTimeout,
		MaxConnLifeTime: maxConnLifeTime,
		MaxConnIdleTime: maxConnIdleTime,
		MaxConns:        int32(maxConns),
		MinConns:        int32(minConns),
	}
}

// DSN с экранированием пароля
func (c Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.QueryEscape(c.User),
		url.QueryEscape(c.Password),
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
	)
}

type Client struct {
	*pgxpool.Pool
}

func New(ctx context.Context, cfg Config) (*Client, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifeTime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	// Таймаут на подключение
	connectCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Client{
		Pool: pool,
	}, nil
}

