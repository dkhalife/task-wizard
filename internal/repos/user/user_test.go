package repos

import (
	"context"
	"fmt"
	"testing"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/utils/test"
	"github.com/lib/pq"
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

	// Create mock config
	s.cfg = &config.Config{
		Server: config.ServerConfig{
			Registration: true,
		},
		SchedulerJobs: config.SchedulerConfig{
			PasswordResetValidity: 24 * time.Hour,
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
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	err := s.repo.CreateUser(ctx, user)
	s.Require().NoError(err)
	s.NotZero(user.ID)

	// Verify user was created
	var fetchedUser models.User
	err = s.DB.First(&fetchedUser, user.ID).Error
	s.Require().NoError(err)
	s.Equal(user.Email, fetchedUser.Email)

	// Verify notification settings were created
	var settings models.NotificationSettings
	err = s.DB.Where("user_id = ?", user.ID).First(&settings).Error
	s.Require().NoError(err)
	s.Equal(user.ID, settings.UserID)
	s.Equal(models.NotificationProviderNone, settings.Provider.Provider)
}

func (s *UserTestSuite) TestCreateUserRegistrationDisabled() {
	ctx := context.Background()

	// Disable registration
	s.cfg.Server.Registration = false

	user := &models.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
	}

	err := s.repo.CreateUser(ctx, user)
	s.Require().Error(err)
	s.Contains(err.Error(), "registration is disabled")
}

func (s *UserTestSuite) TestGetUser() {
	ctx := context.Background()

	testUser := &models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	user, err := s.repo.GetUser(ctx, testUser.ID)
	s.Require().NoError(err)
	s.NotNil(user)
	s.Equal(testUser.Email, user.Email)
}

func (s *UserTestSuite) TestFindByEmail() {
	ctx := context.Background()

	testUser := &models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	user, err := s.repo.FindByEmail(ctx, "test@example.com")
	s.Require().NoError(err)
	s.NotNil(user)
	s.Equal(testUser.ID, user.ID)
}

func (s *UserTestSuite) TestSetPasswordResetToken() {
	ctx := context.Background()

	testUser := &models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	token := "reset-token-123"
	err = s.repo.SetPasswordResetToken(ctx, testUser.Email, token)
	s.Require().NoError(err)

	var reset models.UserPasswordReset
	err = s.DB.Where("user_id = ?", testUser.ID).First(&reset).Error
	s.Require().NoError(err)
	s.Equal(token, reset.Token)
	s.Equal(testUser.Email, reset.Email)
}

func (s *UserTestSuite) TestActivateAccount() {
	ctx := context.Background()

	testUser := &models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Disabled:  true,
		CreatedAt: time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	token := "activation-token-123"
	reset := &models.UserPasswordReset{
		UserID:         testUser.ID,
		Email:          testUser.Email,
		Token:          token,
		ExpirationDate: time.Now().Add(24 * time.Hour),
	}

	err = s.DB.Create(reset).Error
	s.Require().NoError(err)

	activated, err := s.repo.ActivateAccount(ctx, testUser.Email, token)
	s.Require().NoError(err)
	s.True(activated)

	// Verify user is activated
	var updatedUser models.User
	err = s.DB.First(&updatedUser, testUser.ID).Error
	s.Require().NoError(err)
	s.False(updatedUser.Disabled)
}

func (s *UserTestSuite) TestActivateAccountInvalidToken() {
	ctx := context.Background()

	testUser := &models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Disabled:  true,
		CreatedAt: time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	activated, err := s.repo.ActivateAccount(ctx, testUser.Email, "wrong-token")
	s.Require().Error(err)
	s.False(activated)
	s.Contains(err.Error(), "invalid token")
}

func (s *UserTestSuite) TestUpdatePasswordByToken() {
	ctx := context.Background()

	testUser := &models.User{
		Email:     "test@example.com",
		Password:  "old-password",
		CreatedAt: time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	token := "reset-token-123"
	reset := &models.UserPasswordReset{
		UserID:         testUser.ID,
		Email:          testUser.Email,
		Token:          token,
		ExpirationDate: time.Now().Add(24 * time.Hour),
	}

	err = s.DB.Create(reset).Error
	s.Require().NoError(err)

	newPassword := "new-password"
	err = s.repo.UpdatePasswordByToken(ctx, testUser.Email, token, newPassword)
	s.Require().NoError(err)

	// Verify password was updated
	var updatedUser models.User
	err = s.DB.First(&updatedUser, testUser.ID).Error
	s.Require().NoError(err)
	s.Equal(newPassword, updatedUser.Password)

	// Verify the reset token was deleted
	var resetCount int64
	err = s.DB.Model(&models.UserPasswordReset{}).Where("token = ?", token).Count(&resetCount).Error
	s.Require().NoError(err)
	s.Equal(int64(0), resetCount)
}

func (s *UserTestSuite) TestUpdatePasswordByTokenInvalidToken() {
	ctx := context.Background()

	testUser := &models.User{
		Email:     "test@example.com",
		Password:  "old-password",
		CreatedAt: time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	err = s.repo.UpdatePasswordByToken(ctx, testUser.Email, "wrong-token", "new-password")
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid token")
}

func (s *UserTestSuite) TestCreateAppToken() {
	ctx := context.Background()

	testUser := &models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
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
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	// Test with user scope (not allowed)
	scopes := []models.ApiTokenScope{models.ApiTokenScopeUserRead}
	_, err = s.repo.CreateAppToken(ctx, testUser.ID, "Test Token", scopes, 30)
	s.Require().Error(err)
	s.Contains(err.Error(), "user scopes are not allowed")

	// Test with token scope (not allowed)
	scopes = []models.ApiTokenScope{models.ApiTokenScopeTokenWrite}
	_, err = s.repo.CreateAppToken(ctx, testUser.ID, "Test Token", scopes, 30)
	s.Require().Error(err)
	s.Contains(err.Error(), "token scopes are not allowed")
}

func (s *UserTestSuite) TestGetAllUserTokens() {
	ctx := context.Background()

	testUser := &models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	// Create multiple tokens
	token1 := &models.AppToken{
		UserID:    testUser.ID,
		Name:      "Token 1",
		Token:     "token1",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Scopes:    pq.StringArray{"task:read"},
	}

	token2 := &models.AppToken{
		UserID:    testUser.ID,
		Name:      "Token 2",
		Token:     "token2",
		ExpiresAt: time.Now().Add(48 * time.Hour),
		Scopes:    pq.StringArray{"task:write"},
	}

	err = s.DB.Create(token1).Error
	s.Require().NoError(err)

	err = s.DB.Create(token2).Error
	s.Require().NoError(err)

	tokens, err := s.repo.GetAllUserTokens(ctx, testUser.ID)
	s.Require().NoError(err)
	s.Require().Len(tokens, 2)

	// Should be ordered by expiration date (ascending)
	s.Equal(token1.ID, tokens[0].ID)
	s.Equal(token2.ID, tokens[1].ID)
}

func (s *UserTestSuite) TestDeleteAppToken() {
	ctx := context.Background()

	testUser := &models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	// Create a token
	token := &models.AppToken{
		UserID:    testUser.ID,
		Name:      "Test Token",
		Token:     "token123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Scopes:    pq.StringArray{"task:read"},
	}

	err = s.DB.Create(token).Error
	s.Require().NoError(err)

	// Delete the token
	err = s.repo.DeleteAppToken(ctx, testUser.ID, fmt.Sprintf("%d", token.ID))
	s.Require().NoError(err)

	// Verify it's deleted
	var count int64
	err = s.DB.Model(&models.AppToken{}).Where("id = ?", token.ID).Count(&count).Error
	s.Require().NoError(err)
	s.Equal(int64(0), count)
}

func (s *UserTestSuite) TestUpdateNotificationSettings() {
	ctx := context.Background()

	testUser := &models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	// Create initial notification settings
	settings := &models.NotificationSettings{
		UserID: testUser.ID,
		Provider: models.NotificationProvider{
			Provider: models.NotificationProviderNone,
		},
	}

	err = s.DB.Create(settings).Error
	s.Require().NoError(err)

	// Update settings
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

	// Verify settings updated
	var updated models.NotificationSettings
	err = s.DB.Where("user_id = ?", testUser.ID).First(&updated).Error
	s.Require().NoError(err)
	s.Equal(models.NotificationProviderGotify, updated.Provider.Provider)
	s.Equal("https://example.com", updated.Provider.URL)
	s.Equal("api-token", updated.Provider.Token)
	s.True(updated.Triggers.DueDate)
	s.True(updated.Triggers.PreDue)
}

func (s *UserTestSuite) TestUpdatePasswordByUserId() {
	ctx := context.Background()

	testUser := &models.User{
		Email:     "test@example.com",
		Password:  "old-password",
		CreatedAt: time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	// Update password
	newPassword := "new-password"
	err = s.repo.UpdatePasswordByUserId(ctx, testUser.ID, newPassword)
	s.Require().NoError(err)

	// Verify password updated
	var updated models.User
	err = s.DB.First(&updated, testUser.ID).Error
	s.Require().NoError(err)
	s.Equal(newPassword, updated.Password)
}

func (s *UserTestSuite) TestDeleteStalePasswordResets() {
	ctx := context.Background()

	testUser := &models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	anotherUser := &models.User{
		Email:     "test2@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	err = s.DB.Create(anotherUser).Error
	s.Require().NoError(err)

	// Create valid reset token
	validReset := &models.UserPasswordReset{
		UserID:         testUser.ID,
		Email:          testUser.Email,
		Token:          "valid-token",
		ExpirationDate: time.Now().Add(24 * time.Hour), // Future
	}

	err = s.DB.Create(validReset).Error
	s.Require().NoError(err)

	// Create expired reset token
	expiredReset := &models.UserPasswordReset{
		UserID:         anotherUser.ID,
		Email:          anotherUser.Email,
		Token:          "expired-token",
		ExpirationDate: time.Now().Add(-24 * time.Hour), // Past
	}

	err = s.DB.Create(expiredReset).Error
	s.Require().NoError(err)

	// Delete stale resets
	err = s.repo.DeleteStalePasswordResets(ctx)
	s.Require().NoError(err)

	// Verify only expired token was deleted
	var count int64
	err = s.DB.Model(&models.UserPasswordReset{}).Count(&count).Error
	s.Require().NoError(err)
	s.Equal(int64(1), count)

	var remaining models.UserPasswordReset
	err = s.DB.First(&remaining).Error
	s.Require().NoError(err)
	s.Equal("valid-token", remaining.Token)
}

func (s *UserTestSuite) TestGetAppTokensNearingExpiration() {
	ctx := context.Background()

	testUser := &models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	now := time.Now()

	// Create tokens with different expiration times
	token1 := &models.AppToken{
		UserID:    testUser.ID,
		Name:      "Expired Token",
		Token:     "token1",
		ExpiresAt: now.Add(-48 * time.Hour), // Already expired
	}

	token2 := &models.AppToken{
		UserID:    testUser.ID,
		Name:      "Nearing Expiration",
		Token:     "token2",
		ExpiresAt: now.Add(6 * time.Hour), // Will expire soon
	}

	token3 := &models.AppToken{
		UserID:    testUser.ID,
		Name:      "Valid Token",
		Token:     "token3",
		ExpiresAt: now.Add(30 * 24 * time.Hour), // Not close to expiration
	}

	err = s.DB.Create([]*models.AppToken{token1, token2, token3}).Error
	s.Require().NoError(err)

	// Associate user with token
	err = s.DB.Exec("UPDATE app_tokens SET user_id = ?", testUser.ID).Error
	s.Require().NoError(err)

	// Test with 12 hour window
	tokens, err := s.repo.GetAppTokensNearingExpiration(ctx, 12*time.Hour)
	s.Require().NoError(err)
	s.Require().Len(tokens, 1)
	s.Equal("Nearing Expiration", tokens[0].Name)
}

func (s *UserTestSuite) TestDeleteStaleAppTokens() {
	ctx := context.Background()

	testUser := &models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	err := s.DB.Create(testUser).Error
	s.Require().NoError(err)

	now := time.Now()

	// Create tokens with different expiration times
	token1 := &models.AppToken{
		UserID:    testUser.ID,
		Name:      "Expired Token",
		Token:     "token1",
		ExpiresAt: now.Add(-48 * time.Hour), // Already expired
	}

	token2 := &models.AppToken{
		UserID:    testUser.ID,
		Name:      "Valid Token",
		Token:     "token2",
		ExpiresAt: now.Add(24 * time.Hour), // Still valid
	}

	err = s.DB.Create([]*models.AppToken{token1, token2}).Error
	s.Require().NoError(err)

	// Delete stale tokens
	err = s.repo.DeleteStaleAppTokens(ctx)
	s.Require().NoError(err)

	// Verify only expired token was deleted
	var count int64
	err = s.DB.Model(&models.AppToken{}).Count(&count).Error
	s.Require().NoError(err)
	s.Equal(int64(1), count)

	var remaining models.AppToken
	err = s.DB.First(&remaining).Error
	s.Require().NoError(err)
	s.Equal("Valid Token", remaining.Name)
}
