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
		return dbCtx.Exec(`CREATE TABLE sessions (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			token_hash VARCHAR(64) NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE INDEX idx_sessions_token_hash (token_hash),
			INDEX idx_sessions_user_id (user_id),
			INDEX idx_sessions_expires_at (expires_at),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`).Error
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}
}

func (m *SessionsMigration) Down(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec("DROP TABLE IF EXISTS sessions").Error
}
