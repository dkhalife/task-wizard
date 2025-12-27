package migrations

import (
	"context"

	"dkhalife.com/tasks/core/internal/models"
	"gorm.io/gorm"
)

func init() {
	Register(&InitialSchemaMigration{})
}

type InitialSchemaMigration struct{}

func (m *InitialSchemaMigration) Version() int {
	return 1
}

func (m *InitialSchemaMigration) Name() string {
	return "initial_schema"
}

func (m *InitialSchemaMigration) Up(ctx context.Context, db *gorm.DB) error {
	tables := []interface{}{
		&models.User{},
		&models.UserPasswordReset{},
		&models.AppToken{},
		&models.Label{},
		&models.Task{},
		&models.TaskLabel{},
		&models.TaskHistory{},
		&models.NotificationSettings{},
		&models.Notification{},
	}

	for _, table := range tables {
		if err := db.WithContext(ctx).AutoMigrate(table); err != nil {
			return err
		}
	}

	return nil
}

func (m *InitialSchemaMigration) Down(ctx context.Context, db *gorm.DB) error {
	tables := []string{
		"notifications",
		"notification_settings",
		"task_histories",
		"task_labels",
		"tasks",
		"labels",
		"app_tokens",
		"user_password_resets",
		"users",
	}

	for _, table := range tables {
		if err := db.WithContext(ctx).Exec("DROP TABLE IF EXISTS " + table + " CASCADE").Error; err != nil {
			return err
		}
	}

	return nil
}
