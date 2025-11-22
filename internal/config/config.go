package config

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	// server
	ServerPort              string        `env:"SERVER_PORT" envDefault:"8080"`
	ReadTimeout             time.Duration `env:"READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout            time.Duration `env:"WRITE_TIMEOUT" envDefault:"10s"`
	IdleTimeout             time.Duration `env:"IDLE_TIMEOUT" envDefault:"60s"`
	GracefulShutdownTimeout time.Duration `env:"GRACEFUL_SHUTDOWN_TIMEOUT" envDefault:"30s"`

	// db
	MongoURI            string        `env:"MONGO_URI" envRequired:"true"`
	MongoDB             string        `env:"MONGO_DB" envDefault:"assignment_service"`
	MongoConnectTimeout time.Duration `env:"MONGO_CONNECT_TIMEOUT" envDefault:"10s"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse env: %w", err)
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

var fatalLogFunc = func(logger *zap.Logger, msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

func MustLoad(logger *zap.Logger) *Config {
	cfg, err := Load()
	if err != nil {
		fatalLogFunc(logger, "Invalid configuration", zap.Error(err))
	}
	return cfg
}

func validateConfig(c *Config) error {
	/// server
	port, err := strconv.Atoi(c.ServerPort)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("SERVER_PORT must be a valid port (1-65535), got: %s", c.ServerPort)
	}

	if c.ReadTimeout < time.Second {
		return fmt.Errorf("READ_TIMEOUT must be >= 1s, got: %v", c.ReadTimeout)
	}
	if c.WriteTimeout < time.Second {
		return fmt.Errorf("WRITE_TIMEOUT must be >= 1s, got: %v", c.WriteTimeout)
	}
	if c.IdleTimeout < time.Second {
		return fmt.Errorf("IDLE_TIMEOUT must be >= 1s, got: %v", c.IdleTimeout)
	}
	if c.GracefulShutdownTimeout < 10*time.Second {
		return fmt.Errorf("GRACEFUL_SHUTDOWN_TIMEOUT must be >= 10s, got: %v", c.GracefulShutdownTimeout)
	}

	// db
	if strings.TrimSpace(c.MongoDB) == "" {
		return fmt.Errorf("MONGO_DB must not be empty")
	}
	if c.MongoConnectTimeout < 5*time.Second {
		return fmt.Errorf("MONGO_CONNECT_TIMEOUT must be >= 5s, got: %v", c.MongoConnectTimeout)
	}

	return validateMongoURI(c.MongoURI)
}

func validateMongoURI(uri string) error {
	if uri == "" {
		return fmt.Errorf("MONGO_URI is required")
	}
	if !strings.HasPrefix(uri, "mongodb://") && !strings.HasPrefix(uri, "mongodb+srv://") {
		return fmt.Errorf("MONGO_URI must start with mongodb:// or mongodb+srv://, got: %s", uri)
	}

	u, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("MONGO_URI is invalid URL: %w", err)
	}
	if u.Host == "" {
		return fmt.Errorf("MONGO_URI must contain a host")
	}
	if strings.HasPrefix(uri, "mongodb+srv://") {
		host := u.Host
		isLocal := strings.Contains(host, "localhost") || strings.HasPrefix(host, "127.") || host == "localhost"
		if !isLocal && !strings.Contains(host, ".mongodb.net") {
			return fmt.Errorf("mongodb+srv:// URI should point to Atlas (*.mongodb.net) or localhost, got host: %s", host)
		}
	}
	return nil
}

// MongoDB connection string may contain password data, so it must be hidden during logging
func maskMongoURI(uri string) string {
	if uri == "" {
		return "<empty>"
	}
	u, err := url.Parse(uri)
	if err != nil {
		return "<invalid uri>"
	}
	if u.User != nil {
		return u.Redacted()
	}
	return uri
}

func (c *Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("server_port", c.ServerPort)
	enc.AddDuration("read_timeout", c.ReadTimeout)
	enc.AddDuration("write_timeout", c.WriteTimeout)
	enc.AddDuration("idle_timeout", c.IdleTimeout)
	enc.AddDuration("graceful_shutdown_timeout", c.GracefulShutdownTimeout)
	enc.AddString("mongo_uri", maskMongoURI(c.MongoURI))
	enc.AddString("mongo_db", c.MongoDB)
	enc.AddDuration("mongo_connect_timeout", c.MongoConnectTimeout)
	return nil
}
