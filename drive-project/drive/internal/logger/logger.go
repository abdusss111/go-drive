package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger for structured logging.
type Logger struct {
	*zap.Logger
}

// Init initializes and returns a configured logger.
func Init() (*Logger, error) {
	config := zap.NewProductionConfig()
	
	// Set log level from environment
	logLevel := os.Getenv("GODRIVE_LOG_LEVEL")
	switch logLevel {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Use console encoder for development
	if os.Getenv("ENVIRONMENT") == "development" {
		config.Encoding = "console"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config.Encoding = "json"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Build logger
	zapLogger, err := config.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return nil, err
	}

	return &Logger{Logger: zapLogger}, nil
}

// Helper methods for common logging patterns
func (l *Logger) Fatal(msg string, err error) {
	if err != nil {
		l.Logger.Fatal(msg, zap.Error(err))
	} else {
		l.Logger.Fatal(msg)
	}
}

func (l *Logger) Error(msg string, err error) {
	if err != nil {
		l.Logger.Error(msg, zap.Error(err))
	} else {
		l.Logger.Error(msg)
	}
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

