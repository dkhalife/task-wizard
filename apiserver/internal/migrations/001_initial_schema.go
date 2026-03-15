package migrations

import (
	"context"
	"fmt"

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
	dialect := db.Dialector.Name()

	var (
		autoInc     string
		strCol      string
		idxIne      string
		tokenIdxCol string
	)

	switch dialect {
	case "sqlite":
		autoInc = "AUTOINCREMENT"
		strCol = "TEXT"
		idxIne = "IF NOT EXISTS "
		tokenIdxCol = "token"
	case "mysql":
		autoInc = "AUTO_INCREMENT"
		strCol = "VARCHAR(255)"
		idxIne = ""
		tokenIdxCol = "token(255)"
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	statements := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY %s,
			display_name %s NOT NULL DEFAULT '',
			email %s NOT NULL DEFAULT '' UNIQUE,
			password %s NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT NULL,
			disabled BOOLEAN DEFAULT false
		)`, autoInc, strCol, strCol, strCol),
		`CREATE TABLE IF NOT EXISTS user_password_resets (
			user_id INTEGER NOT NULL PRIMARY KEY,
			email TEXT NOT NULL,
			token TEXT NOT NULL,
			expiration_date DATETIME NOT NULL
		)`,
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS app_tokens (
			id INTEGER PRIMARY KEY %s,
			user_id INTEGER NOT NULL,
			name %s NOT NULL,
			token %s NOT NULL,
			scopes TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_users_app_tokens FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`, autoInc, strCol, strCol),
		fmt.Sprintf(`CREATE INDEX %sidx_app_tokens_token ON app_tokens(%s)`, idxIne, tokenIdxCol),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS labels (
			id INTEGER PRIMARY KEY %s,
			name %s NOT NULL,
			color VARCHAR(7) NOT NULL,
			created_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT NULL,
			CONSTRAINT fk_users_labels FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
		)`, autoInc, strCol),
		fmt.Sprintf(`CREATE INDEX %sidx_labels_created_by ON labels(created_by)`, idxIne),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY %s,
			title TEXT NOT NULL,
			frequency_type VARCHAR(9),
			frequency_on VARCHAR(18) DEFAULT NULL,
			frequency_every INTEGER DEFAULT NULL,
			frequency_unit VARCHAR(9) DEFAULT NULL,
			frequency_days TEXT,
			frequency_months TEXT,
			next_due_date DATETIME,
			end_date DATETIME DEFAULT NULL,
			is_rolling BOOLEAN DEFAULT false,
			created_by INTEGER NOT NULL,
			is_active BOOLEAN DEFAULT true,
			notification_enabled BOOLEAN DEFAULT false,
			notification_due_date BOOLEAN DEFAULT false,
			notification_pre_due BOOLEAN DEFAULT false,
			notification_overdue BOOLEAN DEFAULT false,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT NULL,
			CONSTRAINT fk_users_tasks FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
		)`, autoInc),
		fmt.Sprintf(`CREATE INDEX %sidx_tasks_next_due_date ON tasks(next_due_date)`, idxIne),
		fmt.Sprintf(`CREATE INDEX %sidx_tasks_created_by ON tasks(created_by)`, idxIne),
		fmt.Sprintf(`CREATE INDEX %sidx_tasks_is_active ON tasks(is_active)`, idxIne),
		`CREATE TABLE IF NOT EXISTS task_labels (
			task_id INTEGER NOT NULL,
			label_id INTEGER NOT NULL,
			PRIMARY KEY (task_id, label_id),
			CONSTRAINT fk_task_labels_task FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
			CONSTRAINT fk_task_labels_label FOREIGN KEY (label_id) REFERENCES labels(id) ON DELETE CASCADE
		)`,
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS task_histories (
			id INTEGER PRIMARY KEY %s,
			task_id INTEGER NOT NULL,
			completed_date DATETIME,
			due_date DATETIME,
			CONSTRAINT fk_tasks_history FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
		)`, autoInc),
		fmt.Sprintf(`CREATE INDEX %sidx_task_histories_task_id ON task_histories(task_id)`, idxIne),
		`CREATE TABLE IF NOT EXISTS notification_settings (
			user_id INTEGER NOT NULL,
			notifications_provider_type VARCHAR(7),
			notifications_provider_url TEXT,
			notifications_provider_method VARCHAR(4),
			notifications_provider_token TEXT,
			notifications_triggers_enabled BOOLEAN DEFAULT false,
			notifications_triggers_due_date BOOLEAN DEFAULT false,
			notifications_triggers_pre_due BOOLEAN DEFAULT false,
			notifications_triggers_overdue BOOLEAN DEFAULT false,
			CONSTRAINT fk_users_notification_settings FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS notifications (
			id INTEGER PRIMARY KEY %s,
			task_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			text TEXT NOT NULL,
			type VARCHAR(8) NOT NULL,
			is_sent BOOLEAN DEFAULT false,
			scheduled_for DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_tasks_notifications FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
			CONSTRAINT fk_users_notifications FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`, autoInc),
		fmt.Sprintf(`CREATE INDEX %sidx_notifications_task_id ON notifications(task_id)`, idxIne),
		fmt.Sprintf(`CREATE INDEX %sidx_notifications_user_id ON notifications(user_id)`, idxIne),
		fmt.Sprintf(`CREATE INDEX %sidx_notifications_is_sent ON notifications(is_sent)`, idxIne),
		fmt.Sprintf(`CREATE INDEX %sidx_notifications_scheduled_for ON notifications(scheduled_for)`, idxIne),
	}

	for _, stmt := range statements {
		if err := db.WithContext(ctx).Exec(stmt).Error; err != nil {
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
		if err := db.WithContext(ctx).Exec("DROP TABLE IF EXISTS " + table).Error; err != nil {
			return err
		}
	}

	return nil
}
