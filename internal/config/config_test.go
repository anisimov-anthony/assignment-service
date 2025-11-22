package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLoadEnvParseError(t *testing.T) {
	os.Clearenv()
	os.Setenv("MONGO_URI", "mongodb://localhost")

	// READ_TIMEOUT for invalid duration -> env.Parse failed
	os.Setenv("READ_TIMEOUT", "this-is-not-a-duration")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse env")
	assert.Contains(t, err.Error(), "invalid duration")
}

func TestMustLoadCallsFatalOnError(t *testing.T) {
	old := fatalLogFunc
	defer func() { fatalLogFunc = old }()

	called := false
	fatalLogFunc = func(logger *zap.Logger, msg string, fields ...zap.Field) {
		called = true
		assert.Contains(t, msg, "Invalid configuration")
		assert.NotNil(t, fields)
	}

	os.Clearenv()
	MustLoad(zap.NewNop())

	assert.True(t, called)
}

func TestMustLoadSuccessNoFatal(t *testing.T) {
	old := fatalLogFunc
	defer func() { fatalLogFunc = old }()

	called := false
	fatalLogFunc = func(_ *zap.Logger, _ string, _ ...zap.Field) {
		called = true
	}

	os.Clearenv()
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")

	MustLoad(zap.NewNop())

	assert.False(t, called)
}

func TestLoadAllErrorPaths(t *testing.T) {
	cases := []struct {
		name    string
		setup   func()
		wantErr string
	}{
		{
			"invalid port > 65535",
			func() {
				os.Setenv("SERVER_PORT", "65536")
			},
			"SERVER_PORT must be a valid port (1-65535)",
		},
		{
			"invalid port < 1",
			func() {
				os.Setenv("SERVER_PORT", "0")
			},
			"SERVER_PORT must be a valid port (1-65535)",
		},
		{
			"non-numeric port",
			func() {
				os.Setenv("SERVER_PORT", "a")
			},
			"SERVER_PORT must be a valid port (1-65535)",
		},
		{
			"read timeout low",
			func() {
				os.Setenv("READ_TIMEOUT", "999ms")
			},
			"READ_TIMEOUT must be >= 1s",
		},
		{
			"write timeout low",
			func() {
				os.Setenv("WRITE_TIMEOUT", "999ms")
			},
			"WRITE_TIMEOUT must be >= 1s",
		},
		{
			"idle timeout low",
			func() {
				os.Setenv("IDLE_TIMEOUT", "999ms")
			},
			"IDLE_TIMEOUT must be >= 1s",
		},
		{
			"graceful shutdown low",
			func() {
				os.Setenv("GRACEFUL_SHUTDOWN_TIMEOUT", "9.999s")
			},
			"GRACEFUL_SHUTDOWN_TIMEOUT must be >= 10s",
		},
		{
			"empty mongo db",
			func() {
				os.Setenv("MONGO_DB", "   ")
			},
			"MONGO_DB must not be empty",
		},
		{
			"mongo connect low",
			func() {
				os.Setenv("MONGO_CONNECT_TIMEOUT", "4.999s")
			},
			"MONGO_CONNECT_TIMEOUT must be >= 5s",
		},
		{
			"invalid uri scheme",
			func() {
				os.Setenv("MONGO_URI", "http://bad")
			},
			"MONGO_URI must start with mongodb:// or mongodb+srv://",
		},
		{
			"invalid url syntax",
			func() {
				os.Setenv("MONGO_URI", "mongodb://%")
			},
			"MONGO_URI is invalid URL",
		},
		{
			"no host",
			func() {
				os.Setenv("MONGO_URI", "mongodb://")
			},
			"MONGO_URI must contain a host",
		},
		{
			"srvs bad host",
			func() {
				os.Setenv("MONGO_URI", "mongodb+srv://bad.com")
			},
			"mongodb+srv:// URI should point to Atlas (*.mongodb.net) or localhost",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			tc.setup()
			_, err := Load()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestLoadSuccess(t *testing.T) {
	os.Clearenv()
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestMaskMongoURI(t *testing.T) {
	assert.Equal(t, "<empty>", maskMongoURI(""))
	assert.Equal(t, "<invalid uri>", maskMongoURI("mongodb://[::1]:invalid"))
	assert.Contains(t, maskMongoURI("mongodb://user:pass@localhost"), "xxxxx")
	assert.Equal(t, "mongodb://localhost", maskMongoURI("mongodb://localhost"))
}

func TestConfigMarshalLogObject(t *testing.T) {
	cfg := &Config{
		ServerPort:              "8080",
		ReadTimeout:             15 * time.Second,
		WriteTimeout:            15 * time.Second,
		IdleTimeout:             60 * time.Second,
		GracefulShutdownTimeout: 30 * time.Second,
		MongoURI:                "mongodb://u:p@localhost",
		MongoDB:                 "db",
		MongoConnectTimeout:     20 * time.Second,
	}

	enc := zapcore.NewMapObjectEncoder()

	require.NoError(t, cfg.MarshalLogObject(enc))

	assert.Equal(t, "8080", enc.Fields["server_port"])
	assert.Equal(t, 15*time.Second, enc.Fields["read_timeout"])
	assert.Contains(t, enc.Fields["mongo_uri"], "xxxxx")
	assert.Equal(t, "db", enc.Fields["mongo_db"])
}
