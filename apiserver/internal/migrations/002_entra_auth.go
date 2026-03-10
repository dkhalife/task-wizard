package migrations

import (
	"context"

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
	type columnInfo struct {
		Name string `gorm:"column:name"`
	}

	var columns []columnInfo
	if err := db.WithContext(ctx).Raw("PRAGMA table_info(users)").Scan(&columns).Error; err != nil {
		return err
	}

	hasDirectoryID := false
	hasObjectID := false
	for _, col := range columns {
		if col.Name == "directory_id" {
			hasDirectoryID = true
		}
		if col.Name == "object_id" {
			hasObjectID = true
		}
	}

	if !hasDirectoryID {
		if err := db.WithContext(ctx).Exec("ALTER TABLE users ADD COLUMN directory_id TEXT NOT NULL DEFAULT ''").Error; err != nil {
			return err
		}
	}

	if !hasObjectID {
		if err := db.WithContext(ctx).Exec("ALTER TABLE users ADD COLUMN object_id TEXT NOT NULL DEFAULT ''").Error; err != nil {
			return err
		}
	}

	if err := db.WithContext(ctx).Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_users_entra_id ON users(directory_id, object_id) WHERE directory_id != '' AND object_id != ''").Error; err != nil {
		return err
	}

	if err := db.WithContext(ctx).Exec("DROP TABLE IF EXISTS user_password_resets").Error; err != nil {
		return err
	}

	return nil
}

func (m *EntraAuthMigration) Down(ctx context.Context, db *gorm.DB) error {
	if err := db.WithContext(ctx).Exec("DROP INDEX IF EXISTS idx_users_entra_id").Error; err != nil {
		return err
	}

	if err := db.WithContext(ctx).Exec(`CREATE TABLE IF NOT EXISTS user_password_resets (
		user_id INTEGER NOT NULL PRIMARY KEY,
		email TEXT NOT NULL,
		token TEXT NOT NULL,
		expiration_date DATETIME NOT NULL
	)`).Error; err != nil {
		return err
	}

	return nil
}
