package main

import (
	"os"
	"testing"
	"time"

	"dkhalife.com/tasks/migration-tool/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDatabase creates a SQLite database with test data including completed tasks
func setupTestDatabase(t *testing.T, dbPath string) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Auto migrate
	err = db.AutoMigrate(
		&models.User{},
		&models.Label{},
		&models.Task{},
		&models.TaskHistory{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	// Create a test user
	user := models.User{
		DisplayName: "Test User",
		Email:       "test@example.com",
		Password:    "hashed_password",
		CreatedAt:   now,
		UpdatedAt:   now,
		Disabled:    false,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a completed task (IsActive = false, has TaskHistory with CompletedDate)
	completedTask := models.Task{
		Title:       "Completed Task",
		NextDueDate: nil,
		CreatedBy:   user.ID,
		CreatedAt:   now,
	}
	// First create the task
	if err := db.Create(&completedTask).Error; err != nil {
		t.Fatalf("Failed to create completed task: %v", err)
	}
	// Then explicitly update IsActive to false to override the default
	if err := db.Model(&completedTask).Update("is_active", false).Error; err != nil {
		t.Fatalf("Failed to update completed task IsActive: %v", err)
	}
	
	// Verify the task was stored correctly
	var verifyTask models.Task
	if err := db.First(&verifyTask, completedTask.ID).Error; err != nil {
		t.Fatalf("Failed to verify completed task: %v", err)
	}
	if verifyTask.IsActive {
		t.Fatalf("Completed task should have IsActive=false, but has IsActive=true")
	}

	// Create TaskHistory for the completed task
	completedHistory := models.TaskHistory{
		TaskID:        completedTask.ID,
		CompletedDate: &yesterday,
		DueDate:       &yesterday,
	}
	if err := db.Create(&completedHistory).Error; err != nil {
		t.Fatalf("Failed to create task history: %v", err)
	}

	// Create an active task (IsActive = true, no completion history)
	futureDate := now.Add(48 * time.Hour)
	activeTask := models.Task{
		Title:       "Active Task",
		NextDueDate: &futureDate,
		IsActive:    true,
		CreatedBy:   user.ID,
		CreatedAt:   now,
		UpdatedAt:   &now,
	}
	if err := db.Create(&activeTask).Error; err != nil {
		t.Fatalf("Failed to create active task: %v", err)
	}

	t.Logf("Created test database with:")
	t.Logf("  - 1 user (ID: %d)", user.ID)
	t.Logf("  - 1 completed task (ID: %d, IsActive: %v)", completedTask.ID, completedTask.IsActive)
	t.Logf("  - 1 active task (ID: %d, IsActive: %v)", activeTask.ID, activeTask.IsActive)
	t.Logf("  - 1 task history (TaskID: %d, CompletedDate: %v)", completedHistory.TaskID, completedHistory.CompletedDate != nil)
}

// TestCompletedTasksMigration tests that completed tasks are correctly migrated
func TestCompletedTasksMigration(t *testing.T) {
	// Create source SQLite database
	sourceDB := "testdata/test_source.db"
	defer os.Remove(sourceDB)
	defer os.Remove(sourceDB + "-shm")
	defer os.Remove(sourceDB + "-wal")

	setupTestDatabase(t, sourceDB)

	// Create target SQLite database (simulating MariaDB for test)
	targetDB := "testdata/test_target.db"
	defer os.Remove(targetDB)
	defer os.Remove(targetDB + "-shm")
	defer os.Remove(targetDB + "-wal")

	// Open connections
	sourceConn, err := openSQLiteReadOnly(sourceDB)
	if err != nil {
		t.Fatalf("Failed to open source database: %v", err)
	}

	targetConn, err := gorm.Open(sqlite.Open(targetDB), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open target database: %v", err)
	}

	// Run migrations on target
	err = targetConn.AutoMigrate(
		&models.User{},
		&models.Label{},
		&models.Task{},
		&models.TaskHistory{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate target database: %v", err)
	}

	// Perform migration using a modified version that doesn't use MySQL-specific commands
	// For testing purposes, we'll call a version without foreign key checks
	err = targetConn.Transaction(func(tx *gorm.DB) error {
		// Don't disable foreign key checks for SQLite in tests
		// Just perform the migration
		
		// Migrate Users
		var users []models.User
		if err := sourceConn.Find(&users).Error; err != nil {
			return err
		}
		if len(users) > 0 {
			if err := tx.Create(&users).Error; err != nil {
				return err
			}
		}

		// Migrate Tasks
		var tasks []models.Task
		if err := sourceConn.Find(&tasks).Error; err != nil {
			return err
		}
		if len(tasks) > 0 {
			// Use Select("*") to force GORM to insert all fields as-is
			if err := tx.Select("*").Create(&tasks).Error; err != nil {
				return err
			}
		}

		// Migrate TaskHistory
		var taskHistories []models.TaskHistory
		if err := sourceConn.Find(&taskHistories).Error; err != nil {
			return err
		}
		if len(taskHistories) > 0 {
			if err := tx.Create(&taskHistories).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify data in target database
	var tasks []models.Task
	if err := targetConn.Find(&tasks).Error; err != nil {
		t.Fatalf("Failed to read tasks from target: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}

	// Check that we have one active and one inactive task
	var activeCount, inactiveCount int
	for _, task := range tasks {
		if task.IsActive {
			activeCount++
			if task.Title != "Active Task" {
				t.Errorf("Active task has wrong title: %s", task.Title)
			}
		} else {
			inactiveCount++
			if task.Title != "Completed Task" {
				t.Errorf("Inactive task has wrong title: %s", task.Title)
			}
		}
	}

	if activeCount != 1 {
		t.Errorf("Expected 1 active task, got %d", activeCount)
	}
	if inactiveCount != 1 {
		t.Errorf("Expected 1 inactive task, got %d", inactiveCount)
	}

	// Verify TaskHistory was migrated
	var histories []models.TaskHistory
	if err := targetConn.Find(&histories).Error; err != nil {
		t.Fatalf("Failed to read task histories from target: %v", err)
	}

	if len(histories) != 1 {
		t.Errorf("Expected 1 task history, got %d", len(histories))
	}

	if len(histories) > 0 {
		if histories[0].CompletedDate == nil {
			t.Error("Task history missing CompletedDate")
		}
		t.Logf("Task history migrated successfully: TaskID=%d, CompletedDate=%v",
			histories[0].TaskID, histories[0].CompletedDate)
	}
}

// TestTaskHistoryPreservation specifically tests that task history records are preserved
func TestTaskHistoryPreservation(t *testing.T) {
	sourceDB := "testdata/test_history_source.db"
	defer os.Remove(sourceDB)
	defer os.Remove(sourceDB + "-shm")
	defer os.Remove(sourceDB + "-wal")

	targetDB := "testdata/test_history_target.db"
	defer os.Remove(targetDB)
	defer os.Remove(targetDB + "-shm")
	defer os.Remove(targetDB + "-wal")

	// Setup source database
	db, err := gorm.Open(sqlite.Open(sourceDB), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open source database: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Label{}, &models.Task{}, &models.TaskHistory{}, &models.AppToken{}, &models.UserPasswordReset{}, &models.NotificationSettings{}, &models.Notification{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}
	
	// Manually create task_labels table (many-to-many join table)
	err = db.Exec(`CREATE TABLE IF NOT EXISTS task_labels (
		task_id INTEGER,
		label_id INTEGER,
		PRIMARY KEY (task_id, label_id)
	)`).Error
	if err != nil {
		t.Fatalf("Failed to create task_labels table: %v", err)
	}

	now := time.Now()
	user := models.User{
		DisplayName: "Test User",
		Email:       "history@example.com",
		Password:    "password",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	db.Create(&user)

	task := models.Task{
		Title:     "Recurring Task",
		IsActive:  true,
		CreatedBy: user.ID,
		CreatedAt: now,
	}
	db.Create(&task)

	// Create multiple history entries (simulating multiple completions of recurring task)
	for i := 0; i < 3; i++ {
		completedDate := now.Add(time.Duration(-24*(i+1)) * time.Hour)
		history := models.TaskHistory{
			TaskID:        task.ID,
			CompletedDate: &completedDate,
			DueDate:       &completedDate,
		}
		if err := db.Create(&history).Error; err != nil {
			t.Fatalf("Failed to create history entry %d: %v", i, err)
		}
	}

	// Perform migration
	sourceConn, _ := openSQLiteReadOnly(sourceDB)
	targetConn, _ := gorm.Open(sqlite.Open(targetDB), &gorm.Config{})
	targetConn.AutoMigrate(&models.User{}, &models.Label{}, &models.Task{}, &models.TaskHistory{}, &models.AppToken{}, &models.UserPasswordReset{}, &models.NotificationSettings{}, &models.Notification{})
	
	// Manually create task_labels table (many-to-many join table)
	targetConn.Exec(`CREATE TABLE IF NOT EXISTS task_labels (
		task_id INTEGER,
		label_id INTEGER,
		PRIMARY KEY (task_id, label_id)
	)`)

	if err := migrateData(sourceConn, targetConn); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify all history entries were migrated
	var histories []models.TaskHistory
	targetConn.Find(&histories)

	if len(histories) != 3 {
		t.Errorf("Expected 3 history entries, got %d", len(histories))
	}

	for i, h := range histories {
		if h.CompletedDate == nil {
			t.Errorf("History entry %d missing CompletedDate", i)
		}
		if h.TaskID != task.ID {
			t.Errorf("History entry %d has wrong TaskID: expected %d, got %d", i, task.ID, h.TaskID)
		}
	}
}

// TestBooleanFieldMigration tests that boolean fields are correctly migrated from SQLite to target DB
func TestBooleanFieldMigration(t *testing.T) {
	// Create source SQLite database
	sourceDB := "testdata/test_booleans_source.db"
	defer os.Remove(sourceDB)
	defer os.Remove(sourceDB + "-shm")
	defer os.Remove(sourceDB + "-wal")

	db, err := gorm.Open(sqlite.Open(sourceDB), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open source database: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Task{},
		&models.Notification{},
		&models.NotificationSettings{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	now := time.Now()

	// Create users with different Disabled states
	activeUser := models.User{
		DisplayName: "Active User",
		Email:       "active@example.com",
		Password:    "password",
		Disabled:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	db.Create(&activeUser)

	disabledUser := models.User{
		DisplayName: "Disabled User",
		Email:       "disabled@example.com",
		Password:    "password",
		Disabled:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	db.Create(&disabledUser)

	// Create tasks with different boolean combinations
	task1 := models.Task{
		Title:     "Task 1 - Active, Rolling",
		IsActive:  true,
		IsRolling: true,
		CreatedBy: activeUser.ID,
		CreatedAt: now,
	}
	db.Create(&task1)

	task2 := models.Task{
		Title:     "Task 2 - Active, Not Rolling",
		IsActive:  true,
		IsRolling: false,
		CreatedBy: activeUser.ID,
		CreatedAt: now,
	}
	db.Create(&task2)

	task3 := models.Task{
		Title:     "Task 3 - Inactive, Not Rolling",
		CreatedBy: activeUser.ID,
		CreatedAt: now,
	}
	db.Create(&task3)
	// Explicitly set both to false
	db.Model(&task3).Updates(map[string]interface{}{"is_active": false, "is_rolling": false})

	task4 := models.Task{
		Title:     "Task 4 - Inactive, Rolling",
		IsRolling: true,
		CreatedBy: activeUser.ID,
		CreatedAt: now,
	}
	db.Create(&task4)
	db.Model(&task4).Update("is_active", false)

	// Create notifications with different IsSent states
	notif1 := models.Notification{
		TaskID:       task1.ID,
		UserID:       activeUser.ID,
		Text:         "Notification 1 - Sent",
		Type:         models.NotificationTypeDueDate,
		IsSent:       true,
		ScheduledFor: now,
		CreatedAt:    now,
	}
	db.Create(&notif1)

	notif2 := models.Notification{
		TaskID:       task1.ID,
		UserID:       activeUser.ID,
		Text:         "Notification 2 - Not Sent",
		Type:         models.NotificationTypeDueDate,
		IsSent:       false,
		ScheduledFor: now,
		CreatedAt:    now,
	}
	db.Create(&notif2)

	// Create notification settings with different trigger booleans
	settings1 := models.NotificationSettings{
		UserID: activeUser.ID,
		Triggers: models.NotificationTriggerOptions{
			Enabled: true,
			DueDate: true,
			PreDue:  true,
			Overdue: false,
		},
	}
	db.Create(&settings1)

	settings2 := models.NotificationSettings{
		UserID: disabledUser.ID,
		Triggers: models.NotificationTriggerOptions{
			Enabled: false,
			DueDate: false,
			PreDue:  false,
			Overdue: false,
		},
	}
	db.Create(&settings2)

	// Create target SQLite database (simulating MariaDB for test)
	targetDB := "testdata/test_booleans_target.db"
	defer os.Remove(targetDB)
	defer os.Remove(targetDB + "-shm")
	defer os.Remove(targetDB + "-wal")

	// Perform migration
	sourceConn, err := openSQLiteReadOnly(sourceDB)
	if err != nil {
		t.Fatalf("Failed to open source: %v", err)
	}

	targetConn, err := gorm.Open(sqlite.Open(targetDB), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open target: %v", err)
	}

	err = targetConn.AutoMigrate(
		&models.User{},
		&models.Task{},
		&models.Notification{},
		&models.NotificationSettings{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate target: %v", err)
	}

	// Perform migration using a simplified version without foreign key checks
	err = targetConn.Transaction(func(tx *gorm.DB) error {
		// Migrate Users
		var users []models.User
		if err := sourceConn.Find(&users).Error; err != nil {
			return err
		}
		if len(users) > 0 {
			if err := tx.Select("*").Omit("UpdatedAt").Create(&users).Error; err != nil {
				return err
			}
		}

		// Migrate Tasks
		var tasks []models.Task
		if err := sourceConn.Find(&tasks).Error; err != nil {
			return err
		}
		if len(tasks) > 0 {
			if err := tx.Select("*").Omit("UpdatedAt").Create(&tasks).Error; err != nil {
				return err
			}
		}

		// Migrate Notifications
		var notifications []models.Notification
		if err := sourceConn.Find(&notifications).Error; err != nil {
			return err
		}
		if len(notifications) > 0 {
			if err := tx.Select("*").Create(&notifications).Error; err != nil {
				return err
			}
		}

		// Migrate NotificationSettings
		var notificationSettings []models.NotificationSettings
		if err := sourceConn.Find(&notificationSettings).Error; err != nil {
			return err
		}
		if len(notificationSettings) > 0 {
			if err := tx.Select("*").Create(&notificationSettings).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify Users
	var users []models.User
	targetConn.Order("id").Find(&users)
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
	if users[0].Disabled != false {
		t.Errorf("User 1 should not be disabled, got Disabled=%v", users[0].Disabled)
	}
	if users[1].Disabled != true {
		t.Errorf("User 2 should be disabled, got Disabled=%v", users[1].Disabled)
	}

	// Verify Tasks
	var tasks []models.Task
	targetConn.Order("id").Find(&tasks)
	if len(tasks) != 4 {
		t.Errorf("Expected 4 tasks, got %d", len(tasks))
	}

	// Task 1: IsActive=true, IsRolling=true
	if tasks[0].IsActive != true || tasks[0].IsRolling != true {
		t.Errorf("Task 1 should be IsActive=true, IsRolling=true, got IsActive=%v, IsRolling=%v",
			tasks[0].IsActive, tasks[0].IsRolling)
	}

	// Task 2: IsActive=true, IsRolling=false
	if tasks[1].IsActive != true || tasks[1].IsRolling != false {
		t.Errorf("Task 2 should be IsActive=true, IsRolling=false, got IsActive=%v, IsRolling=%v",
			tasks[1].IsActive, tasks[1].IsRolling)
	}

	// Task 3: IsActive=false, IsRolling=false
	if tasks[2].IsActive != false || tasks[2].IsRolling != false {
		t.Errorf("Task 3 should be IsActive=false, IsRolling=false, got IsActive=%v, IsRolling=%v",
			tasks[2].IsActive, tasks[2].IsRolling)
	}

	// Task 4: IsActive=false, IsRolling=true
	if tasks[3].IsActive != false || tasks[3].IsRolling != true {
		t.Errorf("Task 4 should be IsActive=false, IsRolling=true, got IsActive=%v, IsRolling=%v",
			tasks[3].IsActive, tasks[3].IsRolling)
	}

	// Verify Notifications
	var notifications []models.Notification
	targetConn.Order("id").Find(&notifications)
	if len(notifications) != 2 {
		t.Errorf("Expected 2 notifications, got %d", len(notifications))
	}
	if notifications[0].IsSent != true {
		t.Errorf("Notification 1 should be sent, got IsSent=%v", notifications[0].IsSent)
	}
	if notifications[1].IsSent != false {
		t.Errorf("Notification 2 should not be sent, got IsSent=%v", notifications[1].IsSent)
	}

	// Verify NotificationSettings
	var settings []models.NotificationSettings
	targetConn.Order("user_id").Find(&settings)
	if len(settings) != 2 {
		t.Errorf("Expected 2 notification settings, got %d", len(settings))
	}

	// Settings 1: All enabled except Overdue
	if settings[0].Triggers.Enabled != true || settings[0].Triggers.DueDate != true ||
		settings[0].Triggers.PreDue != true || settings[0].Triggers.Overdue != false {
		t.Errorf("Settings 1 incorrect: Enabled=%v, DueDate=%v, PreDue=%v, Overdue=%v",
			settings[0].Triggers.Enabled, settings[0].Triggers.DueDate,
			settings[0].Triggers.PreDue, settings[0].Triggers.Overdue)
	}

	// Settings 2: All disabled
	if settings[1].Triggers.Enabled != false || settings[1].Triggers.DueDate != false ||
		settings[1].Triggers.PreDue != false || settings[1].Triggers.Overdue != false {
		t.Errorf("Settings 2 incorrect: Enabled=%v, DueDate=%v, PreDue=%v, Overdue=%v",
			settings[1].Triggers.Enabled, settings[1].Triggers.DueDate,
			settings[1].Triggers.PreDue, settings[1].Triggers.Overdue)
	}

	t.Log("All boolean fields migrated correctly!")
}
