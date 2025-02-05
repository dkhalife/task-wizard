package database

import (
	"os"

	"donetick.com/core/config"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	path := os.Getenv("DT_SQLITE_PATH")
	if path == "" {
		path = "donetick.db"
	}
	db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})

	if err != nil {
		return nil, err
	}
	return db, nil
}
