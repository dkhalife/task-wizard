package repos

import (
	"context"
	"errors"
	"time"

	config "dkhalife.com/tasks/core/config"
	tModel "dkhalife.com/tasks/core/internal/models/task"
	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB, cfg *config.Config) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) UpsertTask(c context.Context, task *tModel.Task) error {
	return r.db.WithContext(c).Model(&task).Save(task).Error
}

func (r *TaskRepository) CreateTask(c context.Context, task *tModel.Task) (int, error) {
	if err := r.db.WithContext(c).Create(task).Error; err != nil {
		return 0, err
	}
	return task.ID, nil
}

func (r *TaskRepository) GetTask(c context.Context, taskID int) (*tModel.Task, error) {
	var task tModel.Task
	if err := r.db.WithContext(c).Model(&tModel.Task{}).First(&task, taskID).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) GetTasks(c context.Context, userID int) ([]*tModel.Task, error) {
	var tasks []*tModel.Task
	query := r.db.WithContext(c).Where("tasks.created_by = ? AND is_active = 1", userID).Group("tasks.id").Order("next_due_date asc")

	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *TaskRepository) DeleteTask(c context.Context, id int) error {
	r.db.WithContext(c).Where("task_id = ?", id)
	return r.db.WithContext(c).Delete(&tModel.Task{}, id).Error
}

func (r *TaskRepository) IsTaskOwner(c context.Context, taskID int, userID int) error {
	var task tModel.Task
	err := r.db.WithContext(c).Model(&tModel.Task{}).Where("id = ? AND created_by = ?", taskID, userID).First(&task).Error
	return err
}

func (r *TaskRepository) CompleteTask(c context.Context, task *tModel.Task, userID int, dueDate *time.Time, completedDate *time.Time) error {
	err := r.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		ch := &tModel.TaskHistory{
			TaskID:        task.ID,
			CompletedDate: completedDate,
			DueDate:       task.NextDueDate,
		}
		if err := tx.Create(ch).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{}
		updates["next_due_date"] = dueDate

		if dueDate == nil {
			updates["is_active"] = false
		}

		if err := tx.Model(&tModel.Task{}).Where("id = ?", task.ID).Updates(updates).Error; err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *TaskRepository) GetTaskHistory(c context.Context, taskID int) ([]*tModel.TaskHistory, error) {
	var histories []*tModel.TaskHistory
	if err := r.db.WithContext(c).Where("task_id = ?", taskID).Order("completed_date desc").Find(&histories).Error; err != nil {
		return nil, err
	}
	return histories, nil
}

func ScheduleNextDueDate(task *tModel.Task, completedDate time.Time) (*time.Time, error) {
	var freq = task.Frequency
	if freq.Type == "once" {
		return nil, nil
	}

	var baseDate time.Time = *task.NextDueDate
	if task.IsRolling {
		baseDate = completedDate
	}

	if baseDate.IsZero() {
		return nil, errors.New("unable to calculate next due date")
	}

	var nextDueDate time.Time
	if freq.Type == "daily" {
		nextDueDate = baseDate.AddDate(0, 0, 1)
	} else if freq.Type == "weekly" {
		nextDueDate = baseDate.AddDate(0, 0, 7)
	} else if freq.Type == "monthly" {
		nextDueDate = baseDate.AddDate(0, 1, 0)
	} else if freq.Type == "yearly" {
		nextDueDate = baseDate.AddDate(1, 0, 0)
	} else if freq.Type == "custom" {
		if freq.On == "interval" {
			if freq.Unit == "hours" {
				nextDueDate = baseDate.Add(time.Duration(freq.Every) * time.Hour)
			} else if freq.Unit == "days" {
				nextDueDate = baseDate.AddDate(0, 0, freq.Every)
			} else if freq.Unit == "weeks" {
				nextDueDate = baseDate.AddDate(0, 0, 7*freq.Every)
			} else if freq.Unit == "months" {
				nextDueDate = baseDate.AddDate(0, freq.Every, 0)
			} else if freq.Unit == "years" {
				nextDueDate = baseDate.AddDate(freq.Every, 0, 0)
			}
		} else if freq.On == "days_of_the_week" {
			currentWeekDay := int32(baseDate.Weekday())
			days := freq.Days

			if len(days) == 0 {
				return nil, errors.New("days of the week cannot be empty")
			}

			duringThisWeek := false
			for _, day := range days {
				if day > currentWeekDay {
					duringThisWeek = true
					nextDueDate = baseDate.AddDate(0, 0, int(day-currentWeekDay))
					break
				}
			}

			if !duringThisWeek {
				daysUntilNextWeek := 7 - int(currentWeekDay)
				nextDueDate = baseDate.AddDate(0, 0, daysUntilNextWeek+int(days[0]))
			}
		} else if freq.On == "day_of_the_months" {
			currentMonth := int32(baseDate.Month())
			months := freq.Months

			if len(months) == 0 {
				return nil, errors.New("months cannot be empty")
			}

			duringThisYear := false
			for _, month := range months {
				if month > currentMonth {
					duringThisYear = true
					nextDueDate = baseDate.AddDate(0, int(month-currentMonth), 0)
					break
				}
			}

			if !duringThisYear {
				monthsUntilNextYear := 12 - int(currentMonth)
				nextDueDate = baseDate.AddDate(0, monthsUntilNextYear+int(months[0]), 0)
			}
		}
	}

	return &nextDueDate, nil
}
