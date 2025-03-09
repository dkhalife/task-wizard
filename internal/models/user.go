package models

import (
	"time"

	"github.com/lib/pq"
)

type User struct {
	ID          int       `json:"id" gorm:"primary_key;not null"`
	DisplayName string    `json:"display_name" gorm:"column:display_name;not null"`
	Email       string    `json:"email" gorm:"column:email;unique;not null"`
	Password    string    `json:"-" gorm:"column:password;not null"`
	CreatedAt   time.Time `json:"-" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:"-" gorm:"column:updated_at;default:NULL;autoUpdateTime"`
	Disabled    bool      `json:"-" gorm:"column:disabled;default:false"`

	AppTokens            []AppToken           `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	ResetTokens          []UserPasswordReset  `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	NotificationSettings NotificationSettings `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Labels               []Label              `json:"-" gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE;"`
	Tasks                []Task               `json:"-" gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE;"`
}

type IdentityType string

const (
	IdentityTypeUser IdentityType = "user"
	IdentityTypeApp  IdentityType = "app"
)

type SignedInIdentity struct {
	UserID int
	Type   IdentityType
	Scopes []ApiTokenScope
}

type UserPasswordReset struct {
	UserID         int       `gorm:"column:user_id;primary_key;not null"`
	Email          string    `gorm:"column:email;not null"`
	Token          string    `gorm:"column:token;not null"`
	ExpirationDate time.Time `gorm:"column:expiration_date;not null"`
}

type ApiTokenScope string

const (
	ApiTokenScopeTaskRead   ApiTokenScope = "task:read"
	ApiTokenScopeTaskWrite  ApiTokenScope = "task:write"
	ApiTokenScopeLabelRead  ApiTokenScope = "label:read"
	ApiTokenScopeLabelWrite ApiTokenScope = "label:write"
	ApiTokenScopeUserRead   ApiTokenScope = "user:read"
	ApiTokenScopeUserWrite  ApiTokenScope = "user:write"
	ApiTokenScopeTokenWrite ApiTokenScope = "token:write"
)

type AppToken struct {
	ID        int            `json:"id" gorm:"primary_key;not null"`
	UserID    int            `json:"user_id" gorm:"column:user_id;not null"`
	Name      string         `json:"name" gorm:"column:name;not null"`
	Token     string         `json:"token" gorm:"column:token;index;not null"`
	Scopes    pq.StringArray `json:"scopes" gorm:"column:scopes;type:text[]"`
	CreatedAt time.Time      `json:"-" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"column:expires_at;default:CURRENT_TIMESTAMP"`
}
