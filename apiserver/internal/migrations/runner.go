package migrations

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type Runner struct {
	db *gorm.DB
}

func NewRunner(db *gorm.DB) *Runner {
	return &Runner{db: db}
}

func (r *Runner) ensureSchemaVersionsTable(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(&SchemaVersion{})
}

func (r *Runner) GetCurrentVersion(ctx context.Context) (int, error) {
	if err := r.ensureSchemaVersionsTable(ctx); err != nil {
		return 0, fmt.Errorf("failed to ensure schema_versions table: %w", err)
	}

	var version SchemaVersion
	err := r.db.WithContext(ctx).Order("version DESC").First(&version).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get current version: %w", err)
	}
	return version.Version, nil
}

func (r *Runner) GetAppliedMigrations(ctx context.Context) ([]SchemaVersion, error) {
	if err := r.ensureSchemaVersionsTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure schema_versions table: %w", err)
	}

	var versions []SchemaVersion
	err := r.db.WithContext(ctx).Order("version ASC").Find(&versions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	return versions, nil
}

func (r *Runner) MigrateUp(ctx context.Context, targetVersion int) error {
	if err := r.ensureSchemaVersionsTable(ctx); err != nil {
		return fmt.Errorf("failed to ensure schema_versions table: %w", err)
	}

	currentVersion, err := r.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if targetVersion == 0 {
		targetVersion = GetLatestVersion()
	}

	if currentVersion >= targetVersion {
		log.Printf("Database is already at version %d (target: %d), no migrations needed", currentVersion, targetVersion)
		return nil
	}

	log.Printf("Migrating from version %d to %d", currentVersion, targetVersion)

	migrations := GetMigrations()
	for _, m := range migrations {
		version := m.Version()
		if version <= currentVersion || version > targetVersion {
			continue
		}

		log.Printf("Applying migration %d: %s", version, m.Name())

		if err := r.applyMigration(ctx, m); err != nil {
			return err
		}

		log.Printf("Successfully applied migration %d: %s", version, m.Name())
	}

	log.Printf("Migration complete. Database is now at version %d", targetVersion)
	return nil
}

func (r *Runner) MigrateDown(ctx context.Context, targetVersion int) error {
	if err := r.ensureSchemaVersionsTable(ctx); err != nil {
		return fmt.Errorf("failed to ensure schema_versions table: %w", err)
	}

	currentVersion, err := r.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion <= targetVersion {
		log.Printf("Database is already at version %d (target: %d), no rollbacks needed", currentVersion, targetVersion)
		return nil
	}

	log.Printf("Rolling back from version %d to %d", currentVersion, targetVersion)

	migrations := GetMigrations()
	for i := len(migrations) - 1; i >= 0; i-- {
		m := migrations[i]
		version := m.Version()
		if version <= targetVersion || version > currentVersion {
			continue
		}

		log.Printf("Rolling back migration %d: %s", version, m.Name())

		if err := r.rollbackMigration(ctx, m); err != nil {
			return err
		}

		log.Printf("Successfully rolled back migration %d: %s", version, m.Name())
	}

	log.Printf("Rollback complete. Database is now at version %d", targetVersion)
	return nil
}

func (r *Runner) applyMigration(ctx context.Context, m Migration) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := m.Up(ctx, tx); err != nil {
			return fmt.Errorf("migration %d (%s) failed: %w", m.Version(), m.Name(), err)
		}

		record := SchemaVersion{
			Version:   m.Version(),
			Name:      m.Name(),
			AppliedAt: time.Now().UTC(),
		}
		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("failed to record migration %d: %w", m.Version(), err)
		}
		return nil
	})
}

func (r *Runner) rollbackMigration(ctx context.Context, m Migration) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := m.Down(ctx, tx); err != nil {
			return fmt.Errorf("rollback of migration %d (%s) failed: %w", m.Version(), m.Name(), err)
		}

		if err := tx.Where("version = ?", m.Version()).Delete(&SchemaVersion{}).Error; err != nil {
			return fmt.Errorf("failed to remove migration record %d: %w", m.Version(), err)
		}
		return nil
	})
}

func (r *Runner) MigrateTo(ctx context.Context, targetVersion int) error {
	currentVersion, err := r.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion < targetVersion {
		return r.MigrateUp(ctx, targetVersion)
	} else if currentVersion > targetVersion {
		return r.MigrateDown(ctx, targetVersion)
	}

	log.Printf("Database is already at version %d", targetVersion)
	return nil
}

type MigrationStatus struct {
	Version   int
	Name      string
	Applied   bool
	AppliedAt *time.Time
}

func (r *Runner) GetStatus(ctx context.Context) ([]MigrationStatus, error) {
	applied, err := r.GetAppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	appliedMap := make(map[int]SchemaVersion)
	for _, v := range applied {
		appliedMap[v.Version] = v
	}

	migrations := GetMigrations()
	status := make([]MigrationStatus, len(migrations))

	for i, m := range migrations {
		s := MigrationStatus{
			Version: m.Version(),
			Name:    m.Name(),
			Applied: false,
		}

		if v, ok := appliedMap[m.Version()]; ok {
			s.Applied = true
			s.AppliedAt = &v.AppliedAt
		}

		status[i] = s
	}

	return status, nil
}
