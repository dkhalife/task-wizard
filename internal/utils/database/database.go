package database

import (
	"log"
	"os"
	"strings"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/services/logging"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	level := gormLogger.Warn
	switch strings.ToLower(cfg.Server.LogLevel) {
	case "debug":
		level = gormLogger.Info
		logging.DefaultLogger().Warn("DEBUG level set: SQL queries will be logged and may contain sensitive data")
	case "warn", "warning":
		level = gormLogger.Warn
	case "error":
		level = gormLogger.Error
	case "silent":
		level = gormLogger.Silent
	default:
		level = gormLogger.Warn
	}

	logger := gormLogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gormLogger.Config{
			SlowThreshold: time.Second,
			LogLevel:      level,
			Colorful:      false,
		},
	)

	db, err := gorm.Open(sqlite.Open(cfg.Database.FilePath), &gorm.Config{Logger: logger})
	if err != nil {
		return nil, err
	}

	if err := db.Exec("PRAGMA journal_mode=WAL;").Error; err != nil {
		return nil, err
	}
	if err := db.Exec("PRAGMA busy_timeout=5000;").Error; err != nil {
		return nil, err
	}

	return db, nil
}
