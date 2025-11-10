package main

import (
"fmt"
"log"
"os"
"time"

"dkhalife.com/tasks/migration-tool/models"
"gorm.io/driver/sqlite"
"gorm.io/gorm"
)

func main() {
// Setup SQLite database with test data
sqliteDB := "test_booleans.db"
defer os.Remove(sqliteDB)
defer os.Remove(sqliteDB + "-shm")
defer os.Remove(sqliteDB + "-wal")

db, err := gorm.Open(sqlite.Open(sqliteDB), &gorm.Config{})
if err != nil {
log.Fatalf("Failed to open SQLite: %v", err)
}

db.AutoMigrate(&models.User{}, &models.Task{})

now := time.Now()
user := models.User{
DisplayName: "Test User",
Email:       "test@example.com",
Password:    "password",
CreatedAt:   now,
UpdatedAt:   now,
Disabled:    false,
}
db.Create(&user)

// Create an inactive task
inactiveTask := models.Task{
Title:     "Inactive Task",
IsActive:  false,
IsRolling: false,
CreatedBy: user.ID,
CreatedAt: now,
}
db.Create(&inactiveTask)
db.Model(&inactiveTask).Update("is_active", false)
db.Model(&inactiveTask).Update("is_rolling", false)

// Create an active task
activeTask := models.Task{
Title:     "Active Task",
IsActive:  true,
IsRolling: true,
CreatedBy: user.ID,
CreatedAt: now,
}
db.Create(&activeTask)

// Read back to verify SQLite storage
var tasksFromSQLite []models.Task
db.Find(&tasksFromSQLite)

fmt.Println("=== Tasks in SQLite ===")
for _, t := range tasksFromSQLite {
fmt.Printf("Task %d: %s - IsActive=%v, IsRolling=%v\n", t.ID, t.Title, t.IsActive, t.IsRolling)
}

// Check raw SQLite values
var rawTasks []map[string]interface{}
db.Table("tasks").Find(&rawTasks)
fmt.Println("\n=== Raw SQLite values ===")
for _, t := range rawTasks {
fmt.Printf("Task %v: %s - is_active=%v (type: %T), is_rolling=%v (type: %T)\n", 
t["id"], t["title"], t["is_active"], t["is_active"], t["is_rolling"], t["is_rolling"])
}
}
