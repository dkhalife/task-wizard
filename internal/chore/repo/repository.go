package chore

import (
	"context"
	"fmt"
	"time"

	config "donetick.com/core/config"
	chModel "donetick.com/core/internal/chore/model"
	"gorm.io/gorm"
)

type ChoreRepository struct {
	db     *gorm.DB
	dbType string
}

func NewChoreRepository(db *gorm.DB, cfg *config.Config) *ChoreRepository {
	return &ChoreRepository{db: db, dbType: cfg.Database.Type}
}

func (r *ChoreRepository) UpsertChore(c context.Context, chore *chModel.Chore) error {
	return r.db.WithContext(c).Model(&chore).Save(chore).Error
}

func (r *ChoreRepository) UpdateChores(c context.Context, chores []*chModel.Chore) error {
	return r.db.WithContext(c).Save(&chores).Error
}

func (r *ChoreRepository) CreateChore(c context.Context, chore *chModel.Chore) (int, error) {
	if err := r.db.WithContext(c).Create(chore).Error; err != nil {
		return 0, err
	}
	return chore.ID, nil
}

func (r *ChoreRepository) GetChore(c context.Context, choreID int) (*chModel.Chore, error) {
	var chore chModel.Chore
	if err := r.db.Debug().WithContext(c).Model(&chModel.Chore{}).Preload("LabelsV2").First(&chore, choreID).Error; err != nil {
		return nil, err
	}
	return &chore, nil
}

func (r *ChoreRepository) GetChores(c context.Context, userID int, includeArchived bool) ([]*chModel.Chore, error) {
	var chores []*chModel.Chore
	query := r.db.WithContext(c).Preload("LabelsV2").Where("chores.created_by = ?", userID).Group("chores.id").Order("next_due_date asc")
	if !includeArchived {
		query = query.Where("chores.is_active = ?", true)
	}
	return chores, nil
}

func (r *ChoreRepository) GetArchivedChores(c context.Context, userID int) ([]*chModel.Chore, error) {
	var chores []*chModel.Chore
	if err := r.db.WithContext(c).Preload("LabelsV2").Where("chores.created_by = ?", userID).Group("chores.id").Order("next_due_date asc").Find(&chores, "is_active = ?", false).Error; err != nil {
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
			CompletedBy: userID,
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
func (r *ChoreRepository) GetChoreHistoryWithLimit(c context.Context, choreID int, limit int) ([]*chModel.ChoreHistory, error) {
	var histories []*chModel.ChoreHistory
	if err := r.db.WithContext(c).Where("chore_id = ?", choreID).Order("completed_at desc").Limit(limit).Find(&histories).Error; err != nil {
		return nil, err
	}
	return histories, nil
}

func (r *ChoreRepository) GetChoreHistoryByID(c context.Context, choreID int, historyID int) (*chModel.ChoreHistory, error) {
	var history chModel.ChoreHistory
	if err := r.db.WithContext(c).Where("id = ? and chore_id = ? ", historyID, choreID).First(&history).Error; err != nil {
		return nil, err
	}
	return &history, nil
}

func (r *ChoreRepository) UpdateChoreHistory(c context.Context, history *chModel.ChoreHistory) error {
	return r.db.WithContext(c).Save(history).Error
}

func (r *ChoreRepository) DeleteChoreHistory(c context.Context, historyID int) error {
	return r.db.WithContext(c).Delete(&chModel.ChoreHistory{}, historyID).Error
}

func (r *ChoreRepository) GetChoresForNotification(c context.Context) ([]*chModel.Chore, error) {
	var chores []*chModel.Chore
	query := r.db.WithContext(c).Table("chores").Joins("left join notifications n on n.chore_id = chores.id and n.scheduled_for = chores.next_due_date and n.type = 1")
	if err := query.Where("chores.is_active = ? and chores.notification = ? and n.id is null", true, true).Find(&chores).Error; err != nil {
		return nil, err
	}
	return chores, nil
}

func (r *ChoreRepository) GetOverdueChoresForNotification(c context.Context, overdueFor time.Duration, everyDuration time.Duration, untilDuration time.Duration) ([]*chModel.Chore, error) {
	var chores []*chModel.Chore
	now := time.Now().UTC()
	overdueTime := now.Add(-overdueFor)
	everyTime := now.Add(-everyDuration)
	untilTime := now.Add(-untilDuration)

	query := r.db.Debug().WithContext(c).
		Table("chores").
		Select("chores.*, MAX(n.created_at) as max_notification_created_at").
		Joins("left join notifications n on n.chore_id = chores.id and n.type = 2").
		Where("chores.is_active = ? AND chores.notification = ? AND chores.next_due_date < ? AND chores.next_due_date > ?", true, true, overdueTime, untilTime).
		Where(readJSONBooleanField(r.dbType, "chores.notification_meta", "nagging")).
		Group("chores.id").
		Having("MAX(n.created_at) IS NULL OR MAX(n.created_at) < ?", everyTime)

	if err := query.Find(&chores).Error; err != nil {
		return nil, err
	}

	return chores, nil
}

// a predue notfication is a notification send before the due date in 6 hours, 3 hours :
func (r *ChoreRepository) GetPreDueChoresForNotification(c context.Context, preDueDuration time.Duration, everyDuration time.Duration) ([]*chModel.Chore, error) {
	var chores []*chModel.Chore
	query := r.db.WithContext(c).Table("chores").Select("chores.*, MAX(n.created_at) as max_notification_created_at").Joins("left join notifications n on n.chore_id = chores.id and n.scheduled_for = chores.next_due_date and n.type = 3")
	if err := query.Where("chores.is_active = ? and chores.notification = ? and chores.next_due_date > ? and chores.next_due_date < ?", true, true, time.Now().UTC(), time.Now().Add(everyDuration*2).UTC()).Where(readJSONBooleanField(r.dbType, "chores.notification_meta", "predue")).Having("MAX(n.created_at) is null or MAX(n.created_at) < ?", time.Now().Add(everyDuration).UTC()).Group("chores.id").Find(&chores).Error; err != nil {
		return nil, err
	}
	return chores, nil
}

func readJSONBooleanField(dbType string, columnName string, fieldName string) string {
	if dbType == "postgres" {
		return fmt.Sprintf("(%s::json->>'%s')::boolean", columnName, fieldName)
	}
	return fmt.Sprintf("JSON_EXTRACT(%s, '$.%s')", columnName, fieldName)
}

func (r *ChoreRepository) SetDueDate(c context.Context, choreID int, dueDate time.Time) error {
	return r.db.WithContext(c).Model(&chModel.Chore{}).Where("id = ?", choreID).Update("next_due_date", dueDate).Error
}

func (r *ChoreRepository) SetDueDateIfNotExisted(c context.Context, choreID int, dueDate time.Time) error {
	return r.db.WithContext(c).Model(&chModel.Chore{}).Where("id = ? and next_due_date is null", choreID).Update("next_due_date", dueDate).Error
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

func (r *ChoreRepository) ArchiveChore(c context.Context, choreID int, userID int) error {
	return r.db.WithContext(c).Model(&chModel.Chore{}).Where("id = ? and created_by = ?", choreID, userID).Update("is_active", false).Error
}

func (r *ChoreRepository) UnarchiveChore(c context.Context, choreID int, userID int) error {
	return r.db.WithContext(c).Model(&chModel.Chore{}).Where("id = ? and created_by = ?", choreID, userID).Update("is_active", true).Error
}

func (r *ChoreRepository) GetChoresHistoryByUserID(c context.Context, userID int, days int) ([]*chModel.ChoreHistory, error) {
	var chores []*chModel.ChoreHistory
	since := time.Now().AddDate(0, 0, days*-1)
	if err := r.db.WithContext(c).Where("completed_by = ? AND completed_at > ?", userID, since).Order("completed_at desc").Find(&chores).Error; err != nil {
		return nil, err
	}
	return chores, nil
}
