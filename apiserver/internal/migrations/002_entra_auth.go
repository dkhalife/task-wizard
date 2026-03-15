package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func init() {
	Register(&EntraAuthMigration{})
}

type EntraAuthMigration struct{}

func (m *EntraAuthMigration) Version() int {
	return 2
}

func (m *EntraAuthMigration) Name() string {
	return "entra_auth"
}

func (m *EntraAuthMigration) Up(ctx context.Context, db *gorm.DB) error {
	dbCtx := db.WithContext(ctx)
	migrator := dbCtx.Migrator()
	dialect := db.Dialector.Name()

	var colType string
	switch dialect {
	case "sqlite":
		colType = "TEXT"
	case "mysql":
		colType = "VARCHAR(255)"
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	if !migrator.HasColumn("users", "directory_id") {
		if err := dbCtx.Exec(fmt.Sprintf("ALTER TABLE users ADD COLUMN directory_id %s NOT NULL DEFAULT ''", colType)).Error; err != nil {
			return err
		}
	}

	if !migrator.HasColumn("users", "object_id") {
		if err := dbCtx.Exec(fmt.Sprintf("ALTER TABLE users ADD COLUMN object_id %s NOT NULL DEFAULT ''", colType)).Error; err != nil {
			return err
		}
	}

	if !migrator.HasIndex("users", "idx_users_entra_id") {
		switch dialect {
		case "sqlite":
			if err := dbCtx.Exec("CREATE UNIQUE INDEX idx_users_entra_id ON users(directory_id, object_id) WHERE directory_id != '' AND object_id != ''").Error; err != nil {
				return err
			}
		case "mysql":
			for _, stmt := range []string{
				"ALTER TABLE users ADD COLUMN directory_id_idx VARCHAR(255) GENERATED ALWAYS AS (IF(directory_id = '', NULL, directory_id)) VIRTUAL",
				"ALTER TABLE users ADD COLUMN object_id_idx VARCHAR(255) GENERATED ALWAYS AS (IF(object_id = '', NULL, object_id)) VIRTUAL",
				"CREATE UNIQUE INDEX idx_users_entra_id ON users (directory_id_idx, object_id_idx)",
			} {
				if err := dbCtx.Exec(stmt).Error; err != nil {
					return err
				}
			}
		}
	}

	if err := dbCtx.Exec("DROP TABLE IF EXISTS user_password_resets").Error; err != nil {
		return err
	}

	return nil
}

func (m *EntraAuthMigration) Down(ctx context.Context, db *gorm.DB) error {
	dbCtx := db.WithContext(ctx)
	migrator := dbCtx.Migrator()
	dialect := db.Dialector.Name()

	if migrator.HasIndex("users", "idx_users_entra_id") {
		if err := migrator.DropIndex("users", "idx_users_entra_id"); err != nil {
			return err
		}
	}

	if dialect == "mysql" {
		for _, col := range []string{"directory_id_idx", "object_id_idx"} {
			if migrator.HasColumn("users", col) {
				if err := dbCtx.Exec("ALTER TABLE users DROP COLUMN " + col).Error; err != nil {
					return err
				}
			}
		}
	}

	if err := dbCtx.Exec(`CREATE TABLE IF NOT EXISTS user_password_resets (
		user_id INTEGER NOT NULL PRIMARY KEY,
		email TEXT NOT NULL,
		token TEXT NOT NULL,
		expiration_date DATETIME NOT NULL
	)`).Error; err != nil {
		return err
	}

	return nil
}
