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

	if ns.Overdue {
		notifications = append(notifications, generateOverdueNotifications(task)...)
	}

	n.nRepo.BatchInsertNotifications(notifications)
	return true
}

func generateDueNotifications(task *tModel.Task) []*nModel.Notification {
	notifications := make([]*nModel.Notification, 0)

	notification := &nModel.Notification{
		TaskID:       task.ID,
		UserID:       task.CreatedBy,
		IsSent:       false,
		ScheduledFor: *task.NextDueDate,
		Text:         fmt.Sprintf("📅 *%s* is due today", task.Title),
		CreatedAt:    time.Now().UTC(),
	}

	notifications = append(notifications, notification)

	return notifications
}

func generatePreDueNotifications(task *tModel.Task) []*nModel.Notification {
	notifications := make([]*nModel.Notification, 0)

	notification := &nModel.Notification{
		TaskID:       task.ID,
		UserID:       task.CreatedBy,
		IsSent:       false,
		ScheduledFor: task.NextDueDate.Add(-time.Hour * 3),
		Text:         fmt.Sprintf("📢 *%s* is coming up on %s", task.Title, task.NextDueDate.Format("January 2nd")),
		CreatedAt:    time.Now().UTC(),
	}

	notifications = append(notifications, notification)

	return notifications
}

func generateOverdueNotifications(task *tModel.Task) []*nModel.Notification {
	notifications := make([]*nModel.Notification, 0)
	// TODO: This should be done as part of the scheduler and not prescheduled for a set of days
	for _, hours := range []int{24, 48, 72} {
		scheduleTime := task.NextDueDate.Add(time.Hour * time.Duration(hours))
		notification := &nModel.Notification{
			TaskID:       task.ID,
			UserID:       task.CreatedBy,
			IsSent:       false,
			ScheduledFor: scheduleTime,
			Text:         fmt.Sprintf("🚨 *%s* is overdue", task.Title),
			CreatedAt:    time.Now().UTC(),
		}

		notifications = append(notifications, notification)
	}

	return notifications
}
