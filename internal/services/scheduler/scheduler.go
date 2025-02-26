package scheduler

import (
	"context"
	"log"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/services/notifications"
)

type keyType string

const (
	SchedulerKey keyType = "scheduler"
)

type Scheduler struct {
	stopChan chan bool
	notifier *notifications.Notifier
	config   config.SchedulerConfig
}

func NewScheduler(cfg *config.Config, n *notifications.Notifier) *Scheduler {
	return &Scheduler{
		stopChan: make(chan bool),
		notifier: n,
		config:   cfg.SchedulerJobs,
	}
}

func (s *Scheduler) Start(c context.Context) {
	log := logging.FromContext(c)
	log.Debug("Scheduler started")

	go s.runScheduler(c, " NOTIFICATION_SCHEDULER ", s.notifier.GenerateOverdueNotifications, s.config.OverdueFrequency)
	go s.runScheduler(c, " NOTIFICATION_SENDER ", s.notifier.LoadAndSendNotificationJob, s.config.DueFrequency)
	go s.runScheduler(c, " NOTIFICATION_CLEANUP ", s.notifier.CleanupSentNotifications, 2*s.config.DueFrequency)
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
		select {
		case <-s.stopChan:
			log.Println("Scheduler stopped")
			return
		case <-time.After(interval):
		}
	}
}

func (s *Scheduler) Stop() {
	s.stopChan <- true
}
