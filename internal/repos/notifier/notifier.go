package notifier

import (
	"context"
	"fmt"
	"time"

	"dkhalife.com/tasks/core/internal/models"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db}
}

func (r *NotificationRepository) GetUserNotificationSettings(c context.Context, userID int) (*models.NotificationSettings, error) {
	var settings models.NotificationSettings
	if err := r.db.WithContext(c).Model(&models.NotificationSettings{}).First(&settings, userID).Error; err != nil {
		return nil, err
	}
	return &settings, nil
}

func (r *NotificationRepository) DeleteAllTaskNotifications(taskID int) error {
	return r.db.Where("task_id = ?", taskID).Delete(&models.Notification{}).Error
}

func (r *NotificationRepository) BatchInsertNotifications(notifications []models.Notification) error {
	return r.db.Create(notifications).Error
}

func (r *NotificationRepository) GenerateNotifications(c context.Context, task *models.Task) {
	r.DeleteAllTaskNotifications(task.ID)

	ns := task.Notification
	if !ns.Enabled {
		return
	}

	if task.NextDueDate == nil {
		return
	}

	notifications := make([]models.Notification, 0)
	if ns.DueDate {
		notifications = append(notifications, models.Notification{
			TaskID:       task.ID,
			UserID:       task.CreatedBy,
			IsSent:       false,
			ScheduledFor: *task.NextDueDate,
			Text:         fmt.Sprintf("📅 *%s* is due today", task.Title),
		})
	}

	if ns.PreDue {
		notifications = append(notifications, models.Notification{
			TaskID:       task.ID,
			UserID:       task.CreatedBy,
			IsSent:       false,
			ScheduledFor: task.NextDueDate.Add(-time.Hour * 3),
			Text:         fmt.Sprintf("📢 *%s* is coming up on %s", task.Title, task.NextDueDate.Format("January 2nd")),
		})
	}

	if len(notifications) > 0 {
		r.BatchInsertNotifications(notifications)
	}
}

func (r *NotificationRepository) MarkNotificationsAsSent(notifications []*models.Notification) error {
	var ids []int
	for _, notification := range notifications {
		ids = append(ids, notification.ID)
	}

	return r.db.Model(&models.Notification{}).Where("id IN (?)", ids).Update("is_sent", true).Error
}

func (r *NotificationRepository) GetPendingNotification(c context.Context, lookback time.Duration) ([]*models.Notification, error) {
	var notifications []*models.Notification
	cutoff := time.Now()
	if err := r.db.Where("is_sent = 0 AND scheduled_for < ?", cutoff).Preload("User.NotificationSettings").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *NotificationRepository) GetOverdueTasksWithNotifications(c context.Context, now time.Time) ([]*models.Task, error) {
	var tasks []*models.Task
	if err := r.db.WithContext(c).Where("is_active = 1 AND next_due_date <= ? AND notification_overdue = 1", now).Select("id, created_by, title").Find(&tasks).Error; err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *NotificationRepository) DeleteSentNotifications(c context.Context, since time.Time) error {
	return r.db.WithContext(c).Where("is_sent = 1 AND scheduled_for < ?", since).Delete(&models.Notification{}).Error
}
