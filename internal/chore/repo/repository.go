package chore

import (
	"context"
	"time"

	config "donetick.com/core/config"
	chModel "donetick.com/core/internal/chore/model"
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
