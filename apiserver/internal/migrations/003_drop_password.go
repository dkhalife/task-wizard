package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func init() {
	Register(&DropPasswordMigration{})
}

type DropPasswordMigration struct{}

func (m *DropPasswordMigration) Version() int {
	return 3
}

func (m *DropPasswordMigration) Name() string {
	return "drop_password"
}

func (m *DropPasswordMigration) Up(ctx context.Context, db *gorm.DB) error {
	dbCtx := db.WithContext(ctx)
	migrator := dbCtx.Migrator()

	if migrator.HasColumn("users", "password") {
		if err := dbCtx.Exec("ALTER TABLE users DROP COLUMN password").Error; err != nil {
			return err
		}
	}

	return nil
}

func (m *DropPasswordMigration) Down(ctx context.Context, db *gorm.DB) error {
	dbCtx := db.WithContext(ctx)
	migrator := dbCtx.Migrator()
	dialect := db.Dialector.Name()

	if !migrator.HasColumn("users", "password") {
		var colType string
		switch dialect {
		case "sqlite":
			colType = "TEXT"
		case "mysql":
			colType = "VARCHAR(255)"
		default:
			return fmt.Errorf("unsupported dialect: %s", dialect)
		}

		if err := dbCtx.Exec(fmt.Sprintf("ALTER TABLE users ADD COLUMN password %s NOT NULL DEFAULT ''", colType)).Error; err != nil {
			return err
		}
	}

	return nil
}
