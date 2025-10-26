package models

import (
	"time"
)

type User struct {
	ID          int       `json:"id" gorm:"primary_key;not null"`
	DisplayName string    `json:"display_name" gorm:"column:display_name;not null"`
	Email       string    `json:"email" gorm:"column:email;unique;not null"`
	Password    string    `json:"-" gorm:"column:password;not null"`
	CreatedAt   time.Time `json:"-" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:"-" gorm:"column:updated_at;default:NULL;autoUpdateTime"`
	Disabled    bool      `json:"-" gorm:"column:disabled;default:false"`
}

type UserPasswordReset struct {
	UserID         int       `gorm:"column:user_id;primary_key;not null"`
	Email          string    `gorm:"column:email;not null"`
	Token          string    `gorm:"column:token;not null"`
	ExpirationDate time.Time `gorm:"column:expiration_date;not null"`
}

type AppToken struct {
	ID        int       `json:"id" gorm:"primary_key"`
	UserID    int       `json:"user_id" gorm:"column:user_id;not null"`
	Name      string    `json:"name" gorm:"column:name;not null"`
	Token     string    `json:"token" gorm:"column:token;index;not null"`
	Scopes    []string  `json:"scopes" gorm:"column:scopes;serializer:json;type:json"`
	CreatedAt time.Time `json:"-" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	ExpiresAt time.Time `json:"expires_at" gorm:"column:expires_at;default:CURRENT_TIMESTAMP"`
}
