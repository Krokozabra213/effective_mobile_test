package config

import (
	"log/slog"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	App  AppConfig      `yaml:"app"`
	HTTP HTTPConfig     `yaml:"http"`
	PG   PostgresConfig `yaml:"postgres"`
}

// AppConfig — sensitive data only from ENV
type AppConfig struct {
	Environment  string `env:"ENV" env-default:"development"`
	AppSecretKey string `env:"APP_SECRET" env-required:"true"`
}

// PostgresConfig — credentials from ENV, pool settings from YAML
type PostgresConfig struct {
	// Sensitive — from ENV only
	Host     string `env:"POSTGRES_HOST" env-required:"true"`
	Port     string `env:"POSTGRES_PORT" env-default:"5432"`
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DBName   string `env:"POSTGRES_DB" env-required:"true"`

	// Non-sensitive — from YAML (can override via ENV if needed)
	SSLMode         string        `yaml:"sslMode" env:"PG_SSL_MODE" env-default:"disable"`
	ConnectTimeout  time.Duration `yaml:"connectTimeout" env:"PG_CONNECT_TIMEOUT" env-default:"5s"`
	MaxConns        int           `yaml:"maxConns" env:"PG_MAX_CONNS" env-default:"10"`
	MinConns        int           `yaml:"minConns" env:"PG_MIN_CONNS" env-default:"2"`
	MaxConnLifeTime time.Duration `yaml:"maxConnLifeTime" env:"PG_MAX_CONN_LIFETIME" env-default:"1h"`
	MaxConnIdleTime time.Duration `yaml:"maxConnIdleTime" env:"PG_MAX_CONN_IDLE_TIME" env-default:"15m"`
}

// HTTPConfig — from YAML (can override via ENV if needed)
type HTTPConfig struct {
	Host               string        `yaml:"host" env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port               string        `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
	ReadTimeout        time.Duration `yaml:"readTimeout" env:"HTTP_READ_TIMEOUT" env-default:"10s"`
	WriteTimeout       time.Duration `yaml:"writeTimeout" env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	MaxHeaderMegabytes int           `yaml:"maxHeaderBytes" env:"HTTP_MAX_HEADER_BYTES" env-default:"1"`
}

// MustInit loads config or panics — use in main()
func MustInit(configFile, envFile string) *Config {
	cfg, err := Init(configFile, envFile)
	if err != nil {
		panic("config: " + err.Error())
	}
	return cfg
}

// Init loads config from YAML file and ENV variables.
// Priority: defaults -> YAML file -> ENV variables
func Init(configFile, envFile string) (*Config, error) {
	// 1. Load .env into os.Environ (before cleanenv reads ENV)
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			return nil, err
		}
	}

	var cfg Config

	// 2. Read YAML + apply ENV overrides + validate required fields
	if err := cleanenv.ReadConfig(configFile, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LogValue implements slog.LogValuer for safe logging (no secrets)
func (c *Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("env", c.App.Environment),
		slog.Group("http",
			slog.String("address", c.HTTP.Host+":"+c.HTTP.Port),
			slog.Duration("read_timeout", c.HTTP.ReadTimeout),
			slog.Duration("write_timeout", c.HTTP.WriteTimeout),
			slog.Int("max_header_megabytes", c.HTTP.MaxHeaderMegabytes),
		),
		slog.Group("postgres",
			slog.String("address", c.PG.Host+":"+c.PG.Port),
			slog.String("database", c.PG.DBName),
			slog.String("user", c.PG.User),
			// Password intentionally omitted!
			slog.String("ssl_mode", c.PG.SSLMode),
			slog.Duration("connect_timeout", c.PG.ConnectTimeout),
			slog.Int("max_conns", c.PG.MaxConns),
			slog.Int("min_conns", c.PG.MinConns),
			slog.Duration("max_conn_lifetime", c.PG.MaxConnLifeTime),
			slog.Duration("max_conn_idle_time", c.PG.MaxConnIdleTime),
		),
	)
}
