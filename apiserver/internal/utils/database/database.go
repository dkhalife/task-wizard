package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/services/logging"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	var level gormLogger.LogLevel
	switch strings.ToLower(cfg.Server.LogLevel) {
	case "debug":
		level = gormLogger.Info
		logging.DefaultLogger().Error("DEBUG level set: SQL queries will be logged and may contain sensitive data")
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

	var dialector gorm.Dialector
	dbType := strings.ToLower(cfg.Database.Type)
	
	switch dbType {
	case "mysql", "mariadb":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Database,
		)
		dialector = mysql.Open(dsn)
	case "sqlite", "":
		dialector = sqlite.Open(cfg.Database.FilePath)
	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: sqlite, mysql, mariadb)", cfg.Database.Type)
	}

	db, err := gorm.Open(dialector, &gorm.Config{Logger: logger})
	if err != nil {
		return nil, err
	}

	// Only apply SQLite-specific settings if using SQLite
	if dbType == "sqlite" || dbType == "" {
		if err := db.Exec("PRAGMA journal_mode=WAL;").Error; err != nil {
			return nil, err
		}
		if err := db.Exec("PRAGMA busy_timeout=5000;").Error; err != nil {
			return nil, err
		}
	}

	return db, nil
}
