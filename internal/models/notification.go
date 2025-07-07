package models

import (
	"time"
)

type Notification struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	TaskID       int       `json:"task_id" gorm:"column:task_id;not null;index:idx_notifications_task_id"`
	UserID       int       `json:"user_id" gorm:"column:user_id;not null;index:idx_notifications_user_id"`
	Text         string    `json:"text" gorm:"column:text;not null"`
	IsSent       bool      `json:"is_sent" gorm:"column:is_sent;index;default:false"`
	ScheduledFor time.Time `json:"scheduled_for" gorm:"column:scheduled_for;not null;index"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`

	Task Task `json:"-" gorm:"foreignKey:TaskID"`
	User User `json:"-" gorm:"foreignKey:UserID"`
}

type NotificationProviderType string

const (
	NotificationProviderNone    NotificationProviderType = "none"
	NotificationProviderWebhook NotificationProviderType = "webhook"
	NotificationProviderGotify  NotificationProviderType = "gotify"
)

type NotificationProvider struct {
	Provider NotificationProviderType `json:"provider" validate:"required" gorm:"type:varchar(7);column:type"`
	URL      string                   `json:"url" validate:"required_if=Provider webhook gotify" gorm:"column:url"`
	Method   string                   `json:"method" validate:"required_if=Provider webhook" gorm:"type:varchar(4);column:method"`
	Token    string                   `json:"token" validate:"required_if=Provider gotify" gorm:"column:token"`
}

type NotificationTriggerOptions struct {
	Enabled bool `json:"enabled"`
	DueDate bool `json:"due_date" validate:"required_if=Enabled true" gorm:"column:due_date;default:false"`
	PreDue  bool `json:"pre_due" validate:"required_if=Enabled true" gorm:"column:pre_due;default:false"`
	Overdue bool `json:"overdue" validate:"required_if=Enabled true" gorm:"column:overdue;default:false"`
}

type NotificationSettings struct {
	UserID   int                        `json:"-" gorm:"column:user_id;not null"`
	Provider NotificationProvider       `json:"provider" gorm:"embedded;embeddedPrefix:notifications_provider_;"`
	Triggers NotificationTriggerOptions `json:"triggers" gorm:"embedded;embeddedPrefix:notifications_triggers_;"`
}
