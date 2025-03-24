package database

import (
	"dkhalife.com/tasks/core/config"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(cfg.Database.FilePath), &gorm.Config{})
}
