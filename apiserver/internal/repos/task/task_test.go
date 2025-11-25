package repos

import (
	"context"
	"fmt"
	"testing"
	"time"

	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/utils/test"
	"github.com/stretchr/testify/suite"
)

type TaskTestSuite struct {
	test.DatabaseTestSuite
	repo     *TaskRepository
	testUser *models.User
}

func TestTaskTestSuite(t *testing.T) {
	suite.Run(t, new(TaskTestSuite))
}

func (s *TaskTestSuite) SetupTest() {
	s.DatabaseTestSuite.SetupTest()
	s.repo = &TaskRepository{db: s.DB}

	s.testUser = &models.User{
		ID:        1,
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	err := s.DB.Create(s.testUser).Error
	s.Require().NoError(err)
}

func (s *TaskTestSuite) TestCreateTask() {
	ctx := context.Background()
	dueDate := time.Now().Add(24 * time.Hour)

	task := &models.Task{
		Title:       "Test Task",
		CreatedBy:   s.testUser.ID,
		NextDueDate: &dueDate,
		IsRolling:   false,
		IsActive:    true,
		Frequency: models.Frequency{
			Type: models.RepeatOnce,
		},
	}

	id, err := s.repo.CreateTask(ctx, task)
	s.Require().NoError(err)
	s.Require().Greater(id, 0)

	var savedTask models.Task
	err = s.DB.First(&savedTask, id).Error
	s.Require().NoError(err)
	s.Equal("Test Task", savedTask.Title)
	s.Equal(s.testUser.ID, savedTask.CreatedBy)
	s.WithinDuration(*savedTask.NextDueDate, dueDate, time.Second)
	s.Equal(false, savedTask.IsRolling)
	s.Equal(true, savedTask.IsActive)
	s.Equal(string(models.RepeatOnce), string(savedTask.Frequency.Type))
}

func (s *TaskTestSuite) TestCreateTaskWithEndDate() {
	ctx := context.Background()
	dueDate := time.Now().Add(24 * time.Hour)
	endDate := time.Now().Add(30 * 24 * time.Hour) // 30 days from now

	task := &models.Task{
		Title:       "Recurring Task with End Date",
		CreatedBy:   s.testUser.ID,
		NextDueDate: &dueDate,
		EndDate:     &endDate,
		IsRolling:   false,
		IsActive:    true,
		Frequency: models.Frequency{
			Type: models.RepeatWeekly,
		},
	}

	id, err := s.repo.CreateTask(ctx, task)
	s.Require().NoError(err)
	s.Require().Greater(id, 0)

	var savedTask models.Task
	err = s.DB.First(&savedTask, id).Error
	s.Require().NoError(err)
	s.Equal("Recurring Task with End Date", savedTask.Title)
	s.Equal(s.testUser.ID, savedTask.CreatedBy)
	s.WithinDuration(*savedTask.NextDueDate, dueDate, time.Second)
	s.WithinDuration(*savedTask.EndDate, endDate, time.Second)
	s.Equal(false, savedTask.IsRolling)
	s.Equal(true, savedTask.IsActive)
	s.Equal(string(models.RepeatWeekly), string(savedTask.Frequency.Type))
}

func (s *TaskTestSuite) TestUpsertTask() {
	ctx := context.Background()
	dueDate := time.Now().Add(24 * time.Hour)

	task := &models.Task{
		Title:       "Test Task",
		CreatedBy:   s.testUser.ID,
		NextDueDate: &dueDate,
		IsRolling:   false,
		IsActive:    true,
		Frequency: models.Frequency{
			Type: models.RepeatOnce,
		},
	}

	// Create
	err := s.repo.UpsertTask(ctx, task)
	s.Require().NoError(err)
	s.Require().Greater(task.ID, 0)

	// Update
	task.Title = "Updated Test Task"
	err = s.repo.UpsertTask(ctx, task)
	s.Require().NoError(err)

	var updatedTask models.Task
	err = s.DB.First(&updatedTask, task.ID).Error
	s.Require().NoError(err)
	s.Equal("Updated Test Task", updatedTask.Title)
}

func (s *TaskTestSuite) TestGetTask() {
	ctx := context.Background()
	dueDate := time.Now().Add(24 * time.Hour)

	task := &models.Task{
		Title:       "Test Task",
		CreatedBy:   s.testUser.ID,
		NextDueDate: &dueDate,
		IsRolling:   false,
		IsActive:    true,
		Frequency: models.Frequency{
			Type: models.RepeatOnce,
		},
	}

	err := s.DB.Create(task).Error
	s.Require().NoError(err)

	// Create a label and associate it with the task
	label := &models.Label{
		Name:      "Test Label",
		Color:     "#FF0000",
		CreatedBy: s.testUser.ID,
	}
	err = s.DB.Create(label).Error
	s.Require().NoError(err)

	err = s.DB.Exec("INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", task.ID, label.ID).Error
	s.Require().NoError(err)

	// Test retrieval with labels preloaded
	retrievedTask, err := s.repo.GetTask(ctx, task.ID)
	s.Require().NoError(err)
	s.Equal(task.ID, retrievedTask.ID)
	s.Equal("Test Task", retrievedTask.Title)
	s.Equal(s.testUser.ID, retrievedTask.CreatedBy)
	s.WithinDuration(*retrievedTask.NextDueDate, dueDate, time.Second)

	s.Require().Len(retrievedTask.Labels, 1)
	s.Equal("Test Label", retrievedTask.Labels[0].Name)
	s.Equal("#FF0000", retrievedTask.Labels[0].Color)
}

func (s *TaskTestSuite) TestGetTasks() {
	ctx := context.Background()

	// Create multiple tasks
	dueDate1 := time.Now().Add(24 * time.Hour)
	dueDate2 := time.Now().Add(48 * time.Hour)

	tasks := []*models.Task{
		{
			Title:       "Task 1",
			CreatedBy:   s.testUser.ID,
			NextDueDate: &dueDate1,
			IsActive:    true,
			Frequency: models.Frequency{
				Type: models.RepeatOnce,
			},
		},
		{
			Title:       "Task 2",
			CreatedBy:   s.testUser.ID,
			NextDueDate: &dueDate2,
			IsActive:    true,
			Frequency: models.Frequency{
				Type: models.RepeatWeekly,
			},
		},
		{
			ID:          30,
			Title:       "Inactive Task",
			CreatedBy:   s.testUser.ID,
			NextDueDate: &dueDate1,
			IsActive:    true,
			Frequency: models.Frequency{
				Type: models.RepeatOnce,
			},
		},
	}

	for _, task := range tasks {
		err := s.DB.Create(task).Error
		s.Require().NoError(err)
	}

	// Mark task 30 as inactive
	err := s.DB.Model(&models.Task{}).Where("id = ?", 30).Update("is_active", false).Error
	s.Require().NoError(err)

	// Create another user with their own tasks
	anotherUser := &models.User{
		Email: "another@example.com",
	}

	err = s.DB.Create(anotherUser).Error
	s.Require().NoError(err)

	otherTask := &models.Task{
		Title:       "Other Task",
		CreatedBy:   anotherUser.ID,
		NextDueDate: &dueDate1,
		IsActive:    true,
	}
	err = s.DB.Create(otherTask).Error
	s.Require().NoError(err)

	// Test retrieval - should only get active tasks for test user
	retrievedTasks, err := s.repo.GetTasks(ctx, s.testUser.ID)
	s.Require().NoError(err)
	s.Require().Len(retrievedTasks, 2)

	// Should be ordered by due date
	s.Equal("Task 1", retrievedTasks[0].Title)
	s.Equal("Task 2", retrievedTasks[1].Title)
}

func (s *TaskTestSuite) TestDeleteTask() {
	ctx := context.Background()
	dueDate := time.Now().Add(24 * time.Hour)

	task := &models.Task{
		Title:       "Test Task",
		CreatedBy:   s.testUser.ID,
		NextDueDate: &dueDate,
		IsActive:    true,
	}

	err := s.DB.Create(task).Error
	s.Require().NoError(err)

	// Delete the task
	err = s.repo.DeleteTask(ctx, task.ID)
	s.Require().NoError(err)

	// Verify task is deleted
	var count int64
	err = s.DB.Model(&models.Task{}).Where("id = ?", task.ID).Count(&count).Error
	s.Require().NoError(err)
	s.Equal(int64(0), count)
}

func (s *TaskTestSuite) TestIsTaskOwner() {
	ctx := context.Background()
	dueDate := time.Now().Add(24 * time.Hour)

	task := &models.Task{
		Title:       "Test Task",
		CreatedBy:   s.testUser.ID,
		NextDueDate: &dueDate,
		IsActive:    true,
	}

	err := s.DB.Create(task).Error
	s.Require().NoError(err)

	// Test with correct owner
	err = s.repo.IsTaskOwner(ctx, task.ID, s.testUser.ID)
	s.Require().NoError(err)

	// Test with incorrect owner
	anotherUser := &models.User{
		Email: "another@example.com",
	}
	err = s.DB.Create(anotherUser).Error
	s.Require().NoError(err)

	err = s.repo.IsTaskOwner(ctx, task.ID, anotherUser.ID)
	s.Error(err)
}

func (s *TaskTestSuite) TestCompleteTask() {
	ctx := context.Background()
	dueDate := time.Now().Add(24 * time.Hour)
	completedDate := time.Now()

	// Create a non-recurring task
	task := &models.Task{
		Title:       "One-time Task",
		CreatedBy:   s.testUser.ID,
		NextDueDate: &dueDate,
		IsActive:    true,
		Frequency: models.Frequency{
			Type: models.RepeatOnce,
		},
	}

	err := s.DB.Create(task).Error
	s.Require().NoError(err)

	// Complete the one-time task
	err = s.repo.CompleteTask(ctx, task, s.testUser.ID, nil, &completedDate)
	s.Require().NoError(err)

	// Check task is now inactive
	var updatedTask models.Task
	err = s.DB.First(&updatedTask, task.ID).Error
	s.Require().NoError(err)
	s.Equal(false, updatedTask.IsActive)

	// Check task history was created
	var history models.TaskHistory
	err = s.DB.Where("task_id = ?", task.ID).First(&history).Error
	s.Require().NoError(err)
	s.Equal(task.ID, history.TaskID)
	s.WithinDuration(*history.CompletedDate, completedDate, time.Second)
	s.WithinDuration(*history.DueDate, dueDate, time.Second)

	// Create a recurring task
	nextDueDate := time.Now().Add(7 * 24 * time.Hour)
	recurringTask := &models.Task{
		Title:       "Weekly Task",
		CreatedBy:   s.testUser.ID,
		NextDueDate: &dueDate,
		IsActive:    true,
		Frequency: models.Frequency{
			Type: models.RepeatWeekly,
		},
	}

	err = s.DB.Create(recurringTask).Error
	s.Require().NoError(err)

	// Complete the recurring task with a new due date
	err = s.repo.CompleteTask(ctx, recurringTask, s.testUser.ID, &nextDueDate, &completedDate)
	s.Require().NoError(err)

	// Check task is still active with new due date
	var updatedRecurringTask models.Task
	err = s.DB.First(&updatedRecurringTask, recurringTask.ID).Error
	s.Require().NoError(err)
	s.Equal(true, updatedRecurringTask.IsActive)
	s.WithinDuration(*updatedRecurringTask.NextDueDate, nextDueDate, time.Second)
}

func (s *TaskTestSuite) TestUncompleteTask() {
	ctx := context.Background()
	dueDate := time.Now().Add(24 * time.Hour)
	completedDate := time.Now()

	task := &models.Task{
		Title:       "Undo Task",
		CreatedBy:   s.testUser.ID,
		NextDueDate: &dueDate,
		IsActive:    true,
		Frequency: models.Frequency{
			Type: models.RepeatOnce,
		},
	}

	err := s.DB.Create(task).Error
	s.Require().NoError(err)

	err = s.repo.CompleteTask(ctx, task, s.testUser.ID, nil, &completedDate)
	s.Require().NoError(err)

	err = s.repo.UncompleteTask(ctx, task.ID)
	s.Require().NoError(err)

	var updatedTask models.Task
	err = s.DB.First(&updatedTask, task.ID).Error
	s.Require().NoError(err)
	s.True(updatedTask.IsActive)
	s.WithinDuration(dueDate, *updatedTask.NextDueDate, time.Second)

	var count int64
	s.DB.Model(&models.TaskHistory{}).Where("task_id = ?", task.ID).Count(&count)
	s.Equal(int64(0), count)
}

func (s *TaskTestSuite) TestGetTaskHistory() {
	ctx := context.Background()
	dueDate := time.Now().Add(-24 * time.Hour) // Due yesterday
	completedDate := time.Now()

	// Create a task
	task := &models.Task{
		Title:       "Test Task",
		CreatedBy:   s.testUser.ID,
		NextDueDate: &dueDate,
		IsActive:    true,
	}

	err := s.DB.Create(task).Error
	s.Require().NoError(err)

	// Create multiple task history records
	histories := []models.TaskHistory{
		{
			TaskID:        task.ID,
			DueDate:       &dueDate,
			CompletedDate: &completedDate,
		},
		{
			TaskID:        task.ID,
			DueDate:       &completedDate, // Using completed date as due date for second entry
			CompletedDate: &completedDate,
		},
	}

	for _, history := range histories {
		err := s.DB.Create(&history).Error
		s.Require().NoError(err)
	}

	// Test retrieval
	retrievedHistories, err := s.repo.GetTaskHistory(ctx, task.ID)
	s.Require().NoError(err)
	s.Require().Len(retrievedHistories, 2)

	// Check sorting by due date desc (most recent first)
	s.WithinDuration(*retrievedHistories[0].DueDate, completedDate, time.Second)
	s.WithinDuration(*retrievedHistories[1].DueDate, dueDate, time.Second)
}

func (s *TaskTestSuite) TestScheduleNextDueDate() {
	now := time.Now()

	testCases := []struct {
		name          string
		task          *models.Task
		completedDate time.Time
		expectedType  string
		expectedDelta time.Duration
	}{
		{
			name: "Once frequency",
			task: &models.Task{
				NextDueDate: &now,
				Frequency: models.Frequency{
					Type: models.RepeatOnce,
				},
			},
			completedDate: now,
			expectedType:  "nil",
		},
		{
			name: "Daily frequency",
			task: &models.Task{
				NextDueDate: &now,
				Frequency: models.Frequency{
					Type: models.RepeatDaily,
				},
			},
			completedDate: now,
			expectedType:  "time",
			expectedDelta: 24 * time.Hour,
		},
		{
			name: "Weekly frequency",
			task: &models.Task{
				NextDueDate: &now,
				Frequency: models.Frequency{
					Type: models.RepeatWeekly,
				},
			},
			completedDate: now,
			expectedType:  "time",
			expectedDelta: 7 * 24 * time.Hour,
		},
		{
			name: "Monthly frequency",
			task: &models.Task{
				NextDueDate: &now,
				Frequency: models.Frequency{
					Type: models.RepeatMonthly,
				},
			},
			completedDate: now,
			expectedType:  "time",
			// Cannot easily assert exact date due to month length variations
		},
		{
			name: "Yearly frequency",
			task: &models.Task{
				NextDueDate: &now,
				Frequency: models.Frequency{
					Type: models.RepeatYearly,
				},
			},
			completedDate: now,
			expectedType:  "time",
			// Cannot easily assert exact date due to leap year variations
		},
		{
			name: "Custom interval - days",
			task: &models.Task{
				NextDueDate: &now,
				Frequency: models.Frequency{
					Type:  models.RepeatCustom,
					On:    models.Interval,
					Every: 3,
					Unit:  models.Days,
				},
			},
			completedDate: now,
			expectedType:  "time",
			expectedDelta: 3 * 24 * time.Hour,
		},
		{
			name: "Rolling task",
			task: &models.Task{
				NextDueDate: &now,
				IsRolling:   true,
				Frequency: models.Frequency{
					Type: models.RepeatWeekly,
				},
			},
			completedDate: now.Add(48 * time.Hour), // Completed 2 days later
			expectedType:  "time",
			expectedDelta: 7 * 24 * time.Hour, // Should be 7 days from completion
		},
		{
			name: "End date restriction",
			task: &models.Task{
				NextDueDate: &now,
				EndDate:     &now, // End date is today
				Frequency: models.Frequency{
					Type: models.RepeatDaily,
				},
			},
			completedDate: now,
			expectedType:  "nil", // Should not calculate a next due date
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			nextDueDate, err := ScheduleNextDueDate(tc.task, tc.completedDate)

			if tc.expectedType == "nil" {
				s.Nil(nextDueDate)
				s.Nil(err)
			} else {
				s.Require().NotNil(nextDueDate)
				s.Require().NoError(err)

				if tc.expectedDelta > 0 {
					var baseDate time.Time
					if tc.task.IsRolling {
						baseDate = tc.completedDate
					} else {
						baseDate = *tc.task.NextDueDate
					}

					expectedDate := baseDate.Add(tc.expectedDelta)
					s.WithinDuration(expectedDate, *nextDueDate, time.Second)
				}
			}
		})
	}
}

func (s *TaskTestSuite) TestGetCompletedTasks() {
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		t := &models.Task{
			Title:     fmt.Sprintf("Task %d", i+1),
			CreatedBy: s.testUser.ID,
			IsActive:  false,
		}
		err := s.DB.Create(t).Error
		s.Require().NoError(err)
		err = s.DB.Model(&models.Task{}).Where("id = ?", t.ID).Update("is_active", false).Error
		s.Require().NoError(err)
	}

	var count int64
	err := s.DB.Model(&models.Task{}).Where("created_by = ? AND is_active = 0", s.testUser.ID).Count(&count).Error
	s.Require().NoError(err)
	s.Require().Equal(int64(5), count)

	anotherUser := &models.User{Email: "other@example.com"}
	err = s.DB.Create(anotherUser).Error
	s.Require().NoError(err)
	otherTask := &models.Task{Title: "Other", CreatedBy: anotherUser.ID, IsActive: false}
	err = s.DB.Create(otherTask).Error
	s.Require().NoError(err)

	tasks, err := s.repo.GetCompletedTasks(ctx, s.testUser.ID, 2, 0)
	s.Require().NoError(err)
	s.Len(tasks, 2)
	s.Equal("Task 5", tasks[0].Title)
	s.Equal("Task 4", tasks[1].Title)

	tasks, err = s.repo.GetCompletedTasks(ctx, s.testUser.ID, 2, 2)
	s.Require().NoError(err)
	s.Len(tasks, 2)
	s.Equal("Task 3", tasks[0].Title)
	s.Equal("Task 2", tasks[1].Title)

	tasks, err = s.repo.GetCompletedTasks(ctx, s.testUser.ID, 2, 10)
	s.Require().NoError(err)
	s.Len(tasks, 0)
}
