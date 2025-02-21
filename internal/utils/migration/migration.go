package migration

import (
	lModel "dkhalife.com/tasks/core/internal/models/label"
	nModel "dkhalife.com/tasks/core/internal/models/notifier"
	tModel "dkhalife.com/tasks/core/internal/models/task"
	uModel "dkhalife.com/tasks/core/internal/models/user"
	"gorm.io/gorm"
)

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(
		uModel.User{},
		uModel.UserPasswordReset{},
		uModel.APIToken{},
		lModel.Label{},
		tModel.Task{},
		tModel.TaskHistory{},
		nModel.NotificationSettings{},
		nModel.Notification{},
	); err != nil {
		return err
	}

	return nil
}
