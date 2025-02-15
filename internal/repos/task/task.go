package repos

import (
	"context"
	"fmt"
	"strings"
	"time"

	config "donetick.com/core/config"
	tModel "donetick.com/core/internal/models/task"
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
	if err := r.db.Debug().WithContext(c).Model(&tModel.Task{}).First(&task, taskID).Error; err != nil {
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
			TaskID:      task.ID,
			CompletedAt: completedDate,
			DueDate:     task.NextDueDate,
		}
		if err := tx.Create(ch).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{}
		updates["next_due_date"] = dueDate

		if dueDate == nil {
			updates["is_active"] = false
		}
		// Perform the update operation once, using the prepared updates map.
		if err := tx.Model(&tModel.Task{}).Where("id = ?", task.ID).Updates(updates).Error; err != nil {
			return err
		}

		return nil
	})
	return err
}

func (r *TaskRepository) GetTaskHistory(c context.Context, taskID int) ([]*tModel.TaskHistory, error) {
	var histories []*tModel.TaskHistory
	if err := r.db.WithContext(c).Where("task_id = ?", taskID).Order("completed_at desc").Find(&histories).Error; err != nil {
		return nil, err
	}
	return histories, nil
}

func ScheduleNextDueDate(task *tModel.Task, completedDate time.Time) (*time.Time, error) {
	// if task is rolling then the next due date calculated from the completed date, otherwise it's calculated from the due date
	var nextDueDate time.Time
	var baseDate time.Time
	// TODO: Utility to deserialize from task.FrequencyMetadata
	var frequencyMetadata *tModel.FrequencyMetadata

	if task.FrequencyType == "once" {
		return nil, nil
	}

	if task.NextDueDate != nil {
		// no due date set, use the current date
		baseDate = task.NextDueDate.UTC()
	} else {
		baseDate = completedDate.UTC()
	}

	if task.FrequencyType == "day_of_the_month" || task.FrequencyType == "days_of_the_week" || task.FrequencyType == "interval" {
		// time in frequency metadata stored as RFC3339 format like  `2024-07-07T13:27:00-04:00`
		// parse it to time.Time:
		t, err := time.Parse(time.RFC3339, frequencyMetadata.Time)
		if err != nil {
			return nil, fmt.Errorf("error parsing time in frequency metadata")
		}
		// set the time to the time in the frequency metadata:
		baseDate = time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())

	}
	if task.IsRolling {
		baseDate = completedDate.UTC()
	}
	if task.FrequencyType == "daily" {
		nextDueDate = baseDate.AddDate(0, 0, 1)
	} else if task.FrequencyType == "weekly" {
		nextDueDate = baseDate.AddDate(0, 0, 7)
	} else if task.FrequencyType == "monthly" {
		nextDueDate = baseDate.AddDate(0, 1, 0)
	} else if task.FrequencyType == "yearly" {
		nextDueDate = baseDate.AddDate(1, 0, 0)
	} else if task.FrequencyType == "once" {
		// if the task is a one-time task, then the next due date is nil
	} else if task.FrequencyType == "interval" {
		// calculate the difference between the due date and now in days:
		if *frequencyMetadata.Unit == "hours" {
			nextDueDate = baseDate.UTC().Add(time.Hour * time.Duration(task.Frequency))
		} else if *frequencyMetadata.Unit == "days" {
			nextDueDate = baseDate.UTC().AddDate(0, 0, task.Frequency)
		} else if *frequencyMetadata.Unit == "weeks" {
			nextDueDate = baseDate.UTC().AddDate(0, 0, task.Frequency*7)
		} else if *frequencyMetadata.Unit == "months" {
			nextDueDate = baseDate.UTC().AddDate(0, task.Frequency, 0)
		} else if *frequencyMetadata.Unit == "years" {
			nextDueDate = baseDate.UTC().AddDate(task.Frequency, 0, 0)
		} else {

			return nil, fmt.Errorf("invalid frequency unit, cannot calculate next due date")
		}
	} else if task.FrequencyType == "days_of_the_week" {
		//we can only assign to days of the week that part of the frequency metadata.days
		//it's array of days of the week, for example ["monday", "tuesday", "wednesday"]

		// we need to find the next day of the week in the frequency metadata.days that we can schedule
		// if this the last or there is only one. will use same otherwise find the next one:

		// find the index of the task day in the frequency metadata.days
		// loop for next 7 days from the base, if the day in the frequency metadata.days then we can schedule it:
		for i := 1; i <= 7; i++ {
			nextDueDate = baseDate.AddDate(0, 0, i)
			nextDay := strings.ToLower(nextDueDate.Weekday().String())
			for _, day := range frequencyMetadata.Days {
				if strings.ToLower(*day) == nextDay {
					nextDate := nextDueDate.UTC()
					return &nextDate, nil
				}
			}
		}
	} else if task.FrequencyType == "day_of_the_month" {
		for i := 1; i <= 12; i++ {
			nextDueDate = baseDate.AddDate(0, i, 0)
			// set the date to the first day of the month:
			nextDueDate = time.Date(nextDueDate.Year(), nextDueDate.Month(), task.Frequency, nextDueDate.Hour(), nextDueDate.Minute(), 0, 0, nextDueDate.Location())
			nextMonth := strings.ToLower(nextDueDate.Month().String())
			for _, month := range frequencyMetadata.Months {
				if *month == nextMonth {
					nextDate := nextDueDate.UTC()
					return &nextDate, nil
				}
			}
		}
	} else if task.FrequencyType == "no_repeat" {
		return nil, nil
	} else if task.FrequencyType == "trigger" {
		// if the task is a trigger task, then the next due date is nil
		return nil, nil
	} else {
		return nil, fmt.Errorf("invalid frequency type, cannot calculate next due date")
	}
	return &nextDueDate, nil

}
