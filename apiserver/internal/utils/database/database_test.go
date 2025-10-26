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

func (s *DatabaseTestSuite) TestNewDatabase_MySQLMissingHost() {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:     "mysql",
			Database: "testdb",
			Username: "testuser",
		},
	}

	_, err := NewDatabase(cfg)
	s.Require().Error(err)
	s.Contains(err.Error(), "database.host is required")
}

func (s *DatabaseTestSuite) TestNewDatabase_MySQLMissingDatabase() {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:     "mysql",
			Host:     "localhost",
			Username: "testuser",
		},
	}

	_, err := NewDatabase(cfg)
	s.Require().Error(err)
	s.Contains(err.Error(), "database.database is required")
}

func (s *DatabaseTestSuite) TestNewDatabase_MySQLMissingUsername() {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:     "mysql",
			Host:     "localhost",
			Database: "testdb",
		},
	}

	_, err := NewDatabase(cfg)
	s.Require().Error(err)
	s.Contains(err.Error(), "database.username is required")
}

func (s *DatabaseTestSuite) TestNewDatabase_UnsupportedType() {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "postgresql",
		},
	}

	_, err := NewDatabase(cfg)
	s.Require().Error(err)
	s.Contains(err.Error(), "unsupported database type")
}
