package scheduler

import (
	"context"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/services/housekeeper"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/services/notifications"
)

type keyType string

const (
	SchedulerKey keyType = "scheduler"
)

type Scheduler struct {
	stopChan             chan bool
	notifier             *notifications.Notifier
	passwordResetCleaner *housekeeper.PasswordResetCleaner
	config               config.SchedulerConfig
}

func NewScheduler(cfg *config.Config, n *notifications.Notifier, prc *housekeeper.PasswordResetCleaner) *Scheduler {
	return &Scheduler{
		stopChan:             make(chan bool),
		notifier:             n,
		passwordResetCleaner: prc,
		config:               cfg.SchedulerJobs,
	}
}

func (s *Scheduler) Start(c context.Context) {
	log := logging.FromContext(c)
	log.Debug("Scheduler started")

	go s.runScheduler(c, "NOTIFICATION_SCHEDULER", s.notifier.GenerateOverdueNotifications, s.config.OverdueFrequency)
	go s.runScheduler(c, "NOTIFICATION_SENDER", s.notifier.LoadAndSendNotificationJob, s.config.DueFrequency)
	go s.runScheduler(c, "NOTIFICATION_CLEANUP", s.notifier.CleanupSentNotifications, 2*s.config.DueFrequency)
	go s.runScheduler(c, "PASSWORD_RESET_CLEANUP", s.passwordResetCleaner.CleanupStalePasswordResets, s.config.PasswordResetValidity)
}

func (s *Scheduler) runScheduler(c context.Context, jobName string, job func(c context.Context) error, interval time.Duration) {
	log := logging.FromContext(c)
	log.Debugf("%s: [%s] Starting job", time.Now().String(), jobName)

	for {
		select {
		case <-s.stopChan:
			log.Infof("%s: [%s] Stopping job", time.Now().String(), jobName)
			return

		default:
			err := job(c)
			if err != nil {
				log.Errorf("%s: [%s] %s", time.Now().String(), jobName, err)
			}

			time.Sleep(interval)
		}
	}
}

func (s *Scheduler) Stop() {
	s.stopChan <- true
}
