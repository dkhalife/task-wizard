package notifications

import (
	"context"
	"fmt"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/services/logging"

	nRepo "dkhalife.com/tasks/core/internal/repos/notifier"
)

func (n *Notifier) SendNotification(c context.Context, notification *models.Notification) error {
	switch notification.NotificationSettings.Provider.Provider {
	case models.NotificationProviderNone:
		return nil

	case models.NotificationProviderWebhook:
		return SendNotificationViaWebhook(c, notification.NotificationSettings.Provider, notification.Text)

	case models.NotificationProviderGotify:
		return SendNotificationViaGotify(c, notification.NotificationSettings.Provider, notification.Text)

	}

	return nil
}

type Notifier struct {
	DueFrequency time.Duration
	nRepo        *nRepo.NotificationRepository
}

func NewNotifier(cfg *config.Config, nr *nRepo.NotificationRepository) *Notifier {
	return &Notifier{
		DueFrequency: cfg.SchedulerJobs.DueFrequency,
		nRepo:        nr,
	}
}

func (n *Notifier) CleanupSentNotifications(c context.Context) (time.Duration, error) {
	log := logging.FromContext(c)
	startTime := time.Now()
	deleteBefore := time.Now().UTC().Add(-2 * n.DueFrequency)
	err := n.nRepo.DeleteSentNotifications(c, deleteBefore)
	if err != nil {
		log.Error("Error deleting sent notifications", err)
		return time.Since(startTime), err
	}
	return time.Since(startTime), nil
}

func (n *Notifier) LoadAndSendNotificationJob(c context.Context) (time.Duration, error) {
	log := logging.FromContext(c)
	startTime := time.Now()
	pendingNotifications, err := n.nRepo.GetPendingNotification(c, n.DueFrequency)
	log.Debug("Getting pending notifications", " count ", len(pendingNotifications))

	if err != nil {
		log.Error("Error getting pending notifications")
		return time.Since(startTime), err
	}

	for _, notification := range pendingNotifications {
		err := n.SendNotification(c, notification)
		if err != nil {
			log.Error("Error sending notification", err)
			continue
		}
		notification.IsSent = true
	}

	n.nRepo.MarkNotificationsAsSent(pendingNotifications)
	return time.Since(startTime), nil
}

func (n *Notifier) GenerateOverdueNotifications(c context.Context) (time.Duration, error) {
	startTime := time.Now()

	tasks, err := n.nRepo.GetOverdueTasksWithNotifications(c, startTime)

	if err != nil {
		logging.FromContext(c).Error("Error getting overdue tasks", err)
		return time.Since(startTime), err
	}

	if len(tasks) == 0 {
		return time.Since(startTime), nil
	}

	notifications := make([]models.Notification, 0)
	for _, task := range tasks {
		overdueNotification := models.Notification{
			TaskID:       task.ID,
			UserID:       task.CreatedBy,
			IsSent:       false,
			ScheduledFor: startTime,
			Text:         fmt.Sprintf("🚨 *%s* is overdue", task.Title),
		}

		notifications = append(notifications, overdueNotification)
	}

	err = n.nRepo.BatchInsertNotifications(notifications)
	return time.Since(startTime), err
}
