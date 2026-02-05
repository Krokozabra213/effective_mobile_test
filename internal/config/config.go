package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Default values
const (
	defaultHTTPHost               = "0.0.0.0"
	defaultHTTPPort               = "8080"
	defaultHTTPWriteTimeout       = 10 * time.Second
	defaultHTTPReadTimeout        = 10 * time.Second
	defaultHTTPMaxHeaderMegabytes = 1

	defaultPGSSLMode         = "disable"
	defaultPGConnectTimeout  = 5 * time.Second
	defaultPGMaxConns        = 10
	defaultPGMinConns        = 2
	defaultPGMaxConnLifeTime = 1 * time.Hour
	defaultPGMaxConnIdleTime = 15 * time.Minute
)

type (
	Config struct {
		App  AppConfig
		HTTP HTTPConfig
		PG   PostgresConfig
	}

	AppConfig struct {
		AppSecretKey string
		Environment  string
	}

	PostgresConfig struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string

		SSLMode         string        `mapstructure:"sslMode"`
		ConnectTimeout  time.Duration `mapstructure:"connectTimeout"`
		MaxConns        int           `mapstructure:"maxConns"`
		MinConns        int           `mapstructure:"minConns"`
		MaxConnLifeTime time.Duration `mapstructure:"maxConnLifeTime"`
		MaxConnIdleTime time.Duration `mapstructure:"maxConnIdleTime"`
	}

	HTTPConfig struct {
		Host               string        `mapstructure:"host"`
		Port               string        `mapstructure:"port"`
		ReadTimeout        time.Duration `mapstructure:"readTimeout"`
		WriteTimeout       time.Duration `mapstructure:"writeTimeout"`
		MaxHeaderMegabytes int           `mapstructure:"maxHeaderBytes"`
	}
)

func newCfg() Config {
	cfg := Config{
		App:  AppConfig{},
		PG:   PostgresConfig{},
		HTTP: HTTPConfig{},
	}
	return cfg
}

// Init loads config from file and environment variables.
// Priority: defaults -> config file -> env vars
func Init(configfile, envfile string) (*Config, error) {
	populateDefault()

	if err := parseConfigFile(configfile); err != nil {
		return nil, err
	}

	cfg := newCfg()

	if err := unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := setFromEnv(envfile, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func populateDefault() {
	// http config
	viper.SetDefault("http.host", defaultHTTPHost)
	viper.SetDefault("http.port", defaultHTTPPort)
	viper.SetDefault("http.maxHeaderMegabytes", defaultHTTPMaxHeaderMegabytes)
	viper.SetDefault("http.readTimeout", defaultHTTPReadTimeout)
	viper.SetDefault("http.writeTimeout", defaultHTTPWriteTimeout)

	// Postgres defaults
	viper.SetDefault("postgres.sslMode", defaultPGSSLMode)
	viper.SetDefault("postgres.connectTimeout", defaultPGConnectTimeout)
	viper.SetDefault("postgres.maxConns", defaultPGMaxConns)
	viper.SetDefault("postgres.minConns", defaultPGMinConns)
	viper.SetDefault("postgres.maxConnLifeTime", defaultPGMaxConnLifeTime)
	viper.SetDefault("postgres.maxConnIdleTime", defaultPGMaxConnIdleTime)
}

func parseConfigFile(configPath string) error {
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

// setFromEnv loads sensitive data from environment variables
func setFromEnv(envpath string, cfg *Config) error {
	err := godotenv.Load(envpath)
	if err != nil {
		return err
	}

	cfg.App.Environment = os.Getenv("ENV")
	cfg.App.AppSecretKey = os.Getenv("APP_SECRET")

	cfg.PG.Host = os.Getenv("POSTGRES_HOST")
	cfg.PG.Port = os.Getenv("POSTGRES_PORT")
	cfg.PG.User = os.Getenv("POSTGRES_USER")
	cfg.PG.DBName = os.Getenv("POSTGRES_DB")
	cfg.PG.Password = os.Getenv("POSTGRES_PASSWORD")

	return nil
}

func unmarshal(cfg *Config) error {
	if err := viper.UnmarshalKey("http", &cfg.HTTP); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("postgres", &cfg.PG); err != nil {
		return err
	}

	return nil
}

func (c *Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("env", c.App.Environment),
		slog.Group("http",
			slog.String("http_address", c.HTTP.Host+":"+c.HTTP.Port),
			slog.Duration("read_timeout", c.HTTP.ReadTimeout),
			slog.Duration("write_timeout", c.HTTP.WriteTimeout),
			slog.Int("maxHeaderMegabytes", c.HTTP.MaxHeaderMegabytes),
		),
		slog.Group("postgres",
			slog.String("address", c.PG.Host+":"+c.PG.Port),
			slog.String("ssl_mode", c.PG.SSLMode),
			slog.Duration("connect_timeout", c.PG.ConnectTimeout),
			slog.Int("max_conns", c.PG.MaxConns),
			slog.Int("min_conns", c.PG.MinConns),
			slog.Duration("max_conn_lifetime", c.PG.MaxConnLifeTime),
			slog.Duration("max_conn_idletime", c.PG.MaxConnIdleTime),
		),
	)
}
