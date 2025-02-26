package notifications

import (
	"context"
	"fmt"
	"log"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/models"
	nRepo "dkhalife.com/tasks/core/internal/repos/notifier"
	tRepo "dkhalife.com/tasks/core/internal/repos/task"
	uRepo "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/services/logging"
)

type keyType string

const (
	SchedulerKey keyType = "scheduler"
)

type Notifier struct {
}

func NewNotifier() *Notifier {
	return &Notifier{}
}

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

type Scheduler struct {
	taskRepo         *tRepo.TaskRepository
	userRepo         *uRepo.UserRepository
	stopChan         chan bool
	notifier         *Notifier
	notificationRepo *nRepo.NotificationRepository
	config           config.SchedulerConfig
}

func NewScheduler(cfg *config.Config, ur *uRepo.UserRepository, cr *tRepo.TaskRepository, n *Notifier, nr *nRepo.NotificationRepository) *Scheduler {
	return &Scheduler{
		taskRepo:         cr,
		userRepo:         ur,
		stopChan:         make(chan bool),
		notifier:         n,
		notificationRepo: nr,
		config:           cfg.SchedulerJobs,
	}
}

func (s *Scheduler) Start(c context.Context) {
	log := logging.FromContext(c)
	log.Debug("Scheduler started")
	go s.runScheduler(c, " NOTIFICATION_SCHEDULER ", s.generateOverdueNotifications, s.config.OverdueFrequency)
	go s.runScheduler(c, " NOTIFICATION_SENDER ", s.loadAndSendNotificationJob, s.config.DueFrequency)
	go s.runScheduler(c, " NOTIFICATION_CLEANUP ", s.cleanupSentNotifications, 2*s.config.DueFrequency)
}

func (s *Scheduler) cleanupSentNotifications(c context.Context) (time.Duration, error) {
	log := logging.FromContext(c)
	startTime := time.Now()
	deleteBefore := time.Now().UTC().Add(-2 * s.config.DueFrequency)
	err := s.notificationRepo.DeleteSentNotifications(c, deleteBefore)
	if err != nil {
		log.Error("Error deleting sent notifications", err)
		return time.Since(startTime), err
	}
	return time.Since(startTime), nil
}

func (s *Scheduler) loadAndSendNotificationJob(c context.Context) (time.Duration, error) {
	log := logging.FromContext(c)
	startTime := time.Now()
	pendingNotifications, err := s.notificationRepo.GetPendingNotification(c, s.config.DueFrequency)
	log.Debug("Getting pending notifications", " count ", len(pendingNotifications))

	if err != nil {
		log.Error("Error getting pending notifications")
		return time.Since(startTime), err
	}

	for _, notification := range pendingNotifications {
		err := s.notifier.SendNotification(c, notification)
		if err != nil {
			log.Error("Error sending notification", err)
			continue
		}
		notification.IsSent = true
	}

	s.notificationRepo.MarkNotificationsAsSent(pendingNotifications)
	return time.Since(startTime), nil
}

func (s *Scheduler) generateOverdueNotifications(c context.Context) (time.Duration, error) {
	startTime := time.Now()

	tasks, err := s.taskRepo.GetOverdueTasksWithNotifications(c, startTime)

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

	err = s.notificationRepo.BatchInsertNotifications(notifications)
	return time.Since(startTime), err
}

func (s *Scheduler) runScheduler(c context.Context, jobName string, job func(c context.Context) (time.Duration, error), interval time.Duration) {
	for {
		logging.FromContext(c).Debug("Scheduler running ", jobName, " time", time.Now().String())

		select {
		case <-s.stopChan:
			log.Println("Scheduler stopped")
			return
		default:
			elapsedTime, err := job(c)
			if err != nil {
				logging.FromContext(c).Error("Error running scheduler job", err)
			}
			logging.FromContext(c).Debug("Scheduler job completed", jobName, " time: ", elapsedTime.String())
		}
		time.Sleep(interval)
	}
}

func (s *Scheduler) Stop() {
	s.stopChan <- true
}
