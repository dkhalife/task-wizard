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
	ApiTokenScopeTaskRead   ApiTokenScope = "Tasks.Read"
	ApiTokenScopeTaskWrite  ApiTokenScope = "Tasks.Write"
	ApiTokenScopeLabelRead  ApiTokenScope = "Labels.Read"
	ApiTokenScopeLabelWrite ApiTokenScope = "Labels.Write"
	ApiTokenScopeUserRead   ApiTokenScope = "User.Read"
	ApiTokenScopeUserWrite  ApiTokenScope = "User.Write"
	ApiTokenScopeDavRead    ApiTokenScope = "Dav.Read"
	ApiTokenScopeDavWrite   ApiTokenScope = "Dav.Write"
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
