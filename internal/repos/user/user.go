package repos

import (
	"context"
	"fmt"
	"time"

	"dkhalife.com/tasks/core/config"
	nModel "dkhalife.com/tasks/core/internal/models/notifier"
	uModel "dkhalife.com/tasks/core/internal/models/user"
	"dkhalife.com/tasks/core/internal/services/logging"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB, cfg *config.Config) *UserRepository {
	return &UserRepository{db}
}

func (r *UserRepository) CreateUser(c context.Context, user *uModel.User) error {
	if err := r.db.WithContext(c).Save(user).Error; err != nil {
		return err
	}

	return r.db.WithContext(c).Save(&nModel.NotificationSettings{
		UserID: user.ID,
		Provider: nModel.NotificationProvider{
			Provider: nModel.NotificationProviderNone,
		},
	}).Error
}

func (r *UserRepository) GetUser(c context.Context, id int) (*uModel.User, error) {
	var user *uModel.User
	if err := r.db.WithContext(c).Where("ID = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(c context.Context, email string) (*uModel.User, error) {
	var user *uModel.User
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

	if err := r.db.WithContext(c).Model(&uModel.UserPasswordReset{}).Save(&uModel.UserPasswordReset{
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
	upr := &uModel.UserPasswordReset{
		Email: email,
		Token: token,
	}
	result := r.db.WithContext(ctx).Where("email = ?", email).Where("token = ?", token).Delete(upr)
	if result.RowsAffected <= 0 {
		return fmt.Errorf("invalid token")
	}

	chain := r.db.WithContext(ctx).Model(&uModel.User{}).Where("email = ?", email).UpdateColumns(map[string]interface{}{"password": password})
	if chain.Error != nil {
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		return fmt.Errorf("account not found")
	}

	return nil
}

func (r *UserRepository) StoreAPIToken(c context.Context, userID int, name string, tokenCode string) (*uModel.APIToken, error) {
	token := &uModel.APIToken{
		UserID:    userID,
		Name:      name,
		Token:     tokenCode,
		CreatedAt: time.Now().UTC(),
	}
	if err := r.db.WithContext(c).Model(&uModel.APIToken{}).Save(
		token).Error; err != nil {
		return nil, err

	}
	return token, nil
}

func (r *UserRepository) GetAllUserTokens(c context.Context, userID int) ([]*uModel.APIToken, error) {
	var tokens []*uModel.APIToken
	if err := r.db.WithContext(c).Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *UserRepository) DeleteAPIToken(c context.Context, userID int, tokenID string) error {
	return r.db.WithContext(c).Where("id = ? AND user_id = ?", tokenID, userID).Delete(&uModel.APIToken{}).Error
}

func (r *UserRepository) UpdateNotificationSettings(c context.Context, userID int, provider nModel.NotificationProvider, triggers nModel.NotificationTriggerOptions) error {
	return r.db.WithContext(c).Save(&nModel.NotificationSettings{
		UserID:   userID,
		Provider: provider,
		Triggers: triggers,
	}).Error
}

func (r *UserRepository) UpdatePasswordByUserId(c context.Context, userID int, password string) error {
	return r.db.WithContext(c).Model(&uModel.User{}).Where("id = ?", userID).Update("password", password).Error
}
