package repos

import (
	"context"
	"errors"
	"fmt"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/models"
	"gorm.io/gorm"
)

type IUserRepo interface {
	CreateUser(c context.Context, user *models.User) error
	GetUser(c context.Context, id int) (*models.User, error)
	FindByEntraID(c context.Context, directoryID string, objectID string) (*models.User, error)
	EnsureUser(c context.Context, directoryID string, objectID string) (*models.User, error)
	UpdateNotificationSettings(c context.Context, userID int, provider models.NotificationProvider, triggers models.NotificationTriggerOptions) error
	DeleteNotificationsForUser(c context.Context, userID int) error
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

func (r *UserRepository) FindByEntraID(c context.Context, directoryID string, objectID string) (*models.User, error) {
	var user *models.User
	if err := r.db.WithContext(c).Where("directory_id = ? AND object_id = ?", directoryID, objectID).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) EnsureUser(c context.Context, directoryID string, objectID string) (*models.User, error) {
	user, err := r.FindByEntraID(c, directoryID, objectID)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find user by Entra ID: %s", err.Error())
	}

	newUser := &models.User{
		DirectoryID: directoryID,
		ObjectID:    objectID,
	}

	if err := r.CreateUser(c, newUser); err != nil {
		return nil, fmt.Errorf("create user: %s", err.Error())
	}

	return newUser, nil
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
