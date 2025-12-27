package migration

import (
	"dkhalife.com/tasks/core/internal/models"
	"gorm.io/gorm"
)

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(
		models.User{},
		models.UserPasswordReset{},
		models.AppToken{},
		models.Label{},
		models.Task{},
		models.TaskLabel{},
		models.TaskHistory{},
		models.NotificationSettings{},
		models.Notification{},
	)
}
