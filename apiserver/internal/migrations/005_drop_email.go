package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func init() {
	Register(&DropEmailMigration{})
}

type DropEmailMigration struct{}

func (m *DropEmailMigration) Version() int {
	return 5
}

func (m *DropEmailMigration) Name() string {
	return "drop_email"
}

func (m *DropEmailMigration) Up(ctx context.Context, db *gorm.DB) error {
	dbCtx := db.WithContext(ctx)
	dialect := db.Dialector.Name()

	switch dialect {
	case "sqlite":
		stmts := []string{
			`CREATE TABLE users_new (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				display_name TEXT NOT NULL DEFAULT '',
				directory_id TEXT NOT NULL DEFAULT '',
				object_id TEXT NOT NULL DEFAULT '',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT NULL,
				disabled BOOLEAN DEFAULT false
			)`,
			`INSERT INTO users_new (id, display_name, directory_id, object_id, created_at, updated_at, disabled)
				SELECT id, display_name, directory_id, object_id, created_at, updated_at, disabled FROM users`,
			`DROP TABLE users`,
			`ALTER TABLE users_new RENAME TO users`,
		}
		for _, stmt := range stmts {
			if err := dbCtx.Exec(stmt).Error; err != nil {
				return err
			}
		}
	case "mysql":
		if err := dbCtx.Exec("ALTER TABLE users DROP COLUMN email").Error; err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	return nil
}

func (m *DropEmailMigration) Down(ctx context.Context, db *gorm.DB) error {
	dbCtx := db.WithContext(ctx)
	dialect := db.Dialector.Name()

	switch dialect {
	case "sqlite":
		if err := dbCtx.Exec("ALTER TABLE users ADD COLUMN email TEXT NOT NULL DEFAULT ''").Error; err != nil {
			return err
		}
	case "mysql":
		if err := dbCtx.Exec("ALTER TABLE users ADD COLUMN email VARCHAR(255) NOT NULL DEFAULT ''").Error; err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	return nil
}
