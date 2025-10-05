package main

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/services/scheduler"
	"go.uber.org/fx"
)

type mockLifecycle struct{}

func (t *mockLifecycle) Append(hook fx.Hook) {}

func TestNewServer_DebugMode(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			LogLevel:     "debug",
			Port:         8080,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
		Database: config.DatabaseConfig{},
	}
	var db gorm.DB
	ms := &scheduler.Scheduler{}
	lc := &mockLifecycle{}
	r := newServer(lc, cfg, &db, ms)
	assert.NotNil(t, r)
	assert.IsType(t, &gin.Engine{}, r)
}

func TestNewServer_ReleaseMode(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			LogLevel:     "info",
			Port:         8080,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
		Database: config.DatabaseConfig{},
	}
	var db gorm.DB
	ms := &scheduler.Scheduler{}
	lc := &mockLifecycle{}
	r := newServer(lc, cfg, &db, ms)
	assert.NotNil(t, r)
	assert.IsType(t, &gin.Engine{}, r)
}
