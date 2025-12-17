package logger

import (
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	CorrelationIDHeader = "X-Correlation-ID"
	CorrelationIDKey    = "correlation_id"
	loggerKey           = "logger"
)

var baseLogger *zap.Logger

func Init() (*zap.Logger, error) {
	if baseLogger != nil {
		return baseLogger, nil
	}

	level := logLevelFromEnv()

	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Encoding:    "json",
		OutputPaths: []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:       "timestamp",
			LevelKey:      "level",
			MessageKey:    "message",
			CallerKey:     "caller",
			StacktraceKey: "stack",
			EncodeTime:    zapcore.ISO8601TimeEncoder,
			EncodeLevel:   zapcore.LowercaseLevelEncoder,
			EncodeCaller:  zapcore.ShortCallerEncoder,
		},
	}

	l, err := cfg.Build(zap.AddCaller())
	if err != nil {
		return nil, err
	}

	baseLogger = l
	return l, nil
}

func L() *zap.Logger {
	if baseLogger != nil {
		return baseLogger
	}
	l, err := Init()
	if err != nil {
		return zap.NewNop()
	}
	return l
}

func logLevelFromEnv() zapcore.Level {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		return zap.DebugLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		corrID := c.Request.Header.Get(CorrelationIDHeader)
		if corrID == "" {
			corrID = uuid.NewString()
		}

		c.Writer.Header().Set(CorrelationIDHeader, corrID)
		c.Set(CorrelationIDKey, corrID)

		reqLogger := L().With(zap.String(CorrelationIDKey, corrID))
		c.Set(loggerKey, reqLogger)

		c.Next()

		if time.Since(start) > time.Second {
			reqLogger.Warn(
				"slow request",
				zap.String("method", c.Request.Method),
				zap.String("path", c.FullPath()),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("duration", time.Since(start)),
			)
		}
	}
}

func FromContext(c *gin.Context) *zap.Logger {
	if v, ok := c.Get(loggerKey); ok {
		if l, ok := v.(*zap.Logger); ok {
			return l
		}
	}
	return L()
}
