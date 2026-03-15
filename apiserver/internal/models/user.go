package models

import (
	"time"
)

type User struct {
	ID          int       `json:"id" gorm:"primary_key;not null"`
	DirectoryID string    `json:"-" gorm:"column:directory_id;not null;default:''"`
	ObjectID    string    `json:"-" gorm:"column:object_id;not null;default:''"`
	CreatedAt   time.Time `json:"-" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:"-" gorm:"column:updated_at;default:NULL;autoUpdateTime"`
	Disabled    bool      `json:"-" gorm:"column:disabled;default:false"`

	NotificationSettings NotificationSettings `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Labels               []Label              `json:"-" gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE;"`
	Tasks                []Task               `json:"-" gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE;"`
}

type IdentityType string

const (
	IdentityTypeUser IdentityType = "user"
)

type SignedInIdentity struct {
	UserID int
	Type   IdentityType
	Scopes []ApiTokenScope
}

type ApiTokenScope string

const (
	ApiTokenScopeTaskRead   ApiTokenScope = "task:read"
	ApiTokenScopeTaskWrite  ApiTokenScope = "task:write"
	ApiTokenScopeLabelRead  ApiTokenScope = "label:read"
	ApiTokenScopeLabelWrite ApiTokenScope = "label:write"
	ApiTokenScopeUserRead   ApiTokenScope = "user:read"
	ApiTokenScopeUserWrite  ApiTokenScope = "user:write"
	ApiTokenScopeDavRead    ApiTokenScope = "dav:read"
	ApiTokenScopeDavWrite   ApiTokenScope = "dav:write"
)

func AllUserScopes() []ApiTokenScope {
	return []ApiTokenScope{
		ApiTokenScopeTaskRead,
		ApiTokenScopeTaskWrite,
		ApiTokenScopeLabelRead,
		ApiTokenScopeLabelWrite,
		ApiTokenScopeUserRead,
		ApiTokenScopeUserWrite,
		ApiTokenScopeDavRead,
		ApiTokenScopeDavWrite,
	}
}
