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
		Jwt: config.JwtConfig{
			Secret: "test-secret",
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

func (s *UserTestSuite) TestCreateAppToken() {
	ctx := context.Background()

	testUser := &models.User{
		DirectoryID: "test-dir",
		ObjectID:    "test-obj",
		Email:       "test@example.com",
		CreatedAt:   time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	scopes := []models.ApiTokenScope{models.ApiTokenScopeTaskRead, models.ApiTokenScopeTaskWrite}
	token, err := s.repo.CreateAppToken(ctx, testUser.ID, "Test Token", scopes, 30)
	s.Require().NoError(err)
	s.NotNil(token)
	s.Equal("Test Token", token.Name)
	s.Equal(testUser.ID, token.UserID)
	s.NotEmpty(token.Token)
	s.NotZero(token.ExpiresAt)
}

func (s *UserTestSuite) TestCreateAppTokenInvalidScope() {
	ctx := context.Background()

	testUser := &models.User{
		DirectoryID: "test-dir",
		ObjectID:    "test-obj",
		Email:       "test@example.com",
		CreatedAt:   time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	scopes := []models.ApiTokenScope{models.ApiTokenScopeUserRead}
	_, err = s.repo.CreateAppToken(ctx, testUser.ID, "Test Token", scopes, 30)
	s.Require().Error(err)
	s.Contains(err.Error(), "user scopes are not allowed")

	scopes = []models.ApiTokenScope{models.ApiTokenScopeTokenWrite}
	_, err = s.repo.CreateAppToken(ctx, testUser.ID, "Test Token", scopes, 30)
	s.Require().Error(err)
	s.Contains(err.Error(), "token scopes are not allowed")
}

func (s *UserTestSuite) TestGetAllUserTokens() {
	ctx := context.Background()

	testUser := &models.User{
		DirectoryID: "test-dir",
		ObjectID:    "test-obj",
		Email:       "test@example.com",
		CreatedAt:   time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	token1 := &models.AppToken{
		UserID:    testUser.ID,
		Name:      "Token 1",
		Token:     "token1",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Scopes:    []string{"task:read"},
	}

	token2 := &models.AppToken{
		UserID:    testUser.ID,
		Name:      "Token 2",
		Token:     "token2",
		ExpiresAt: time.Now().Add(48 * time.Hour),
		Scopes:    []string{"task:write"},
	}

	err = s.DB.Create(token1).Error
	s.Require().NoError(err)

	err = s.DB.Create(token2).Error
	s.Require().NoError(err)

	tokens, err := s.repo.GetAllUserTokens(ctx, testUser.ID)
	s.Require().NoError(err)
	s.Require().Len(tokens, 2)

	s.Equal(token1.ID, tokens[0].ID)
	s.Equal(token2.ID, tokens[1].ID)
}

func (s *UserTestSuite) TestDeleteAppToken() {
	ctx := context.Background()

	testUser := &models.User{
		DirectoryID: "test-dir",
		ObjectID:    "test-obj",
		Email:       "test@example.com",
		CreatedAt:   time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	token := &models.AppToken{
		UserID:    testUser.ID,
		Name:      "Test Token",
		Token:     "token123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Scopes:    []string{"task:read"},
	}

	err = s.DB.Create(token).Error
	s.Require().NoError(err)

	err = s.repo.DeleteAppToken(ctx, testUser.ID, token.ID)
	s.Require().NoError(err)

	var count int64
	err = s.DB.Model(&models.AppToken{}).Where("id = ?", token.ID).Count(&count).Error
	s.Require().NoError(err)
	s.Equal(int64(0), count)
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

func (s *UserTestSuite) TestGetAppTokensNearingExpiration() {
	ctx := context.Background()

	testUser := &models.User{
		DirectoryID: "test-dir",
		ObjectID:    "test-obj",
		Email:       "test@example.com",
		CreatedAt:   time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	now := time.Now()

	token1 := &models.AppToken{
		UserID:    testUser.ID,
		Name:      "Expired Token",
		Token:     "token1",
		ExpiresAt: now.Add(-48 * time.Hour),
	}

	token2 := &models.AppToken{
		UserID:    testUser.ID,
		Name:      "Nearing Expiration",
		Token:     "token2",
		ExpiresAt: now.Add(6 * time.Hour),
	}

	token3 := &models.AppToken{
		UserID:    testUser.ID,
		Name:      "Valid Token",
		Token:     "token3",
		ExpiresAt: now.Add(30 * 24 * time.Hour),
	}

	err = s.DB.Create([]*models.AppToken{token1, token2, token3}).Error
	s.Require().NoError(err)

	err = s.DB.Exec("UPDATE app_tokens SET user_id = ?", testUser.ID).Error
	s.Require().NoError(err)

	tokens, err := s.repo.GetAppTokensNearingExpiration(ctx, 12*time.Hour)
	s.Require().NoError(err)
	s.Require().Len(tokens, 1)
	s.Equal("Nearing Expiration", tokens[0].Name)
}

func (s *UserTestSuite) TestDeleteStaleAppTokens() {
	ctx := context.Background()

	testUser := &models.User{
		DirectoryID: "test-dir",
		ObjectID:    "test-obj",
		Email:       "test@example.com",
		CreatedAt:   time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	now := time.Now()

	token1 := &models.AppToken{
		UserID:    testUser.ID,
		Name:      "Expired Token",
		Token:     "token1",
		ExpiresAt: now.Add(-48 * time.Hour),
	}

	token2 := &models.AppToken{
		UserID:    testUser.ID,
		Name:      "Valid Token",
		Token:     "token2",
		ExpiresAt: now.Add(24 * time.Hour),
	}

	err = s.DB.Create([]*models.AppToken{token1, token2}).Error
	s.Require().NoError(err)

	err = s.repo.DeleteStaleAppTokens(ctx)
	s.Require().NoError(err)

	var count int64
	err = s.DB.Model(&models.AppToken{}).Count(&count).Error
	s.Require().NoError(err)
	s.Equal(int64(1), count)

	var remaining models.AppToken
	err = s.DB.First(&remaining).Error
	s.Require().NoError(err)
	s.Equal("Valid Token", remaining.Name)
}
