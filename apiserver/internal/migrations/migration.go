package migrations

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"gorm.io/gorm"
)

type Migration interface {
	Version() int
	Name() string
	Up(ctx context.Context, db *gorm.DB) error
	Down(ctx context.Context, db *gorm.DB) error
}

type SchemaVersion struct {
	Version   int       `gorm:"primaryKey"`
	Name      string    `gorm:"size:255;not null"`
	AppliedAt time.Time `gorm:"not null"`
}

func (SchemaVersion) TableName() string {
	return "schema_versions"
}

var (
	registry     = make(map[int]Migration)
	registryLock sync.RWMutex
)

func Register(m Migration) {
	registryLock.Lock()
	defer registryLock.Unlock()

	version := m.Version()
	if existing, ok := registry[version]; ok {
		panic(fmt.Sprintf("migration version %d already registered by %q, cannot register %q",
			version, existing.Name(), m.Name()))
	}
	registry[version] = m
	log.Printf("Registered migration %d: %s", version, m.Name())
}

func GetMigrations() []Migration {
	registryLock.RLock()
	defer registryLock.RUnlock()

	migrations := make([]Migration, 0, len(registry))
	for _, m := range registry {
		migrations = append(migrations, m)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version() < migrations[j].Version()
	})

	return migrations
}

func GetMigration(version int) (Migration, bool) {
	registryLock.RLock()
	defer registryLock.RUnlock()

	m, ok := registry[version]
	return m, ok
}

func GetLatestVersion() int {
	registryLock.RLock()
	defer registryLock.RUnlock()

	latest := 0
	for version := range registry {
		if version > latest {
			latest = version
		}
	}
	return latest
}
