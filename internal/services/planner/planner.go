package planner

import (
	"context"
	"fmt"
	"time"

	"dkhalife.com/tasks/core/internal/models"
	nRepo "dkhalife.com/tasks/core/internal/repos/notifier"
)

type NotificationPlanner struct {
	nRepo *nRepo.NotificationRepository
}

func NewNotificationPlanner(nr *nRepo.NotificationRepository) *NotificationPlanner {
	return &NotificationPlanner{nRepo: nr}
}

func (n *NotificationPlanner) GenerateNotifications(c context.Context, task *models.Task) {
	n.nRepo.DeleteAllTaskNotifications(task.ID)

	ns := task.Notification
	if !ns.Enabled {
		return
	}

	if task.NextDueDate == nil {
		return
	}

	notifications := make([]models.Notification, 2)

	if ns.DueDate {
		notifications = append(notifications, models.Notification{
			TaskID:       task.ID,
			UserID:       task.CreatedBy,
			IsSent:       false,
			ScheduledFor: *task.NextDueDate,
			Text:         fmt.Sprintf("📅 *%s* is due today", task.Title),
		})
	}

	if ns.PreDue {
		notifications = append(notifications, models.Notification{
			TaskID:       task.ID,
			UserID:       task.CreatedBy,
			IsSent:       false,
			ScheduledFor: task.NextDueDate.Add(-time.Hour * 3),
			Text:         fmt.Sprintf("📢 *%s* is coming up on %s", task.Title, task.NextDueDate.Format("January 2nd")),
		})
	}

	n.nRepo.BatchInsertNotifications(notifications)
}
