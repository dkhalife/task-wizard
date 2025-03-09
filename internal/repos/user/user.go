package repos

import (
	"context"
	"fmt"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/utils/auth"
	"github.com/golang-jwt/jwt/v4"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type UserRepository struct {
	cfg *config.Config
	db  *gorm.DB
}

func NewUserRepository(db *gorm.DB, cfg *config.Config) *UserRepository {
	return &UserRepository{cfg, db}
}

func (r *UserRepository) CreateUser(c context.Context, user *models.User) error {
	if !r.cfg.Server.Registration {
		return fmt.Errorf("new account registration is disabled")
	}

	return r.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		if err := tx.Create(&models.NotificationSettings{
			UserID: user.ID,
			Provider: models.NotificationProvider{
				Provider: models.NotificationProviderNone,
			},
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *UserRepository) GetUser(c context.Context, id int) (*models.User, error) {
	var user *models.User
	if err := r.db.WithContext(c).Where("ID = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(c context.Context, email string) (*models.User, error) {
	var user *models.User
	if err := r.db.WithContext(c).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) SetPasswordResetToken(c context.Context, email string, token string) error {
	user, err := r.FindByEmail(c, email)
	if err != nil {
		return err
	}

	if err := r.db.WithContext(c).Where("user_id = ?", user.ID).Delete(&models.UserPasswordReset{}).Error; err != nil {
		return err
	}

	if err := r.db.WithContext(c).Model(&models.UserPasswordReset{}).Create(&models.UserPasswordReset{
		UserID:         user.ID,
		Token:          token,
		Email:          email,
		ExpirationDate: time.Now().UTC().Add(r.cfg.SchedulerJobs.PasswordResetValidity),
	}).Error; err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) ActivateAccount(c context.Context, email string, code string) (bool, error) {
	user, err := r.FindByEmail(c, email)
	if err != nil {
		return false, err
	}

	if !user.Disabled {
		return false, nil
	}

	result := r.db.WithContext(c).Where("email = ?", email).Where("token = ?", code).Delete(&models.UserPasswordReset{})
	if result.RowsAffected <= 0 {
		return false, fmt.Errorf("invalid token")
	}

	err = r.db.WithContext(c).Model(&models.User{}).Where("email = ? AND disabled = 1", email).Update("disabled", false).Error
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *UserRepository) UpdatePasswordByToken(ctx context.Context, email string, token string, password string) error {
	logger := logging.FromContext(ctx)

	logger.Debugw("account.db.UpdatePasswordByToken", "email", email)
	upr := &models.UserPasswordReset{
		Email: email,
		Token: token,
	}
	result := r.db.WithContext(ctx).Where("email = ?", email).Where("token = ?", token).Delete(upr)
	if result.RowsAffected <= 0 {
		return fmt.Errorf("invalid token")
	}

	chain := r.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).UpdateColumns(map[string]interface{}{"password": password})
	if chain.Error != nil {
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		return fmt.Errorf("account not found")
	}

	return nil
}

func convertScopesToStringArray(scopes []models.ApiTokenScope) []string {
	strScopes := make([]string, len(scopes))
	for i, scope := range scopes {
		strScopes[i] = string(scope)
	}

	return pq.StringArray(strScopes)
}

func (r *UserRepository) CreateAppToken(c context.Context, userID int, name string, scopes []models.ApiTokenScope, days int) (*models.AppToken, error) {
	duration := time.Duration(days) * 24 * time.Hour
	expiresAt := time.Now().UTC().Add(duration)
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		auth.IdentityKey: fmt.Sprintf("%d", userID),
		"exp":            expiresAt,
		"type":           "app",
		"scopes":         scopes,
	})

	signedToken, err := jwtToken.SignedString([]byte(r.cfg.Jwt.Secret))
	if err != nil {
		logging.FromContext(c).Errorw("failed to sign token", "err", err)
		return nil, err
	}

	for _, scope := range scopes {
		if scope == models.ApiTokenScopeUserRead || scope == models.ApiTokenScopeUserWrite {
			return nil, fmt.Errorf("user scopes are not allowed")
		}

		if scope == models.ApiTokenScopeTokenWrite {
			return nil, fmt.Errorf("token scopes are not allowed")
		}
	}

	token := &models.AppToken{
		UserID:    userID,
		Name:      name,
		Token:     signedToken,
		ExpiresAt: expiresAt,
		Scopes:    convertScopesToStringArray(scopes),
	}

	if err := r.db.WithContext(c).Create(token).Error; err != nil {
		logging.FromContext(c).Errorw("failed to save", "err", err)
		return nil, err
	}

	return token, nil
}

func (r *UserRepository) GetAllUserTokens(c context.Context, userID int) ([]*models.AppToken, error) {
	var tokens []*models.AppToken
	if err := r.db.WithContext(c).
		Where("user_id = ?", userID).
		Order("expires_at ASC").
		Select("id, name, scopes, expires_at").
		Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *UserRepository) DeleteAppToken(c context.Context, userID int, tokenID string) error {
	return r.db.WithContext(c).Where("id = ? AND user_id = ?", tokenID, userID).Delete(&models.AppToken{}).Error
}

func (r *UserRepository) UpdateNotificationSettings(c context.Context, userID int, provider models.NotificationProvider, triggers models.NotificationTriggerOptions) error {
	return r.db.WithContext(c).Where("user_id = ?", userID).Updates(&models.NotificationSettings{
		Provider: provider,
		Triggers: triggers,
	}).Error
}

func (r *UserRepository) DeleteNotificationsForUser(c context.Context, userID int) error {
	return r.db.WithContext(c).Where("user_id = ?", userID).Delete(&models.NotificationSettings{}).Error
}

func (r *UserRepository) UpdatePasswordByUserId(c context.Context, userID int, password string) error {
	return r.db.WithContext(c).Model(&models.User{}).Where("id = ?", userID).Update("password", password).Error
}

func (r *UserRepository) DeleteStalePasswordResets(c context.Context) error {
	now := time.Now().UTC()
	return r.db.WithContext(c).Where("expiration_date <= ?", now).Delete(&models.UserPasswordReset{}).Error
}

func (r *UserRepository) GetAppTokensNearingExpiration(c context.Context, before time.Duration) ([]*models.AppToken, error) {
	lowerBound := time.Now().UTC().Add(-before)
	var tokens []*models.AppToken
	if err := r.db.WithContext(c).
		Where("expires_at > ?", lowerBound).
		Where("expires_at <= ?", lowerBound.Add(before)).
		Preload("User").
		Find(&tokens).Error; err != nil {
		return nil, err
	}

	return tokens, nil
}

func (r *UserRepository) DeleteStaleAppTokens(c context.Context) error {
	now := time.Now().UTC()
	return r.db.WithContext(c).Where("expires_at <= ?", now).Delete(&models.AppToken{}).Error
}
