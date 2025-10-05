package migration

import (
	"testing"

	"dkhalife.com/tasks/core/internal/models"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MigrationTestSuite struct {
	suite.Suite
	db *gorm.DB
}

func TestMigrationTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}

func (s *MigrationTestSuite) SetupTest() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err)
	s.db = db
}

func (s *MigrationTestSuite) TearDownTest() {
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

func (s *MigrationTestSuite) TestMigration() {
	// Run the migration
	err := Migration(s.db)
	s.Require().NoError(err)

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
		s.True(s.db.Migrator().HasTable(model), "Table for model %T should exist", model)
	}
}
