package database

import (
	"testing"

	"dkhalife.com/tasks/core/config"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type DatabaseTestSuite struct {
	suite.Suite
	db *gorm.DB
}

func TestDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}

func (s *DatabaseTestSuite) SetupTest() {
	// Mock configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			FilePath: ":memory:", // Use in-memory SQLite database for testing
		},
	}

	// Test database connection
	db, err := NewDatabase(cfg)
	s.Require().NoError(err)
	s.Require().NotNil(db)
	s.db = db
}

func (s *DatabaseTestSuite) TearDownTest() {
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

func (s *DatabaseTestSuite) TestNewDatabaseConnection() {
	sqlDB, err := s.db.DB()
	s.Require().NoError(err)
	s.NoError(sqlDB.Ping())
}
