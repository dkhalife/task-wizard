package task

import (
	"time"

	lModel "donetick.com/core/internal/models/label"
)

type FrequencyType string

const (
	FrequancyTypeOnce          FrequencyType = "once"
	FrequancyTypeDaily         FrequencyType = "daily"
	FrequancyTypeWeekly        FrequencyType = "weekly"
	FrequancyTypeMonthly       FrequencyType = "monthly"
	FrequancyTypeYearly        FrequencyType = "yearly"
	FrequancyTypeInterval      FrequencyType = "interval"
	FrequancyTypeDayOfTheWeek  FrequencyType = "days_of_the_week"
	FrequancyTypeDayOfTheMonth FrequencyType = "day_of_the_month"
	FrequancyTypeNoRepeat      FrequencyType = "no_repeat"
)

type Task struct {
	// TODO: Frequency metadata should either be a different set of columns or be deleted
	// TODO: Notification metadata should be separate columns
	ID                   int                   `json:"id" gorm:"primary_key"`
	Name                 string                `json:"name" gorm:"column:name"`
	FrequencyType        FrequencyType         `json:"frequencyType" gorm:"column:frequency_type"`
	Frequency            int                   `json:"frequency" gorm:"column:frequency"`
	FrequencyMetadata    *FrequencyMetadata    `json:"frequencyMetadata" gorm:"column:frequency_meta"`
	NextDueDate          *time.Time            `json:"nextDueDate" gorm:"column:next_due_date;index"`
	IsRolling            bool                  `json:"isRolling" gorm:"column:is_rolling"`
	CreatedBy            int                   `json:"createdBy" gorm:"column:created_by"`
	IsActive             bool                  `json:"isActive" gorm:"column:is_active"`
	Notification         bool                  `json:"notification" gorm:"column:notification"`
	NotificationMetadata *NotificationMetadata `json:"notificationMetadata" gorm:"column:notification_meta"`
	Labels               *[]lModel.Label       `json:"labels" gorm:"many2many:task_labels"`
	CreatedAt            time.Time             `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt            time.Time             `json:"updatedAt" gorm:"column:updated_at"`
}

type TaskHistory struct {
	ID          int        `json:"id" gorm:"primary_key"`
	TaskID      int        `json:"taskId" gorm:"column:task_id"`
	CompletedAt *time.Time `json:"completedAt" gorm:"column:completed_at"`
	DueDate     *time.Time `json:"dueDate" gorm:"column:due_date"`
}

type FrequencyMetadata struct {
	Days   []*string `json:"days,omitempty"`
	Months []*string `json:"months,omitempty"`
	Unit   *string   `json:"unit,omitempty"`
	Time   string    `json:"time,omitempty"`
}

type NotificationMetadata struct {
	DueDate    bool `json:"dueDate,omitempty"`
	Completion bool `json:"completion,omitempty"`
	Nagging    bool `json:"nagging,omitempty"`
	PreDue     bool `json:"predue,omitempty"`
}

type TaskDetail struct {
	ID                  int           `json:"id" gorm:"column:id"`
	Name                string        `json:"name" gorm:"column:name"`
	FrequencyType       FrequencyType `json:"frequencyType" gorm:"column:frequency_type"`
	NextDueDate         *time.Time    `json:"nextDueDate" gorm:"column:next_due_date"`
	LastCompletedDate   *time.Time    `json:"lastCompletedDate" gorm:"column:last_completed_date"`
	LastCompletedBy     *int          `json:"lastCompletedBy" gorm:"column:last_completed_by"`
	TotalCompletedCount int           `json:"totalCompletedCount" gorm:"column:total_completed"`
	CreatedBy           int           `json:"createdBy" gorm:"column:created_by"`
}

type TaskLabels struct {
	TaskID  int `json:"taskId" gorm:"primaryKey;autoIncrement:false;not null"`
	LabelID int `json:"labelId" gorm:"primaryKey;autoIncrement:false;not null"`
	UserID  int `json:"userId" gorm:"primaryKey;autoIncrement:false;not null"`
	Label   lModel.Label
}
