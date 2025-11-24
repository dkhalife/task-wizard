package models

import (
	"time"
)

type User struct {
	ID          int       `json:"id" gorm:"primary_key;not null"`
	DisplayName string    `json:"display_name" gorm:"column:display_name;not null"`
	Email       string    `json:"email" gorm:"column:email;unique;not null"`
	CreatedAt   time.Time `json:"-" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:"-" gorm:"column:updated_at;default:NULL;autoUpdateTime"`
	Disabled    bool      `json:"-" gorm:"column:disabled;default:false"`

	AppTokens            []AppToken           `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
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
	UserID  int
	TokenID int
	Type    IdentityType
	Scopes  []ApiTokenScope
}

type ApiTokenScope string

const (
	ApiTokenScopeTasksRead   ApiTokenScope = "Tasks.Read"
	ApiTokenScopeTasksWrite  ApiTokenScope = "Tasks.Write"
	ApiTokenScopeLabelsRead  ApiTokenScope = "Labels.Read"
	ApiTokenScopeLabelsWrite ApiTokenScope = "Labels.Write"
	ApiTokenScopeUserRead    ApiTokenScope = "User.Read"
	ApiTokenScopeUserWrite   ApiTokenScope = "User.Write"
	ApiTokenScopeTokensWrite ApiTokenScope = "Tokens.Write"
	ApiTokenScopeDavRead     ApiTokenScope = "Dav.Read"
	ApiTokenScopeDavWrite    ApiTokenScope = "Dav.Write"
)

type AppToken struct {
	ID        int       `json:"id" gorm:"primary_key"`
	UserID    int       `json:"user_id" gorm:"column:user_id;not null"`
	Name      string    `json:"name" gorm:"column:name;not null"`
	Token     string    `json:"token" gorm:"column:token;index;not null"`
	Scopes    []string  `json:"scopes" gorm:"column:scopes;serializer:json;type:json"`
	CreatedAt time.Time `json:"-" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	ExpiresAt time.Time `json:"expires_at" gorm:"column:expires_at;default:CURRENT_TIMESTAMP"`
	User      User      `json:"-" gorm:"foreignKey:UserID"`
}

type CreateAppTokenRequest struct {
	Name       string          `json:"name" binding:"required"`
	Scopes     []ApiTokenScope `json:"scopes" binding:"required"`
	Expiration int             `json:"expiration" binding:"required"`
}
