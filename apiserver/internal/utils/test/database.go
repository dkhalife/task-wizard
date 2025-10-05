package test

import (
	"fmt"
	"log"
	"os"
	"time"

	"dkhalife.com/tasks/core/config"
	dbutil "dkhalife.com/tasks/core/internal/utils/database"
	"dkhalife.com/tasks/core/internal/utils/migration"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type DatabaseTestSuite struct {
	suite.Suite
	DB         *gorm.DB
	dbFilePath string
}

func (suite *DatabaseTestSuite) SetupTest() {
	suite.dbFilePath = fmt.Sprintf("%s/testdb_%d.db", os.TempDir(), time.Now().UnixNano())
	cfg := &config.Config{Database: config.DatabaseConfig{FilePath: suite.dbFilePath}}
	db, err := dbutil.NewDatabase(cfg)
	suite.Require().NoError(err)

	err = migration.Migration(db)
	suite.Require().NoError(err)

	suite.DB = db
}

func (suite *DatabaseTestSuite) TearDownTest() {
	// Close the database connection
	db, err := suite.DB.DB()
	if err != nil {
		log.Printf("failed to get database connection: %v", err)
		return
	}
	db.Close()

	// Remove the temporary database file created for the test
	if suite.dbFilePath != "" {
		if err := os.Remove(suite.dbFilePath); err != nil && !os.IsNotExist(err) {
			suite.Require().NoError(err, fmt.Sprintf("failed to remove db file %s", suite.dbFilePath))
		}
	}
}
