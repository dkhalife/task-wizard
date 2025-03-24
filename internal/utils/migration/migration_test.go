package migration

import (
	"testing"

	"dkhalife.com/tasks/core/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMigration(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Run the migration
	err = Migration(db)
	assert.NoError(t, err)

	// Verify that tables were created
	for _, model := range []interface{}{
		&models.User{},
		&models.UserPasswordReset{},
		&models.AppToken{},
		&models.Label{},
		&models.Task{},
		&models.TaskHistory{},
		&models.NotificationSettings{},
		&models.Notification{},
	} {
		assert.True(t, db.Migrator().HasTable(model), "Table for model should exist")
	}
}
