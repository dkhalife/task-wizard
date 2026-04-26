package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func init() {
	Register(&UniqueLabelNamesMigration{})
}

type UniqueLabelNamesMigration struct{}

func (m *UniqueLabelNamesMigration) Version() int {
	return 8
}

func (m *UniqueLabelNamesMigration) Name() string {
	return "unique_label_names"
}

func (m *UniqueLabelNamesMigration) Up(ctx context.Context, db *gorm.DB) error {
	dbCtx := db.WithContext(ctx)

	dedup := []string{
		// Reassign task_labels from duplicate labels to the kept (lowest ID) label,
		// skipping rows that would collide with an existing (task_id, keep_id) pair
		`INSERT OR IGNORE INTO task_labels (task_id, label_id)
		SELECT tl.task_id, keep.id
		FROM task_labels tl
		JOIN labels dup ON tl.label_id = dup.id
		JOIN (
			SELECT MIN(id) AS id, name, created_by
			FROM labels
			GROUP BY created_by, name
		) keep ON dup.created_by = keep.created_by AND dup.name = keep.name
		WHERE dup.id != keep.id`,

		// Remove task_labels pointing to duplicate labels
		`DELETE FROM task_labels WHERE label_id IN (
			SELECT l.id FROM labels l
			JOIN (
				SELECT MIN(id) AS keep_id, name, created_by
				FROM labels
				GROUP BY created_by, name
			) k ON l.created_by = k.created_by AND l.name = k.name
			WHERE l.id != k.keep_id
		)`,

		// Delete the duplicate labels themselves
		`DELETE FROM labels WHERE id NOT IN (
			SELECT MIN(id) FROM labels GROUP BY created_by, name
		)`,
	}

	dialect := db.Dialector.Name()
	if dialect == "mysql" {
		dedup = []string{
			`INSERT IGNORE INTO task_labels (task_id, label_id)
			SELECT tl.task_id, keep_tbl.keep_id
			FROM task_labels tl
			JOIN labels dup ON tl.label_id = dup.id
			JOIN (
				SELECT MIN(id) AS keep_id, name, created_by
				FROM labels
				GROUP BY created_by, name
			) keep_tbl ON dup.created_by = keep_tbl.created_by AND dup.name = keep_tbl.name
			WHERE dup.id != keep_tbl.keep_id`,

			`DELETE tl FROM task_labels tl
			JOIN labels l ON tl.label_id = l.id
			JOIN (
				SELECT MIN(id) AS keep_id, name, created_by
				FROM labels
				GROUP BY created_by, name
			) k ON l.created_by = k.created_by AND l.name = k.name
			WHERE l.id != k.keep_id`,

			`DELETE l FROM labels l
			JOIN (
				SELECT MIN(id) AS keep_id, name, created_by
				FROM labels
				GROUP BY created_by, name
			) k ON l.created_by = k.created_by AND l.name = k.name
			WHERE l.id != k.keep_id`,
		}
	}

	for _, stmt := range dedup {
		if err := dbCtx.Exec(stmt).Error; err != nil {
			return fmt.Errorf("failed to deduplicate labels: %w", err)
		}
	}

	return dbCtx.Exec("CREATE UNIQUE INDEX idx_labels_created_by_name ON labels(created_by, name)").Error
}

func (m *UniqueLabelNamesMigration) Down(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec("DROP INDEX IF EXISTS idx_labels_created_by_name").Error
}
