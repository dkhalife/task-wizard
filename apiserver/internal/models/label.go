package models

import (
	"time"
)

type Label struct {
	ID        int        `json:"id" gorm:"primary_key"`
	Name      string     `json:"name" gorm:"column:name;not null"`
	Color     string     `json:"color" gorm:"type:varchar(7);column:color;not null"`
	CreatedBy int        `json:"created_by" gorm:"column:created_by;not null;index:idx_labels_created_by"`
	CreatedAt time.Time  `json:"-" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time `json:"-" gorm:"column:updated_at;default:NULL;autoUpdateTime"`

	User  User   `json:"user" gorm:"foreignKey:CreatedBy"`
	Tasks []Task `json:"-" gorm:"many2many:task_labels;constraint:OnDelete:CASCADE"`
}

type CreateLabelReq struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

type UpdateLabelReq struct {
	ID int `json:"id" binding:"required"`
	CreateLabelReq
}
