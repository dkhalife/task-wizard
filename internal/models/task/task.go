package task

import (
	"time"
)

type FrequencyType string

const (
	FrequencyTypeOnce          FrequencyType = "once"
	FrequencyTypeDaily         FrequencyType = "daily"
	FrequencyTypeWeekly        FrequencyType = "weekly"
	FrequencyTypeMonthly       FrequencyType = "monthly"
	FrequencyTypeYearly        FrequencyType = "yearly"
	FrequencyTypeInterval      FrequencyType = "interval"
	FrequencyTypeDayOfTheWeek  FrequencyType = "days_of_the_week"
	FrequencyTypeDayOfTheMonth FrequencyType = "day_of_the_month"
	FrequencyTypeNoRepeat      FrequencyType = "no_repeat"
)

type Task struct {
	// TODO: Frequency metadata should either be a different set of columns or be deleted
	// TODO: Notification metadata should be separate columns
	ID                   int           `json:"id" gorm:"primary_key"`
	Title                string        `json:"title" gorm:"column:title"`
	FrequencyType        FrequencyType `json:"frequency_type" gorm:"column:frequency_type"`
	Frequency            int           `json:"frequency" gorm:"column:frequency"`
	FrequencyMetadata    *string       `json:"frequency_metadata" gorm:"column:frequency_metadata"`
	NextDueDate          *time.Time    `json:"next_due_date" gorm:"column:next_due_date;index"`
	IsRolling            bool          `json:"is_rolling" gorm:"column:is_rolling"`
	CreatedBy            int           `json:"created_by" gorm:"column:created_by"`
	IsActive             bool          `json:"is_active" gorm:"column:is_active"`
	Notification         bool          `json:"notification" gorm:"column:notification"`
	NotificationMetadata *string       `json:"notification_metadata" gorm:"column:notification_metadata"`
	CreatedAt            time.Time     `json:"created_at" gorm:"column:created_at"`
	UpdatedAt            time.Time     `json:"updated_at" gorm:"column:updated_at"`
}

type TaskHistory struct {
	ID          int        `json:"id" gorm:"primary_key"`
	TaskID      int        `json:"task_id" gorm:"column:task_id"`
	CompletedAt *time.Time `json:"completed_at" gorm:"column:completed_at"`
	DueDate     *time.Time `json:"due_date" gorm:"column:due_date"`
}

type FrequencyMetadata struct {
	Days   []*string `json:"days,omitempty"`
	Months []*string `json:"months,omitempty"`
	Unit   *string   `json:"unit,omitempty"`
	Time   string    `json:"time,omitempty"`
}

type NotificationMetadata struct {
	DueDate    bool `json:"due_date,omitempty"`
	Completion bool `json:"completion,omitempty"`
	Nagging    bool `json:"nagging,omitempty"`
	PreDue     bool `json:"predue,omitempty"`
}

type TaskLabels struct {
	TaskID  int `json:"task_id" gorm:"primaryKey;autoIncrement:false;not null"`
	LabelID int `json:"label_id" gorm:"primaryKey;autoIncrement:false;not null"`
	UserID  int `json:"user_id" gorm:"primaryKey;autoIncrement:false;not null"`
}
