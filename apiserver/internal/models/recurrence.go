package models

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
	Days   []int32       `json:"days" validate:"required_if=Type custom On days_of_the_week,dive,gte=0,lte=6" gorm:"serializer:json;type:json"`
	Months []int32       `json:"months" validate:"required_if=Type custom On day_of_the_months,dive,gte=0,lte=11" gorm:"serializer:json;type:json"`
}
