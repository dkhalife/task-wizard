package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"dkhalife.com/tasks/migration-tool/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func main() {
	// Define command-line flags
	sqlitePath := flag.String("sqlite", "", "Path to the SQLite database file (read-only)")
	mariaHost := flag.String("maria-host", "localhost", "MariaDB host")
	mariaPort := flag.Int("maria-port", 3306, "MariaDB port")
	mariaDB := flag.String("maria-db", "", "MariaDB database name")
	mariaUser := flag.String("maria-user", "", "MariaDB username")
	mariaPass := flag.String("maria-pass", "", "MariaDB password")
	
	flag.Parse()

	// Validate required flags
	if *sqlitePath == "" {
		log.Fatal("Error: --sqlite flag is required")
	}
	if *mariaDB == "" {
		log.Fatal("Error: --maria-db flag is required")
	}
	if *mariaUser == "" {
		log.Fatal("Error: --maria-user flag is required")
	}

	// Check if SQLite file exists
	if _, err := os.Stat(*sqlitePath); os.IsNotExist(err) {
		log.Fatalf("Error: SQLite file does not exist: %s", *sqlitePath)
	}

	log.Println("Starting migration from SQLite to MariaDB...")
	log.Printf("SQLite file: %s", *sqlitePath)
	log.Printf("MariaDB: %s@%s:%d/%s", *mariaUser, *mariaHost, *mariaPort, *mariaDB)

	// Open SQLite connection (read-only)
	sqliteDB, err := openSQLiteReadOnly(*sqlitePath)
	if err != nil {
		log.Fatalf("Error opening SQLite database: %v", err)
	}
	log.Println("✓ Connected to SQLite database (read-only)")

	// Open MariaDB connection
	mariaDBConn, err := openMariaDB(*mariaHost, *mariaPort, *mariaDB, *mariaUser, *mariaPass)
	if err != nil {
		log.Fatalf("Error opening MariaDB database: %v", err)
	}
	log.Println("✓ Connected to MariaDB database")

	// Perform migration within a transaction
	if err := migrateData(sqliteDB, mariaDBConn); err != nil {
		log.Fatalf("Error during migration: %v", err)
	}

	log.Println("✓ Migration completed successfully!")
}

// openSQLiteReadOnly opens a SQLite database in read-only mode
func openSQLiteReadOnly(path string) (*gorm.DB, error) {
	// Open SQLite with read-only mode and immutable flag
	dsn := fmt.Sprintf("file:%s?mode=ro&immutable=1", path)
	
	logger := gormLogger.Default.LogMode(gormLogger.Warn)
	
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	return db, nil
}

// openMariaDB opens a connection to MariaDB
func openMariaDB(host string, port int, database, username, password string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, port, database)
	
	logger := gormLogger.Default.LogMode(gormLogger.Warn)
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open MariaDB database: %w", err)
	}

	return db, nil
}

// migrateData copies all data from SQLite to MariaDB in the correct order
func migrateData(sqliteDB, mariaDB *gorm.DB) error {
	// Start a transaction on the MariaDB connection
	return mariaDB.Transaction(func(tx *gorm.DB) error {
		// Migration order based on foreign key dependencies:
		// 1. Users (no dependencies)
		// 2. Labels (depends on Users via created_by)
		// 3. Tasks (depends on Users via created_by)
		// 4. TaskHistory (depends on Tasks)
		// 5. Notifications (depends on Tasks and Users)
		// 6. AppTokens (depends on Users)
		// 7. UserPasswordReset (depends on Users)
		// 8. NotificationSettings (depends on Users)
		// 9. TaskLabel (join table - depends on Tasks and Labels)

		// 1. Migrate Users
		log.Println("Migrating users...")
		var users []models.User
		if err := sqliteDB.Find(&users).Error; err != nil {
			return fmt.Errorf("failed to read users from SQLite: %w", err)
		}
		if len(users) > 0 {
			if err := tx.Create(&users).Error; err != nil {
				return fmt.Errorf("failed to insert users into MariaDB: %w", err)
			}
			log.Printf("  ✓ Migrated %d users", len(users))
		} else {
			log.Println("  ✓ No users to migrate")
		}

		// 2. Migrate Labels
		log.Println("Migrating labels...")
		var labels []models.Label
		if err := sqliteDB.Find(&labels).Error; err != nil {
			return fmt.Errorf("failed to read labels from SQLite: %w", err)
		}
		if len(labels) > 0 {
			if err := tx.Create(&labels).Error; err != nil {
				return fmt.Errorf("failed to insert labels into MariaDB: %w", err)
			}
			log.Printf("  ✓ Migrated %d labels", len(labels))
		} else {
			log.Println("  ✓ No labels to migrate")
		}

		// 3. Migrate Tasks
		log.Println("Migrating tasks...")
		var tasks []models.Task
		if err := sqliteDB.Find(&tasks).Error; err != nil {
			return fmt.Errorf("failed to read tasks from SQLite: %w", err)
		}
		if len(tasks) > 0 {
			if err := tx.Create(&tasks).Error; err != nil {
				return fmt.Errorf("failed to insert tasks into MariaDB: %w", err)
			}
			log.Printf("  ✓ Migrated %d tasks", len(tasks))
		} else {
			log.Println("  ✓ No tasks to migrate")
		}

		// 4. Migrate TaskHistory
		log.Println("Migrating task history...")
		var taskHistories []models.TaskHistory
		if err := sqliteDB.Find(&taskHistories).Error; err != nil {
			return fmt.Errorf("failed to read task histories from SQLite: %w", err)
		}
		if len(taskHistories) > 0 {
			if err := tx.Create(&taskHistories).Error; err != nil {
				return fmt.Errorf("failed to insert task histories into MariaDB: %w", err)
			}
			log.Printf("  ✓ Migrated %d task history records", len(taskHistories))
		} else {
			log.Println("  ✓ No task history records to migrate")
		}

		// 5. Migrate Notifications
		log.Println("Migrating notifications...")
		var notifications []models.Notification
		if err := sqliteDB.Find(&notifications).Error; err != nil {
			return fmt.Errorf("failed to read notifications from SQLite: %w", err)
		}
		if len(notifications) > 0 {
			if err := tx.Create(&notifications).Error; err != nil {
				return fmt.Errorf("failed to insert notifications into MariaDB: %w", err)
			}
			log.Printf("  ✓ Migrated %d notifications", len(notifications))
		} else {
			log.Println("  ✓ No notifications to migrate")
		}

		// 6. Migrate AppTokens
		log.Println("Migrating app tokens...")
		var appTokens []models.AppToken
		if err := sqliteDB.Find(&appTokens).Error; err != nil {
			return fmt.Errorf("failed to read app tokens from SQLite: %w", err)
		}
		if len(appTokens) > 0 {
			if err := tx.Create(&appTokens).Error; err != nil {
				return fmt.Errorf("failed to insert app tokens into MariaDB: %w", err)
			}
			log.Printf("  ✓ Migrated %d app tokens", len(appTokens))
		} else {
			log.Println("  ✓ No app tokens to migrate")
		}

		// 7. Migrate UserPasswordReset
		log.Println("Migrating password reset tokens...")
		var passwordResets []models.UserPasswordReset
		if err := sqliteDB.Find(&passwordResets).Error; err != nil {
			return fmt.Errorf("failed to read password resets from SQLite: %w", err)
		}
		if len(passwordResets) > 0 {
			if err := tx.Create(&passwordResets).Error; err != nil {
				return fmt.Errorf("failed to insert password resets into MariaDB: %w", err)
			}
			log.Printf("  ✓ Migrated %d password reset tokens", len(passwordResets))
		} else {
			log.Println("  ✓ No password reset tokens to migrate")
		}

		// 8. Migrate NotificationSettings
		log.Println("Migrating notification settings...")
		var notificationSettings []models.NotificationSettings
		if err := sqliteDB.Find(&notificationSettings).Error; err != nil {
			return fmt.Errorf("failed to read notification settings from SQLite: %w", err)
		}
		if len(notificationSettings) > 0 {
			if err := tx.Create(&notificationSettings).Error; err != nil {
				return fmt.Errorf("failed to insert notification settings into MariaDB: %w", err)
			}
			log.Printf("  ✓ Migrated %d notification settings", len(notificationSettings))
		} else {
			log.Println("  ✓ No notification settings to migrate")
		}

		// 9. Migrate TaskLabel (many-to-many join table)
		log.Println("Migrating task-label associations...")
		var taskLabels []models.TaskLabel
		if err := sqliteDB.Table("task_labels").Find(&taskLabels).Error; err != nil {
			return fmt.Errorf("failed to read task-label associations from SQLite: %w", err)
		}
		if len(taskLabels) > 0 {
			if err := tx.Table("task_labels").Create(&taskLabels).Error; err != nil {
				return fmt.Errorf("failed to insert task-label associations into MariaDB: %w", err)
			}
			log.Printf("  ✓ Migrated %d task-label associations", len(taskLabels))
		} else {
			log.Println("  ✓ No task-label associations to migrate")
		}

		return nil
	})
}
