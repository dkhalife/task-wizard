package logging

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestSetConfigAndNewLogger(t *testing.T) {
	cfg := &Config{
		Encoding:    "console",
		Level:       zapcore.DebugLevel,
		Development: true,
	}
	SetConfig(cfg)
	logger := newLogger(cfg)
	assert.NotNil(t, logger)
	logger.Debug("debug message")
}

func TestDefaultLogger(t *testing.T) {
	logger := DefaultLogger()
	assert.NotNil(t, logger)
	logger.Info("info message")
}

func TestFromContext_GinContext(t *testing.T) {
	ginCtx, _ := gin.CreateTestContext(nil)
	ginCtx.Request, _ = http.NewRequest("GET", "/", nil)
	logger := FromContext(ginCtx)
	assert.NotNil(t, logger)
}
