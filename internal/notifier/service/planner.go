package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	chModel "donetick.com/core/internal/chore/model"
	nModel "donetick.com/core/internal/notifier/model"
	nRepo "donetick.com/core/internal/notifier/repo"
	"donetick.com/core/logging"
)

type NotificationPlanner struct {
	nRepo *nRepo.NotificationRepository
}

func NewNotificationPlanner(nr *nRepo.NotificationRepository) *NotificationPlanner {
	return &NotificationPlanner{nRepo: nr}
}

func (n *NotificationPlanner) GenerateNotifications(c context.Context, chore *chModel.Chore) bool {
	log := logging.FromContext(c)

	n.nRepo.DeleteAllChoreNotifications(chore.ID)
	notifications := make([]*nModel.Notification, 0)
	if !chore.Notification || chore.FrequencyType == "trigger" {
		return true
	}
	var mt *chModel.NotificationMetadata
	if err := json.Unmarshal([]byte(*chore.NotificationMetadata), &mt); err != nil {
		log.Error("Error unmarshalling notification metadata", err)
		return false
	}
	if chore.NextDueDate == nil {
		return true
	}
	if mt.DueDate {
		notifications = append(notifications, generateDueNotifications(chore)...)
	}
	if mt.PreDue {
		notifications = append(notifications, generatePreDueNotifications(chore)...)
	}
	if mt.Nagging {
		notifications = append(notifications, generateOverdueNotifications(chore)...)
	}

	n.nRepo.BatchInsertNotifications(notifications)
	return true
}

func generateDueNotifications(chore *chModel.Chore) []*nModel.Notification {
	notifications := make([]*nModel.Notification, 0)

	notification := &nModel.Notification{
		ChoreID:      chore.ID,
		IsSent:       false,
		ScheduledFor: *chore.NextDueDate,
		CreatedAt:    time.Now().UTC(),
		//TypeID:       user.NotificationType,
		//UserID:       user.ID,
		Text: fmt.Sprintf("📅 Reminder: *%s* is due today.", chore.Name),
	}
	if notification.IsValid() {
		notifications = append(notifications, notification)
	}

	return notifications
}

func generatePreDueNotifications(chore *chModel.Chore) []*nModel.Notification {
	notifications := make([]*nModel.Notification, 0)

	notification := &nModel.Notification{
		ChoreID:      chore.ID,
		IsSent:       false,
		ScheduledFor: *chore.NextDueDate,
		CreatedAt:    time.Now().UTC().Add(-time.Hour * 3),
		// TypeID:       user.NotificationType,
		// UserID:       user.ID,
		Text: fmt.Sprintf("📢 Heads up! *%s* is due soon (on %s)", chore.Name, chore.NextDueDate.Format("January 2nd")),
	}
	if notification.IsValid() {
		notifications = append(notifications, notification)
	}

	return notifications
}

func generateOverdueNotifications(chore *chModel.Chore) []*nModel.Notification {
	notifications := make([]*nModel.Notification, 0)
	for _, hours := range []int{24, 48, 72} {
		scheduleTime := chore.NextDueDate.Add(time.Hour * time.Duration(hours))
		notification := &nModel.Notification{
			ChoreID:      chore.ID,
			IsSent:       false,
			ScheduledFor: scheduleTime,
			CreatedAt:    time.Now().UTC(),
			// TypeID:       user.NotificationType,
			// UserID:       user.ID,
			Text: fmt.Sprintf("🚨 *%s* is now %d hours overdue. Please complete it as soon as possible.", chore.Name, hours),
		}
		if notification.IsValid() {
			notifications = append(notifications, notification)
		}
	}

	return notifications
}
