package repos

import (
	"context"
	"fmt"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/utils/auth"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

type IUserRepo interface {
	CreateUser(c context.Context, user *models.User) error
	GetUser(c context.Context, id int) (*models.User, error)
	FindByEmail(c context.Context, email string) (*models.User, error)
	CreateAppToken(c context.Context, userID int, name string, scopes []models.ApiTokenScope, days int) (*models.AppToken, error)
	GetAppTokenByID(c context.Context, tokenId int) (*models.AppToken, error)
	GetAllUserTokens(c context.Context, userID int) ([]*models.AppToken, error)
	DeleteAppToken(c context.Context, userID int, tokenID int) error
	UpdateNotificationSettings(c context.Context, userID int, provider models.NotificationProvider, triggers models.NotificationTriggerOptions) error
	DeleteNotificationsForUser(c context.Context, userID int) error
	GetAppTokensNearingExpiration(c context.Context, before time.Duration) ([]*models.AppToken, error)
	DeleteStaleAppTokens(c context.Context) error
	GetLastCreatedOrModifiedForUserResources(c context.Context, userID int) (string, error)
}

type UserRepository struct {
	cfg *config.Config
	db  *gorm.DB
}

var _ IUserRepo = (*UserRepository)(nil)

func NewUserRepository(db *gorm.DB, cfg *config.Config) IUserRepo {
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

func (r *UserRepository) CreateAppToken(c context.Context, userID int, name string, scopes []models.ApiTokenScope, days int) (*models.AppToken, error) {
	var token *models.AppToken
	err := r.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		var nextID int
		if err := tx.Raw("SELECT COALESCE(MAX(id), 0)+1 AS next_id FROM app_tokens").Scan(&nextID).Error; err != nil {
			return fmt.Errorf("failed to get next token id: %w", err)
		}

		for _, scope := range scopes {
			if scope == models.ApiTokenScopeUserRead || scope == models.ApiTokenScopeUserWrite {
				return fmt.Errorf("user scopes are not allowed")
			}

			if scope == models.ApiTokenScopeTokensWrite {
				return fmt.Errorf("token scopes are not allowed")
			}
		}

		duration := time.Duration(days) * 24 * time.Hour
		expiresAt := time.Now().UTC().Add(duration)
		jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			auth.AppTokenKey: fmt.Sprintf("%d", nextID),
			auth.IdentityKey: fmt.Sprintf("%d", userID),
			"exp":            expiresAt.Unix(),
			"type":           "app",
			"scopes":         scopes,
		})

		signedToken, err := jwtToken.SignedString([]byte(r.cfg.Jwt.Secret))
		if err != nil {
			return fmt.Errorf("failed to sign token: %s", err.Error())
		}

		token = &models.AppToken{
			ID:        nextID,
			UserID:    userID,
			Name:      name,
			Token:     signedToken,
			ExpiresAt: expiresAt,
			Scopes:    auth.ConvertScopesToStringArray(scopes),
		}

		if err := tx.Create(token).Error; err != nil {
			return fmt.Errorf("failed to save token: %s", err.Error())
		}
		return nil
	})

	return token, err
}

func (r *UserRepository) GetAppTokenByID(c context.Context, tokenId int) (*models.AppToken, error) {
	var token models.AppToken
	if err := r.db.WithContext(c).Where("id = ?", tokenId).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
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

func (r *UserRepository) DeleteAppToken(c context.Context, userID int, tokenID int) error {
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

func (r *UserRepository) GetAppTokensNearingExpiration(c context.Context, before time.Duration) ([]*models.AppToken, error) {
	lowerBound := time.Now().UTC().Add(-before)
	var tokens []*models.AppToken
	if err := r.db.WithContext(c).
		Where("expires_at > ?", lowerBound).
		Where("expires_at <= ?", lowerBound.Add(2*before)).
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

func (r *UserRepository) GetLastCreatedOrModifiedForUserResources(c context.Context, userID int) (string, error) {
	var result string
	err := r.db.WithContext(c).Raw(`
		SELECT 
			MAX(
				COALESCE(MAX(updated_at), '1970-01-01 00:00:00'),
				COALESCE(MAX(created_at), '1970-01-01 00:00:00')
			) AS last_modified
		FROM (
			SELECT updated_at, created_at FROM labels WHERE created_by = ?
			UNION ALL
			SELECT updated_at, created_at FROM tasks WHERE created_by = ?
		) AS combined_dates
	`, userID, userID).Scan(&result).Error

	return result, err
}
