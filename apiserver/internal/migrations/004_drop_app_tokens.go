package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func init() {
	Register(&DropAppTokensMigration{})
}

type DropAppTokensMigration struct{}

func (m *DropAppTokensMigration) Version() int {
	return 4
}

func (m *DropAppTokensMigration) Name() string {
	return "drop_app_tokens"
}

func (m *DropAppTokensMigration) Up(ctx context.Context, db *gorm.DB) error {
	dbCtx := db.WithContext(ctx)
	migrator := dbCtx.Migrator()

	if migrator.HasTable("app_tokens") {
		if err := dbCtx.Exec("DROP TABLE app_tokens").Error; err != nil {
			return err
		}
	}

	return nil
}

func (m *DropAppTokensMigration) Down(ctx context.Context, db *gorm.DB) error {
	dbCtx := db.WithContext(ctx)
	migrator := dbCtx.Migrator()
	dialect := db.Dialector.Name()

	if !migrator.HasTable("app_tokens") {
		switch dialect {
		case "sqlite":
			if err := dbCtx.Exec(`CREATE TABLE app_tokens (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				name TEXT NOT NULL,
				token TEXT NOT NULL,
				scopes TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				expires_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			)`).Error; err != nil {
				return err
			}
		case "mysql":
			if err := dbCtx.Exec(`CREATE TABLE app_tokens (
				id INT AUTO_INCREMENT PRIMARY KEY,
				user_id INT NOT NULL,
				name VARCHAR(255) NOT NULL,
				token TEXT NOT NULL,
				scopes TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				expires_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			)`).Error; err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported dialect: %s", dialect)
		}

		if err := dbCtx.Exec("CREATE INDEX idx_app_tokens_token ON app_tokens(token(255))").Error; err != nil {
			return err
		}
	}

	return nil
}
