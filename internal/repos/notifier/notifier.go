package notifier

import (
	"context"
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
	if err := r.db.Debug().WithContext(c).Model(&models.NotificationSettings{}).First(&settings, userID).Error; err != nil {
		return nil, err
	}
	return &settings, nil
}

func (r *NotificationRepository) DeleteAllTaskNotifications(taskID int) error {
	return r.db.Where("task_id = ?", taskID).Delete(&models.Notification{}).Error
}

func (r *NotificationRepository) BatchInsertNotifications(notifications []*models.Notification) error {
	return r.db.Create(&notifications).Error
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
	if err := r.db.Debug().Where("is_sent = 0 AND scheduled_for < ?", cutoff).Preload("NotificationSettings").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *NotificationRepository) DeleteSentNotifications(c context.Context, since time.Time) error {
	return r.db.WithContext(c).Where("is_sent = 1 AND scheduled_for < ?", since).Delete(&models.Notification{}).Error
}
