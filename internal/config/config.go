package config

import (
	"log"
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
	ServerPort              string        `env:"SERVER_PORT"               envDefault:"8080"`
	ReadTimeout             time.Duration `env:"READ_TIMEOUT"              envDefault:"10s"`
	WriteTimeout            time.Duration `env:"WRITE_TIMEOUT"             envDefault:"10s"`
	IdleTimeout             time.Duration `env:"IDLE_TIMEOUT"              envDefault:"60s"`
	GracefulShutdownTimeout time.Duration `env:"GRACEFUL_SHUTDOWN_TIMEOUT" envDefault:"30s"`

	// db
	MongoURI            string        `env:"MONGO_URI"             envRequired:"true"`
	MongoDB             string        `env:"MONGO_DB"              envDefault:"assignment_service"`
	MongoConnectTimeout time.Duration `env:"MONGO_CONNECT_TIMEOUT" envDefault:"10s"`
}

func Load() *Config {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		log.Fatal("Failed to parse environment variables", zap.Error(err))
	}

	// server
	if port, _ := strconv.Atoi(cfg.ServerPort); port < 1 || port > 65535 {
		log.Fatal("SERVER_PORT must be between 1 and 65535", zap.Int("got", port))
	}
	if cfg.ReadTimeout < 1*time.Second {
		log.Fatal("READ_TIMEOUT must be at least 1s", zap.Int("got", int(cfg.ReadTimeout)))
	}
	if cfg.WriteTimeout < 1*time.Second {
		log.Fatal("WRITE_TIMEOUT must be at least 1s", zap.Int("got", int(cfg.WriteTimeout)))
	}
	if cfg.IdleTimeout < 1*time.Second {
		log.Fatal("IDLE_TIMEOUT must be at least 1s", zap.Int("got", int(cfg.IdleTimeout)))
	}
	if cfg.GracefulShutdownTimeout < 10*time.Second {
		log.Fatal("GRACEFUL_SHUTDOWN_TIMEOUT must be at least 10s", zap.Int("got", int(cfg.GracefulShutdownTimeout)))
	}

	// db
	if cfg.MongoDB == "" {
		log.Fatal("MONGO_DB must not be empty")
	}
	if cfg.MongoConnectTimeout < 5*time.Second {
		log.Fatal("MONGO_CONNECT_TIMEOUT must be at least 5s", zap.Int("got", int(cfg.MongoConnectTimeout)))
	}
	validateMongoURI(cfg.MongoURI)

	return cfg
}

func validateMongoURI(uri string) {
	if uri == "" {
		log.Fatal("MONGO_URI is required but empty")
	}

	if !strings.HasPrefix(uri, "mongodb://") && !strings.HasPrefix(uri, "mongodb+srv://") {
		log.Fatal("MONGO_URI must start with mongodb:// or mongodb+srv://")
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		log.Fatal("MONGO_URI is not a valid URL")
	}

	if parsed.Host == "" {
		log.Fatal("MONGO_URI must contain a host")
	}

	if strings.HasPrefix(uri, "mongodb+srv://") && !strings.Contains(parsed.Host, ".mongodb.net") && !strings.Contains(parsed.Host, "localhost") {
		log.Fatal("mongodb+srv:// URI should point to Atlas cluster (*.mongodb.net)")
	}
}

// MongoDB connection string may contain password data, so it must be hidden during logging
func maskMongoURI(uri string) string {
	if uri == "" {
		return "<empty>"
	}
	parsed, err := url.Parse(uri)
	if err != nil {
		return "<invalid uri>"
	}
	if parsed.User != nil {
		return parsed.Redacted()
	}
	return uri
}

func (c *Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	// server
	enc.AddString("server_port", c.ServerPort)
	enc.AddDuration("read_timeout", c.ReadTimeout)
	enc.AddDuration("write_timeout", c.WriteTimeout)
	enc.AddDuration("idle_timeout", c.IdleTimeout)
	enc.AddDuration("graceful_shutdown_timeout", c.GracefulShutdownTimeout)

	// db
	enc.AddString("mongo_uri", maskMongoURI(c.MongoURI))
	enc.AddString("mongo_db", c.MongoDB)
	enc.AddDuration("mongo_connect_timeout", c.MongoConnectTimeout)

	return nil
}
