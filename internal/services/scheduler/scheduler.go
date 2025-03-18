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
	appTokenCleaner      *housekeeper.AppTokenCleaner
	config               config.SchedulerConfig
}

func NewScheduler(cfg *config.Config, n *notifications.Notifier, prc *housekeeper.PasswordResetCleaner, atk *housekeeper.AppTokenCleaner) *Scheduler {
	return &Scheduler{
		stopChan:             make(chan bool),
		notifier:             n,
		passwordResetCleaner: prc,
		appTokenCleaner:      atk,
		config:               cfg.SchedulerJobs,
	}
}

func (s *Scheduler) Start(c context.Context) {
	log := logging.FromContext(c)
	log.Info("Scheduler started")

	go s.runScheduler(c, "NOTIFICATION_SCHEDULER", s.notifier.GenerateOverdueNotifications, s.config.OverdueFrequency)
	go s.runScheduler(c, "NOTIFICATION_SENDER", s.notifier.LoadAndSendNotificationJob, s.config.DueFrequency)
	go s.runScheduler(c, "NOTIFICATION_CLEANUP", s.notifier.CleanupSentNotifications, 2*s.config.DueFrequency)
	go s.runScheduler(c, "PASSWORD_RESET_CLEANUP", s.passwordResetCleaner.CleanupStalePasswordResets, s.config.PasswordResetValidity)
	go s.runScheduler(c, "TOKEN_EXPIRATION_REMINDER", s.appTokenCleaner.SendTokenExpirationReminder, s.config.TokenExpirationReminder)
	go s.runScheduler(c, "TOKEN_EXPIRATION_CLEANUP", s.appTokenCleaner.CleanupExpiredTokens, time.Duration(24)*time.Hour)
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
			}

			time.Sleep(interval)
		}
	}
}

func (s *Scheduler) Stop() {
	s.stopChan <- true
}
