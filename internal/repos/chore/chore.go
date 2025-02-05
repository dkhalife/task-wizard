package repos

import (
	"context"
	"fmt"
	"strings"
	"time"

	config "donetick.com/core/config"
	chModel "donetick.com/core/internal/models/chore"
	"gorm.io/gorm"
)

type ChoreRepository struct {
	db *gorm.DB
}

func NewChoreRepository(db *gorm.DB, cfg *config.Config) *ChoreRepository {
	return &ChoreRepository{db: db}
}

func (r *ChoreRepository) UpsertChore(c context.Context, chore *chModel.Chore) error {
	return r.db.WithContext(c).Model(&chore).Save(chore).Error
}

func (r *ChoreRepository) CreateChore(c context.Context, chore *chModel.Chore) (int, error) {
	if err := r.db.WithContext(c).Create(chore).Error; err != nil {
		return 0, err
	}
	return chore.ID, nil
}

func (r *ChoreRepository) GetChore(c context.Context, choreID int) (*chModel.Chore, error) {
	var chore chModel.Chore
	if err := r.db.Debug().WithContext(c).Model(&chModel.Chore{}).Preload("Labels").First(&chore, choreID).Error; err != nil {
		return nil, err
	}
	return &chore, nil
}

func (r *ChoreRepository) GetChores(c context.Context, userID int) ([]*chModel.Chore, error) {
	var chores []*chModel.Chore
	query := r.db.WithContext(c).Preload("Labels").Where("chores.created_by = ?", userID).Group("chores.id").Order("next_due_date asc")

	if err := query.Find(&chores).Error; err != nil {
		return nil, err
	}

	return chores, nil
}

func (r *ChoreRepository) DeleteChore(c context.Context, id int) error {
	r.db.WithContext(c).Where("chore_id = ?", id)
	return r.db.WithContext(c).Delete(&chModel.Chore{}, id).Error
}

func (r *ChoreRepository) IsChoreOwner(c context.Context, choreID int, userID int) error {
	var chore chModel.Chore
	err := r.db.WithContext(c).Model(&chModel.Chore{}).Where("id = ? AND created_by = ?", choreID, userID).First(&chore).Error
	return err
}

func (r *ChoreRepository) CompleteChore(c context.Context, chore *chModel.Chore, userID int, dueDate *time.Time, completedDate *time.Time) error {
	err := r.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		ch := &chModel.ChoreHistory{
			ChoreID:     chore.ID,
			CompletedAt: completedDate,
			DueDate:     chore.NextDueDate,
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
		if err := tx.Model(&chModel.Chore{}).Where("id = ?", chore.ID).Updates(updates).Error; err != nil {
			return err
		}

		return nil
	})
	return err
}

func (r *ChoreRepository) GetChoreHistory(c context.Context, choreID int) ([]*chModel.ChoreHistory, error) {
	var histories []*chModel.ChoreHistory
	if err := r.db.WithContext(c).Where("chore_id = ?", choreID).Order("completed_at desc").Find(&histories).Error; err != nil {
		return nil, err
	}
	return histories, nil
}

func (r *ChoreRepository) GetChoreDetailByID(c context.Context, choreID int) (*chModel.ChoreDetail, error) {
	var choreDetail chModel.ChoreDetail
	if err := r.db.WithContext(c).
		Table("chores").
		Select(`
        chores.id, 
        chores.name, 
        chores.frequency_type, 
        chores.next_due_date, 
        chores.created_by,
        chores.created_by,
		chores.completion_window,
        recent_history.last_completed_date,
		recent_history.notes,
        recent_history.last_assigned_to as last_completed_by,
        COUNT(chore_histories.id) as total_completed`).
		Joins("LEFT JOIN chore_histories ON chores.id = chore_histories.chore_id").
		Joins(`LEFT JOIN (
        SELECT 
            chore_id, 
            assigned_to AS last_assigned_to, 
            completed_at AS last_completed_date,
			notes
			
        FROM chore_histories
        WHERE (chore_id, completed_at) IN (
            SELECT chore_id, MAX(completed_at)
            FROM chore_histories
            GROUP BY chore_id
        )
    ) AS recent_history ON chores.id = recent_history.chore_id`).
		Where("chores.id = ?", choreID).
		Group("chores.id, recent_history.last_completed_date, recent_history.last_assigned_to, recent_history.notes").
		First(&choreDetail).Error; err != nil {
		return nil, err

	}
	return &choreDetail, nil
}

func ScheduleNextDueDate(chore *chModel.Chore, completedDate time.Time) (*time.Time, error) {
	// if Chore is rolling then the next due date calculated from the completed date, otherwise it's calculated from the due date
	var nextDueDate time.Time
	var baseDate time.Time
	var frequencyMetadata chModel.FrequencyMetadata = *chore.FrequencyMetadata

	if chore.FrequencyType == "once" {
		return nil, nil
	}

	if chore.NextDueDate != nil {
		// no due date set, use the current date
		baseDate = chore.NextDueDate.UTC()
	} else {
		baseDate = completedDate.UTC()
	}

	if chore.FrequencyType == "day_of_the_month" || chore.FrequencyType == "days_of_the_week" || chore.FrequencyType == "interval" {
		// time in frequency metadata stored as RFC3339 format like  `2024-07-07T13:27:00-04:00`
		// parse it to time.Time:
		t, err := time.Parse(time.RFC3339, frequencyMetadata.Time)
		if err != nil {
			return nil, fmt.Errorf("error parsing time in frequency metadata")
		}
		// set the time to the time in the frequency metadata:
		baseDate = time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())

	}
	if chore.IsRolling {
		baseDate = completedDate.UTC()
	}
	if chore.FrequencyType == "daily" {
		nextDueDate = baseDate.AddDate(0, 0, 1)
	} else if chore.FrequencyType == "weekly" {
		nextDueDate = baseDate.AddDate(0, 0, 7)
	} else if chore.FrequencyType == "monthly" {
		nextDueDate = baseDate.AddDate(0, 1, 0)
	} else if chore.FrequencyType == "yearly" {
		nextDueDate = baseDate.AddDate(1, 0, 0)
	} else if chore.FrequencyType == "once" {
		// if the chore is a one-time chore, then the next due date is nil
	} else if chore.FrequencyType == "interval" {
		// calculate the difference between the due date and now in days:
		if *frequencyMetadata.Unit == "hours" {
			nextDueDate = baseDate.UTC().Add(time.Hour * time.Duration(chore.Frequency))
		} else if *frequencyMetadata.Unit == "days" {
			nextDueDate = baseDate.UTC().AddDate(0, 0, chore.Frequency)
		} else if *frequencyMetadata.Unit == "weeks" {
			nextDueDate = baseDate.UTC().AddDate(0, 0, chore.Frequency*7)
		} else if *frequencyMetadata.Unit == "months" {
			nextDueDate = baseDate.UTC().AddDate(0, chore.Frequency, 0)
		} else if *frequencyMetadata.Unit == "years" {
			nextDueDate = baseDate.UTC().AddDate(chore.Frequency, 0, 0)
		} else {

			return nil, fmt.Errorf("invalid frequency unit, cannot calculate next due date")
		}
	} else if chore.FrequencyType == "days_of_the_week" {
		//we can only assign to days of the week that part of the frequency metadata.days
		//it's array of days of the week, for example ["monday", "tuesday", "wednesday"]

		// we need to find the next day of the week in the frequency metadata.days that we can schedule
		// if this the last or there is only one. will use same otherwise find the next one:

		// find the index of the chore day in the frequency metadata.days
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
	} else if chore.FrequencyType == "day_of_the_month" {
		for i := 1; i <= 12; i++ {
			nextDueDate = baseDate.AddDate(0, i, 0)
			// set the date to the first day of the month:
			nextDueDate = time.Date(nextDueDate.Year(), nextDueDate.Month(), chore.Frequency, nextDueDate.Hour(), nextDueDate.Minute(), 0, 0, nextDueDate.Location())
			nextMonth := strings.ToLower(nextDueDate.Month().String())
			for _, month := range frequencyMetadata.Months {
				if *month == nextMonth {
					nextDate := nextDueDate.UTC()
					return &nextDate, nil
				}
			}
		}
	} else if chore.FrequencyType == "no_repeat" {
		return nil, nil
	} else if chore.FrequencyType == "trigger" {
		// if the chore is a trigger chore, then the next due date is nil
		return nil, nil
	} else {
		return nil, fmt.Errorf("invalid frequency type, cannot calculate next due date")
	}
	return &nextDueDate, nil

}
