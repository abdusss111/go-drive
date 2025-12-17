package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init returns a production-ready JSON logger writing to stdout.
func Init() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "json"
	// Keep timestamps in ISO-like format for external collectors.
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg.Build()
}
