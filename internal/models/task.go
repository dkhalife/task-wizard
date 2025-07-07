package models

import (
	"time"
)

type Task struct {
	ID           int                        `json:"id" gorm:"primary_key"`
	Title        string                     `json:"title" gorm:"column:title;not null"`
	Frequency    Frequency                  `json:"frequency" gorm:"embedded;embeddedPrefix:frequency_"`
	NextDueDate  *time.Time                 `json:"next_due_date" gorm:"column:next_due_date;index"`
	EndDate      *time.Time                 `json:"end_date" gorm:"column:end_date;default:NULL"`
	IsRolling    bool                       `json:"is_rolling" gorm:"column:is_rolling;default:false"`
	CreatedBy    int                        `json:"-" gorm:"column:created_by;not null;index:idx_tasks_created_by"`
	IsActive     bool                       `json:"-" gorm:"column:is_active;default:true"`
	Notification NotificationTriggerOptions `json:"notification" gorm:"embedded;embeddedPrefix:notification_"`
	CreatedAt    time.Time                  `json:"-" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt    *time.Time                 `json:"-" gorm:"column:updated_at;default:NULL;autoUpdateTime"`

	Labels        []Label        `json:"labels" gorm:"many2many:task_labels;constraint:OnDelete:CASCADE"`
	History       []TaskHistory  `json:"-" gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE;"`
	Notifications []Notification `json:"-" gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE;"`
}

type TaskHistory struct {
	ID            int        `json:"id" gorm:"primary_key"`
	TaskID        int        `json:"task_id" gorm:"column:task_id;not null;index:idx_task_histories_task_id"`
	CompletedDate *time.Time `json:"completed_date" gorm:"column:completed_date"`
	DueDate       *time.Time `json:"due_date" gorm:"column:due_date"`
}

type TaskLabel struct {
	TaskID  int `json:"task_id"`
	LabelID int `json:"label_id"`
}
