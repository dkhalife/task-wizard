package scheduler

import (
	"context"
	"time"

	"dkhalife.com/tasks/core/config"
	sRepo "dkhalife.com/tasks/core/internal/repos/session"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/services/notifications"
	"dkhalife.com/tasks/core/internal/services/users"
	"dkhalife.com/tasks/core/internal/telemetry"
)

type Scheduler struct {
	stopChan    chan bool
	notifier    *notifications.Notifier
	userService *users.UserService
	sessionRepo sRepo.ISessionRepo
	config      config.SchedulerConfig
}

func NewScheduler(cfg *config.Config, n *notifications.Notifier, us *users.UserService, sr sRepo.ISessionRepo) *Scheduler {
	return &Scheduler{
		stopChan:    make(chan bool),
		notifier:    n,
		userService: us,
		sessionRepo: sr,
		config:      cfg.SchedulerJobs,
	}
}

func (s *Scheduler) Start(c context.Context) {
	log := logging.FromContext(c)
	log.Info("Scheduler started")

	go s.runScheduler(c, "NOTIFICATION_SCHEDULER", s.notifier.GenerateOverdueNotifications, s.config.OverdueFrequency)
	go s.runScheduler(c, "NOTIFICATION_SENDER", s.notifier.LoadAndSendNotificationJob, s.config.DueFrequency)
	go s.runScheduler(c, "NOTIFICATION_CLEANUP", s.notifier.CleanupNotifications, s.config.NotificationCleanup)
	go s.runScheduler(c, "ACCOUNT_DELETION", s.userService.ProcessDeletions, s.config.AccountDeletionFrequency)
	go s.runScheduler(c, "SESSION_CLEANUP", s.sessionRepo.CleanupExpired, 1*time.Hour)
}

func (s *Scheduler) runScheduler(c context.Context, jobName string, job func(c context.Context) error, interval time.Duration) {
	log := logging.FromContext(c)
	log.Infof("[%s] Starting job", jobName)

	for {
		select {
		case <-s.stopChan:
			log.Infof("[%s] Stopping job", jobName)
			return

		default:
			err := job(c)
			if err != nil {
				log.Errorf("[%s] %s", jobName, err)
				telemetry.TrackError(c, "scheduler_job_failed", "scheduler", err, map[string]string{"job": jobName})
			}

			time.Sleep(interval)
		}
	}
}

func (s *Scheduler) Stop() {
	s.stopChan <- true
}
