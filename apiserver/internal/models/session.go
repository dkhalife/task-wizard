package models

import "time"

type Session struct {
	ID        int       `gorm:"primary_key;autoIncrement"`
	UserID    int       `gorm:"not null;index"`
	TokenHash string    `gorm:"column:token_hash;size:64;not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
