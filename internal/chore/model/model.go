package model

import (
	"time"
)

type FrequencyType string

const (
	FrequancyTypeOnce          FrequencyType = "once"
	FrequancyTypeDaily         FrequencyType = "daily"
	FrequancyTypeWeekly        FrequencyType = "weekly"
	FrequancyTypeMonthly       FrequencyType = "monthly"
	FrequancyTypeYearly        FrequencyType = "yearly"
	FrequancyTypeAdaptive      FrequencyType = "adaptive"
	FrequancyTypeIntervel      FrequencyType = "interval"
	FrequancyTypeDayOfTheWeek  FrequencyType = "days_of_the_week"
	FrequancyTypeDayOfTheMonth FrequencyType = "day_of_the_month"
	FrequancyTypeNoRepeat      FrequencyType = "no_repeat"
)

type Chore struct {
	ID                   int           `json:"id" gorm:"primary_key"`
	Name                 string        `json:"name" gorm:"column:name"`                              // Chore description
	FrequencyType        FrequencyType `json:"frequencyType" gorm:"column:frequency_type"`           // "daily", "weekly", "monthly", "yearly", "adaptive",or "custom"
	Frequency            int           `json:"frequency" gorm:"column:frequency"`                    // Number of days, weeks, months, or years between chores
	FrequencyMetadata    *string       `json:"frequencyMetadata" gorm:"column:frequency_meta"`       // Additional frequency information
	NextDueDate          *time.Time    `json:"nextDueDate" gorm:"column:next_due_date;index"`        // When the chore is due
	IsRolling            bool          `json:"isRolling" gorm:"column:is_rolling"`                   // Whether the chore is rolling
	CreatedBy            int           `json:"createdBy" gorm:"column:created_by"`                   // Who the chore was created by
	IsActive             bool          `json:"isActive" gorm:"column:is_active"`                     // Whether the chore is active
	Notification         bool          `json:"notification" gorm:"column:notification"`              // Whether the chore has notification
	NotificationMetadata *string       `json:"notificationMetadata" gorm:"column:notification_meta"` // Additional notification information
	Labels               *string       `json:"labels" gorm:"column:labels"`                          // Labels for the chore
	LabelsV2             *[]Label      `json:"labelsV2" gorm:"many2many:chore_labels"`               // Labels for the chore
	CreatedAt            time.Time     `json:"createdAt" gorm:"column:created_at"`                   // When the chore was created
	UpdatedAt            time.Time     `json:"updatedAt" gorm:"column:updated_at"`                   // When the chore was last updated
	Status               int           `json:"status" gorm:"column:status"`
}

type ChoreHistory struct {
	ID          int        `json:"id" gorm:"primary_key"`                  // Unique identifier
	ChoreID     int        `json:"choreId" gorm:"column:chore_id"`         // The chore this history is for
	CompletedAt *time.Time `json:"completedAt" gorm:"column:completed_at"` // When the chore was completed
	CompletedBy int        `json:"completedBy" gorm:"column:completed_by"` // Who completed the chore
	DueDate     *time.Time `json:"dueDate" gorm:"column:due_date"`         // When the chore was due
	UpdatedAt   *time.Time `json:"updatedAt" gorm:"column:updated_at"`     // When the record was last updated
}
type ChoreHistoryStatus int8

const (
	ChoreHistoryStatusPending       ChoreHistoryStatus = 0
	ChoreHistoryStatusCompleted     ChoreHistoryStatus = 1
	ChoreHistoryStatusCompletedLate ChoreHistoryStatus = 2
	ChoreHistoryStatusMissed        ChoreHistoryStatus = 3
	ChoreHistoryStatusSkipped       ChoreHistoryStatus = 4
)

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

type ChoreDetail struct {
	ID                  int        `json:"id" gorm:"column:id"`
	Name                string     `json:"name" gorm:"column:name"`
	FrequencyType       string     `json:"frequencyType" gorm:"column:frequency_type"`
	NextDueDate         *time.Time `json:"nextDueDate" gorm:"column:next_due_date"`
	LastCompletedDate   *time.Time `json:"lastCompletedDate" gorm:"column:last_completed_date"`
	LastCompletedBy     *int       `json:"lastCompletedBy" gorm:"column:last_completed_by"`
	TotalCompletedCount int        `json:"totalCompletedCount" gorm:"column:total_completed"`
	CreatedBy           int        `json:"createdBy" gorm:"column:created_by"`
}

type Label struct {
	ID        int    `json:"id" gorm:"primary_key"`
	Name      string `json:"name" gorm:"column:name"`
	Color     string `json:"color" gorm:"column:color"`
	CreatedBy int    `json:"created_by" gorm:"column:created_by"`
}

type ChoreLabels struct {
	ChoreID int `json:"choreId" gorm:"primaryKey;autoIncrement:false;not null"`
	LabelID int `json:"labelId" gorm:"primaryKey;autoIncrement:false;not null"`
	UserID  int `json:"userId" gorm:"primaryKey;autoIncrement:false;not null"`
	Label   Label
}
