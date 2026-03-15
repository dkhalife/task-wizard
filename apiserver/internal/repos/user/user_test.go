package repos

import (
	"context"
	"testing"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/utils/test"
	"github.com/stretchr/testify/suite"
)

type UserTestSuite struct {
	test.DatabaseTestSuite
	repo IUserRepo
	cfg  *config.Config
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

func (s *UserTestSuite) SetupTest() {
	s.DatabaseTestSuite.SetupTest()

	s.cfg = &config.Config{
		Server: config.ServerConfig{
			Registration: true,
		},
	}
	s.repo = NewUserRepository(s.DB, s.cfg)
}

func (s *UserTestSuite) TestCreateUser() {
	ctx := context.Background()

	user := &models.User{
		DirectoryID: "test-dir",
		ObjectID:    "test-obj",
		DisplayName: "Test User",
		Email:       "test@example.com",
		CreatedAt:   time.Now(),
	}

	err := s.repo.CreateUser(ctx, user)
	s.Require().NoError(err)
	s.NotZero(user.ID)

	var fetchedUser models.User
	err = s.DB.First(&fetchedUser, user.ID).Error
	s.Require().NoError(err)
	s.Equal(user.Email, fetchedUser.Email)
	s.Equal(user.DirectoryID, fetchedUser.DirectoryID)
	s.Equal(user.ObjectID, fetchedUser.ObjectID)

	var settings models.NotificationSettings
	err = s.DB.Where("user_id = ?", user.ID).First(&settings).Error
	s.Require().NoError(err)
	s.Equal(user.ID, settings.UserID)
	s.Equal(models.NotificationProviderNone, settings.Provider.Provider)
}

func (s *UserTestSuite) TestCreateUserRegistrationDisabled() {
	ctx := context.Background()

	s.cfg.Server.Registration = false

	user := &models.User{
		DirectoryID: "test-dir",
		ObjectID:    "test-obj",
		Email:       "test@example.com",
	}

	err := s.repo.CreateUser(ctx, user)
	s.Require().Error(err)
	s.Contains(err.Error(), "registration is disabled")
}

func (s *UserTestSuite) TestGetUser() {
	ctx := context.Background()

	testUser := &models.User{
		DirectoryID: "test-dir",
		ObjectID:    "test-obj",
		Email:       "test@example.com",
		CreatedAt:   time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	user, err := s.repo.GetUser(ctx, testUser.ID)
	s.Require().NoError(err)
	s.NotNil(user)
	s.Equal(testUser.Email, user.Email)
}

func (s *UserTestSuite) TestFindByEntraID() {
	ctx := context.Background()

	testUser := &models.User{
		DirectoryID: "dir-123",
		ObjectID:    "obj-456",
		DisplayName: "Entra User",
		Email:       "entra@example.com",
		CreatedAt:   time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	user, err := s.repo.FindByEntraID(ctx, "dir-123", "obj-456")
	s.Require().NoError(err)
	s.NotNil(user)
	s.Equal(testUser.ID, user.ID)
	s.Equal("Entra User", user.DisplayName)
}

func (s *UserTestSuite) TestFindByEntraIDNotFound() {
	ctx := context.Background()

	_, err := s.repo.FindByEntraID(ctx, "nonexistent-dir", "nonexistent-obj")
	s.Require().Error(err)
}

func (s *UserTestSuite) TestEnsureUserCreatesNew() {
	ctx := context.Background()

	user, err := s.repo.EnsureUser(ctx, "new-dir", "new-obj", "New User", "new@example.com")
	s.Require().NoError(err)
	s.NotNil(user)
	s.NotZero(user.ID)
	s.Equal("new-dir", user.DirectoryID)
	s.Equal("new-obj", user.ObjectID)
	s.Equal("New User", user.DisplayName)
}

func (s *UserTestSuite) TestEnsureUserReturnsExisting() {
	ctx := context.Background()

	existing := &models.User{
		DirectoryID: "existing-dir",
		ObjectID:    "existing-obj",
		DisplayName: "Existing User",
		Email:       "existing@example.com",
		CreatedAt:   time.Now(),
	}

	err := s.DB.Create(existing).Error
	s.Require().NoError(err)

	user, err := s.repo.EnsureUser(ctx, "existing-dir", "existing-obj", "Different Name", "different@example.com")
	s.Require().NoError(err)
	s.Equal(existing.ID, user.ID)
	s.Equal("Existing User", user.DisplayName)
}

func (s *UserTestSuite) TestEnsureUserRegistrationDisabled() {
	ctx := context.Background()

	s.cfg.Server.Registration = false

	_, err := s.repo.EnsureUser(ctx, "new-dir", "new-obj", "New User", "new@example.com")
	s.Require().Error(err)
	s.Contains(err.Error(), "registration is disabled")
}

func (s *UserTestSuite) TestUpdateNotificationSettings() {
	ctx := context.Background()

	testUser := &models.User{
		DirectoryID: "test-dir",
		ObjectID:    "test-obj",
		Email:       "test@example.com",
		CreatedAt:   time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	settings := &models.NotificationSettings{
		UserID: testUser.ID,
		Provider: models.NotificationProvider{
			Provider: models.NotificationProviderNone,
		},
	}

	err = s.DB.Create(settings).Error
	s.Require().NoError(err)

	provider := models.NotificationProvider{
		Provider: models.NotificationProviderGotify,
		URL:      "https://example.com",
		Token:    "api-token",
	}

	triggers := models.NotificationTriggerOptions{
		DueDate: true,
		PreDue:  true,
	}

	err = s.repo.UpdateNotificationSettings(ctx, testUser.ID, provider, triggers)
	s.Require().NoError(err)

	var updated models.NotificationSettings
	err = s.DB.Where("user_id = ?", testUser.ID).First(&updated).Error
	s.Require().NoError(err)
	s.Equal(models.NotificationProviderGotify, updated.Provider.Provider)
	s.Equal("https://example.com", updated.Provider.URL)
	s.Equal("api-token", updated.Provider.Token)
	s.True(updated.Triggers.DueDate)
	s.True(updated.Triggers.PreDue)
}
