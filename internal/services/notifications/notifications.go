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

func (n *Notifier) sendNotification(c context.Context, notification *models.Notification) error {
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

func (n *Notifier) cleanupInvalidNotifications(c context.Context) error {
	log := logging.FromContext(c)
	log.Debug("Cleaning up notifications for invalid or inactive tasks")

	invalidNotifications, err := n.nRepo.GetNotificationsWithMissingUserOrTask(c)
	if err != nil {
		return fmt.Errorf("error fetching invalid notifications: %s", err.Error())
	}

	if len(invalidNotifications) == 0 {
		return nil
	}

	ids := make([]int, 0, len(invalidNotifications))
	for _, notif := range invalidNotifications {
		ids = append(ids, notif.ID)
	}

	if err := n.nRepo.DeleteNotificationsByIDs(c, ids); err != nil {
		return fmt.Errorf("error deleting invalid notifications: %s", err.Error())
	}

	log.Debugf("Deleted %d invalid/inactive notifications", len(ids))
	return nil
}

func (n *Notifier) cleanupSentNotifications(c context.Context) error {
	log := logging.FromContext(c)
	log.Debug("Cleaning sent notifications")

	deleteBefore := time.Now().UTC().Add(-2 * n.DueFrequency)
	err := n.nRepo.DeleteSentNotifications(c, deleteBefore)
	if err != nil {
		return fmt.Errorf("error deleting sent notifications: %s", err.Error())
	}

	return nil
}

func (n *Notifier) CleanupNotifications(c context.Context) error {
	if err := n.cleanupInvalidNotifications(c); err != nil {
		return fmt.Errorf("error cleaning up invalid notifications: %s", err.Error())
	}

	if err := n.cleanupSentNotifications(c); err != nil {
		return fmt.Errorf("error cleaning up sent notifications: %s", err.Error())
	}

	return nil
}

func (n *Notifier) LoadAndSendNotificationJob(c context.Context) error {
	log := logging.FromContext(c)

	pendingNotifications, err := n.nRepo.GetPendingNotification(c, n.DueFrequency)
	log.Debugf("Getting pending notifications, count=%d", len(pendingNotifications))

	if err != nil {
		return fmt.Errorf("error getting pending notifications: %s", err.Error())
	}

	for _, notification := range pendingNotifications {
		err := n.sendNotification(c, notification)
		if err != nil {
			log.Errorf("Error sending notification: %s", err.Error())
			continue
		}
		notification.IsSent = true
	}

	if err := n.nRepo.MarkNotificationsAsSent(pendingNotifications); err != nil {
		return fmt.Errorf("error marking notifications as sent: %s", err.Error())
	}

	return nil
}

func (n *Notifier) GenerateOverdueNotifications(c context.Context) error {
	log := logging.FromContext(c)
	log.Debug("Generating overdue notifications")

	startTime := time.Now()
	tasks, err := n.nRepo.GetOverdueTasksWithNotifications(c, startTime)

	if err != nil {
		return fmt.Errorf("error getting overdue tasks: %s", err.Error())
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

	return n.nRepo.BatchInsertNotifications(notifications)
}
