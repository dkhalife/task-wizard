package logging

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const loggerKey contextKey = "logger"

var (
	defaultLogger     *zap.SugaredLogger
	defaultLoggerOnce sync.Once
)

var conf = &Config{
	Encoding:    "console",
	Level:       zapcore.WarnLevel,
	Development: false,
}

type Config struct {
	Encoding    string
	Level       zapcore.Level
	Development bool
}

// SetConfig sets given logging configs for DefaultLogger's logger.
// Must set configs before calling DefaultLogger()
func SetConfig(c *Config) {
	conf = &Config{
		Encoding:    c.Encoding,
		Level:       c.Level,
		Development: c.Development,
	}
}

// newLogger creates a new logger with the given log level
func newLogger(conf *Config) *zap.SugaredLogger {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg := zap.Config{
		Encoding:         conf.Encoding,
		EncoderConfig:    ec,
		Level:            zap.NewAtomicLevelAt(conf.Level),
		Development:      conf.Development,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if conf.Development {
		cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	} else {
		cfg.EncoderConfig.EncodeCaller = nil
	}

	logger, err := cfg.Build()
	if err != nil {
		logger = zap.NewNop()
	}
	return logger.Sugar()
}

func DefaultLogger() *zap.SugaredLogger {
	defaultLoggerOnce.Do(func() {
		defaultLogger = newLogger(conf)
	})
	return defaultLogger
}

func ContextWithLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	if ctx == nil {
		panic("nil context passed to ContextWithLogger")
	}
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) *zap.SugaredLogger {
	if ctx == nil {
		return DefaultLogger()
	}

	if gCtx, ok := ctx.(*gin.Context); ok && gCtx != nil {
		ctx = gCtx.Request.Context()
	}

	if logger, ok := ctx.Value(loggerKey).(*zap.SugaredLogger); ok {
		return logger
	}

	return DefaultLogger()
}
