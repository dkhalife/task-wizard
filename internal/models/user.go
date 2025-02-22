package models

import (
	"time"
)

type User struct {
	ID          int       `json:"id" gorm:"primary_key"`
	DisplayName string    `json:"display_name" gorm:"column:display_name"`
	Email       string    `json:"email" gorm:"column:email;unique"`
	Password    string    `json:"-" gorm:"column:password"`
	CreatedAt   time.Time `json:"-" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"-" gorm:"column:updated_at"`
	Disabled    bool      `json:"-" gorm:"column:disabled"`

	APITokens            []APIToken           `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	ResetTokens          []UserPasswordReset  `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	NotificationSettings NotificationSettings `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Labels               []Label              `json:"-" gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE;"`
	Tasks                []Task               `json:"-" gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE;"`
}

type UserPasswordReset struct {
	ID             int       `gorm:"column:id"`
	UserID         int       `gorm:"column:user_id"`
	Email          string    `gorm:"column:email"`
	Token          string    `gorm:"column:token"`
	ExpirationDate time.Time `gorm:"column:expiration_date"`
}

type APIToken struct {
	ID        int       `json:"id" gorm:"primary_key"`
	UserID    int       `json:"user_id" gorm:"column:user_id"`
	Name      string    `json:"name" gorm:"column:name;unique"`
	Token     string    `json:"token" gorm:"column:token;index"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}
