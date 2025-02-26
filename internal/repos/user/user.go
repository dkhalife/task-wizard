package repos

import (
	"context"
	"fmt"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/services/logging"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB, cfg *config.Config) *UserRepository {
	return &UserRepository{db}
}

func (r *UserRepository) CreateUser(c context.Context, user *models.User) error {
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

func (r *UserRepository) SetPasswordResetToken(c context.Context, email, token string) error {
	user, err := r.FindByEmail(c, email)
	if err != nil {
		return err
	}

	if err := r.db.WithContext(c).Model(&models.UserPasswordReset{}).Save(&models.UserPasswordReset{
		UserID:         user.ID,
		Token:          token,
		Email:          email,
		ExpirationDate: time.Now().UTC().Add(time.Hour * 24),
	}).Error; err != nil {
		return err
	}

	return nil
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

func (r *UserRepository) StoreAPIToken(c context.Context, userID int, name string, tokenCode string) (*models.APIToken, error) {
	token := &models.APIToken{
		UserID: userID,
		Name:   name,
		Token:  tokenCode,
	}
	if err := r.db.WithContext(c).Model(&models.APIToken{}).Save(
		token).Error; err != nil {
		return nil, err

	}
	return token, nil
}

func (r *UserRepository) GetAllUserTokens(c context.Context, userID int) ([]*models.APIToken, error) {
	var tokens []*models.APIToken
	if err := r.db.WithContext(c).Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *UserRepository) DeleteAPIToken(c context.Context, userID int, tokenID string) error {
	return r.db.WithContext(c).Where("id = ? AND user_id = ?", tokenID, userID).Delete(&models.APIToken{}).Error
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
