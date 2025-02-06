package migration

import (
	lModel "donetick.com/core/internal/models/label"
	nModel "donetick.com/core/internal/models/notifier"
	tModel "donetick.com/core/internal/models/task"
	uModel "donetick.com/core/internal/models/user"
	"gorm.io/gorm"
)

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(
		uModel.User{},
		uModel.UserPasswordReset{},
		uModel.APIToken{},
		lModel.Label{},
		tModel.Task{},
		tModel.TaskOccurrence{},
		tModel.TaskLabels{},
		tModel.TaskHistory{},
		nModel.Notification{},
	); err != nil {
		return err
	}

	return nil
}
