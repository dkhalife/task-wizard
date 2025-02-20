package planner

import (
	"context"
	"fmt"
	"time"

	nModel "dkhalife.com/tasks/core/internal/models/notifier"
	tModel "dkhalife.com/tasks/core/internal/models/task"
	nRepo "dkhalife.com/tasks/core/internal/repos/notifier"
)

type NotificationPlanner struct {
	nRepo *nRepo.NotificationRepository
}

func NewNotificationPlanner(nr *nRepo.NotificationRepository) *NotificationPlanner {
	return &NotificationPlanner{nRepo: nr}
}

func (n *NotificationPlanner) GenerateNotifications(c context.Context, task *tModel.Task) bool {
	n.nRepo.DeleteAllTaskNotifications(task.ID)

	ns := task.Notification
	if !ns.Enabled {
		return true
	}

	if task.NextDueDate == nil {
		return true
	}

	notifications := make([]*nModel.Notification, 0)

	if ns.DueDate {
		notifications = append(notifications, generateDueNotifications(task)...)
	}

	if ns.PreDue {
		notifications = append(notifications, generatePreDueNotifications(task)...)
	}

	if ns.Nag {
		notifications = append(notifications, generateOverdueNotifications(task)...)
	}

	n.nRepo.BatchInsertNotifications(notifications)
	return true
}

func generateDueNotifications(task *tModel.Task) []*nModel.Notification {
	notifications := make([]*nModel.Notification, 0)

	notification := &nModel.Notification{
		TaskID:       task.ID,
		IsSent:       false,
		ScheduledFor: *task.NextDueDate,
		CreatedAt:    time.Now().UTC(),
		//TypeID:       user.NotificationType,
		//UserID:       user.ID,
		Text: fmt.Sprintf("📅 Reminder: *%s* is due today.", task.Title),
	}

	notifications = append(notifications, notification)

	return notifications
}

func generatePreDueNotifications(task *tModel.Task) []*nModel.Notification {
	notifications := make([]*nModel.Notification, 0)

	notification := &nModel.Notification{
		TaskID:       task.ID,
		IsSent:       false,
		ScheduledFor: *task.NextDueDate,
		CreatedAt:    time.Now().UTC().Add(-time.Hour * 3),
		// TypeID:       user.NotificationType,
		// UserID:       user.ID,
		Text: fmt.Sprintf("📢 Heads up! *%s* is due soon (on %s)", task.Title, task.NextDueDate.Format("January 2nd")),
	}

	notifications = append(notifications, notification)

	return notifications
}

func generateOverdueNotifications(task *tModel.Task) []*nModel.Notification {
	notifications := make([]*nModel.Notification, 0)
	for _, hours := range []int{24, 48, 72} {
		scheduleTime := task.NextDueDate.Add(time.Hour * time.Duration(hours))
		notification := &nModel.Notification{
			TaskID:       task.ID,
			IsSent:       false,
			ScheduledFor: scheduleTime,
			CreatedAt:    time.Now().UTC(),
			// TypeID:       user.NotificationType,
			// UserID:       user.ID,
			Text: fmt.Sprintf("🚨 *%s* is now %d hours overdue. Please complete it as soon as possible.", task.Title, hours),
		}

		notifications = append(notifications, notification)
	}

	return notifications
}
