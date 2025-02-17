package task

import (
	"time"

	"github.com/lib/pq"
)

type FrequencyType string

const (
	RepeatOnce    = "once"
	RepeatDaily   = "daily"
	RepeatWeekly  = "weekly"
	RepeatMonthly = "monthly"
	RepeatYearly  = "yearly"
	RepeatCustom  = "custom"
)

type IntervalUnit string

const (
	Hours  IntervalUnit = "hours"
	Days   IntervalUnit = "days"
	Weeks  IntervalUnit = "weeks"
	Months IntervalUnit = "months"
	Years  IntervalUnit = "years"
)

type RepeatOn string

const (
	Interval       RepeatOn = "interval"
	DaysOfTheWeek  RepeatOn = "days_of_the_week"
	DayOfTheMonths RepeatOn = "day_of_the_months"
)

type Frequency struct {
	Type   FrequencyType `json:"type" validate:"required" gorm:"type:varchar(9)"`
	On     RepeatOn      `json:"on" validate:"required_if=Type interval custom" gorm:"type:varchar(18);default:null"`
	Every  int           `json:"every" validate:"required_if=On interval" gorm:"type:int;default:null"`
	Unit   IntervalUnit  `json:"unit" validate:"required_if=On interval" gorm:"type:varchar(9);default:null"`
	Days   pq.Int32Array `json:"days" validate:"required_if=Type custom On days_of_the_week,dive,gte=0,lte=6" gorm:"type:integer[];default:null"`
	Months pq.Int32Array `json:"months" validate:"required_if=Type custom On day_of_the_months,dive,gte=0,lte=11" gorm:"type:integer[];default:null"`
}

type Task struct {
	// TODO: Notification metadata should be separate columns
	ID                   int        `json:"id" gorm:"primary_key"`
	Title                string     `json:"title" gorm:"column:title"`
	Frequency            Frequency  `json:"frequency" gorm:"embedded;embeddedPrefix:frequency_"`
	NextDueDate          *time.Time `json:"next_due_date" gorm:"column:next_due_date;index"`
	IsRolling            bool       `json:"is_rolling" gorm:"column:is_rolling"`
	CreatedBy            int        `json:"created_by" gorm:"column:created_by"`
	IsActive             bool       `json:"is_active" gorm:"column:is_active"`
	Notification         bool       `json:"notification" gorm:"column:notification"`
	NotificationMetadata *string    `json:"notification_metadata" gorm:"column:notification_metadata"`
	CreatedAt            time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt            time.Time  `json:"updated_at" gorm:"column:updated_at"`
}

type TaskHistory struct {
	ID          int        `json:"id" gorm:"primary_key"`
	TaskID      int        `json:"task_id" gorm:"column:task_id"`
	CompletedAt *time.Time `json:"completed_at" gorm:"column:completed_at"`
	DueDate     *time.Time `json:"due_date" gorm:"column:due_date"`
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
