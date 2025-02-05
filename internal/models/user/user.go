package user

import (
	"time"

	nModel "donetick.com/core/internal/models/notifier"
)

type User struct {
	ID               int                     `json:"id" gorm:"primary_key"`
	DisplayName      string                  `json:"display_name" gorm:"column:display_name"`
	Username         string                  `json:"username" gorm:"column:username;unique"`
	Email            string                  `json:"email" gorm:"column:email;unique"`
	Password         string                  `json:"-" gorm:"column:password"`
	CreatedAt        time.Time               `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time               `json:"updated_at" gorm:"column:updated_at"`
	Disabled         bool                    `json:"disabled" gorm:"column:disabled"`
	NotificationType nModel.NotificationType `json:"notification_type" gorm:"column:type"`
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
	Name      string    `json:"name" gorm:"column:name;unique"`
	UserID    int       `json:"user_id" gorm:"column:user_id;index"`
	Token     string    `json:"token" gorm:"column:token;index"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}
