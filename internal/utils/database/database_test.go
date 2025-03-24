package database

import (
	"testing"

	"dkhalife.com/tasks/core/config"
	"github.com/stretchr/testify/assert"
)

func TestNewDatabase(t *testing.T) {
	// Mock configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			FilePath: ":memory:", // Use in-memory SQLite database for testing
		},
	}

	// Test database connection
	db, err := NewDatabase(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Verify the connection
	sqlDB, err := db.DB()
	assert.NoError(t, err)
	assert.NoError(t, sqlDB.Ping())
}
