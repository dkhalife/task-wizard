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
	switch notification.User.NotificationSettings.Provider.Provider {
	case models.NotificationProviderNone:
		return nil

	case models.NotificationProviderWebhook:
		return SendNotificationViaWebhook(c, notification.User.NotificationSettings.Provider, notification.Text)

	case models.NotificationProviderGotify:
		return SendNotificationViaGotify(c, notification.User.NotificationSettings.Provider, notification.Text)

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

func (n *Notifier) CleanupSentNotifications(c context.Context) error {
	log := logging.FromContext(c)
	deleteBefore := time.Now().UTC().Add(-2 * n.DueFrequency)

	err := n.nRepo.DeleteSentNotifications(c, deleteBefore)
	if err != nil {
		log.Error("Error deleting sent notifications", err)
		return err
	}

	return nil
}

func (n *Notifier) LoadAndSendNotificationJob(c context.Context) error {
	log := logging.FromContext(c)

	pendingNotifications, err := n.nRepo.GetPendingNotification(c, n.DueFrequency)
	log.Debug("Getting pending notifications", " count ", len(pendingNotifications))

	if err != nil {
		log.Error("Error getting pending notifications")
		return err
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
	return nil
}

func (n *Notifier) GenerateOverdueNotifications(c context.Context) error {
	startTime := time.Now()
	tasks, err := n.nRepo.GetOverdueTasksWithNotifications(c, startTime)

	if err != nil {
		logging.FromContext(c).Error("Error getting overdue tasks", err)
		return err
	}

	if len(tasks) == 0 {
		return nil
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
	return err
}
