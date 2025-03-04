package migration

import (
	"dkhalife.com/tasks/core/internal/models"
	"gorm.io/gorm"
)

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(
		models.User{},
		models.UserPasswordReset{},
		models.AppToken{},
		models.Label{},
		models.Task{},
		models.TaskHistory{},
		models.NotificationSettings{},
		models.Notification{},
	); err != nil {
		return err
	}

	return nil
}
