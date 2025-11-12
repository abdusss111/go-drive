package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config aggregates runtime configuration for the GoDrive API.
type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	MinIO    MinIOConfig
	Auth     AuthConfig
	Metrics  MetricsConfig
}

// ServerConfig parameterizes the HTTP server.
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Address returns the listen address in host:port form.
func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// PostgresConfig contains PostgreSQL connection details.
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// DSN returns the PostgreSQL DSN string.
func (p PostgresConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.Database, p.SSLMode)
}

// MinIOConfig carries MinIO connection and bucket information.
type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	UseSSL          bool
	Region          string
}

// AuthConfig groups authentication-related settings.
type AuthConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	BcryptCost         int
}

// MetricsConfig groups observability settings.
type MetricsConfig struct {
	PrometheusPath string
}

// Load reads configuration values from environment variables, applying defaults.
func Load() (Config, error) {
	cfg := Config{
		Server: ServerConfig{
			Host:         getString("GODRIVE_API_HOST", "0.0.0.0"),
			Port:         getInt("GODRIVE_API_PORT", 8080),
			ReadTimeout:  getDuration("GODRIVE_API_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDuration("GODRIVE_API_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDuration("GODRIVE_API_IDLE_TIMEOUT", 60*time.Second),
		},
		Postgres: PostgresConfig{
			Host:     getString("POSTGRES_HOST", "localhost"),
			Port:     getInt("POSTGRES_PORT", 5432),
			User:     getString("POSTGRES_USER", "godrive_app"),
			Password: getString("POSTGRES_PASSWORD", "change-me"),
			Database: getString("POSTGRES_DB", "godrive"),
			SSLMode:  strings.ToLower(getString("POSTGRES_SSL_MODE", "disable")),
		},
		MinIO: MinIOConfig{
			Endpoint:        getString("MINIO_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getString("MINIO_ROOT_USER", "godrive"),
			SecretAccessKey: getString("MINIO_ROOT_PASSWORD", "change-me-strong-password"),
			Bucket:          getString("MINIO_BUCKET", "godrive"),
			UseSSL:          getBool("MINIO_USE_SSL", false),
			Region:          getString("MINIO_REGION", ""),
		},
		Auth: loadAuthConfig(),
		Metrics: MetricsConfig{
			PrometheusPath: getString("GODRIVE_METRICS_PATH", "/metrics"),
		},
	}

	return cfg, nil
}

func getString(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		val = strings.ToLower(strings.TrimSpace(val))
		switch val {
		case "1", "true", "t", "yes", "y":
			return true
		case "0", "false", "f", "no", "n":
			return false
		}
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		if parsed, err := time.ParseDuration(val); err == nil {
			return parsed
		}
	}
	return fallback
}

func loadAuthConfig() AuthConfig {
	cost := getInt("GODRIVE_AUTH_BCRYPT_COST", 12)
	if cost < 4 || cost > 31 {
		cost = 12
	}

	return AuthConfig{
		AccessTokenSecret:  getString("GODRIVE_JWT_SECRET", "change-me-to-a-32-byte-secret"),
		RefreshTokenSecret: getString("GODRIVE_JWT_REFRESH_SECRET", "change-me-to-a-64-byte-secret"),
		AccessTokenTTL:     getDuration("GODRIVE_AUTH_ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:    getDuration("GODRIVE_AUTH_REFRESH_TOKEN_TTL", 720*time.Hour),
		BcryptCost:         cost,
	}
}
