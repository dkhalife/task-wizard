package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func init() {
	Register(&SessionsMigration{})
}

type SessionsMigration struct{}

func (m *SessionsMigration) Version() int {
	return 7
}

func (m *SessionsMigration) Name() string {
	return "sessions"
}

func (m *SessionsMigration) Up(ctx context.Context, db *gorm.DB) error {
	dbCtx := db.WithContext(ctx)
	dialect := db.Dialector.Name()

	switch dialect {
	case "sqlite":
		stmts := []string{
			`CREATE TABLE sessions (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				token_hash TEXT NOT NULL,
				expires_at DATETIME NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			)`,
			`CREATE UNIQUE INDEX idx_sessions_token_hash ON sessions(token_hash)`,
			`CREATE INDEX idx_sessions_user_id ON sessions(user_id)`,
			`CREATE INDEX idx_sessions_expires_at ON sessions(expires_at)`,
		}
		for _, stmt := range stmts {
			if err := dbCtx.Exec(stmt).Error; err != nil {
				return err
			}
		}
		return nil
	case "mysql":
		// Detect the actual column type of users.id. Existing deployments may
		// have users.id as BIGINT (e.g. created by an earlier GORM AutoMigrate)
		// while fresh installs have INT. InnoDB requires the FK column type to
		// match exactly, so derive sessions.user_id from users.id at runtime.
		var userIDType string
		row := dbCtx.Raw(`SELECT COLUMN_TYPE FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'id'`).Row()
		if err := row.Scan(&userIDType); err != nil {
			return fmt.Errorf("failed to detect users.id column type: %s", err.Error())
		}
		if userIDType == "" {
			return fmt.Errorf("users.id column type could not be determined")
		}

		stmts := []string{
			fmt.Sprintf(`CREATE TABLE sessions (
				id %s AUTO_INCREMENT PRIMARY KEY,
				user_id %s NOT NULL,
				token_hash VARCHAR(64) NOT NULL,
				expires_at DATETIME NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				CONSTRAINT fk_users_sessions FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			)`, userIDType, userIDType),
			`CREATE UNIQUE INDEX idx_sessions_token_hash ON sessions(token_hash)`,
			`CREATE INDEX idx_sessions_user_id ON sessions(user_id)`,
			`CREATE INDEX idx_sessions_expires_at ON sessions(expires_at)`,
		}
		for _, stmt := range stmts {
			if err := dbCtx.Exec(stmt).Error; err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}
}

func (m *SessionsMigration) Down(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec("DROP TABLE IF EXISTS sessions").Error
}
