package database

import (
	"dkhalife.com/tasks/core/config"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	db, err = gorm.Open(sqlite.Open(cfg.Database.FilePath), &gorm.Config{})

	if err != nil {
		return nil, err
	}
	return db, nil
}
