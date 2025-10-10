package models

import "time"

// KubeContext represents a Kubernetes context stored in the database
type KubeContext struct {
	ID        int       `json:"id" gorm:"primary_key"`
	UserID    int       `json:"user_id" gorm:"column:user_id;not null;index"`
	Name      string    `json:"name" gorm:"column:name;not null"`
	IsActive  bool      `json:"is_active" gorm:"column:is_active;default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;default:NULL;autoUpdateTime"`

	// Cluster information
	ClusterName              string `json:"cluster_name" gorm:"column:cluster_name;not null"`
	Server                   string `json:"server" gorm:"column:server;not null"`
	CertificateAuthorityData string `json:"-" gorm:"column:certificate_authority_data;type:text"` // Base64 encoded, encrypted

	// User information
	ContextUser       string `json:"context_user" gorm:"column:context_user;not null"`
	Token             string `json:"-" gorm:"column:token;type:text"`             // Encrypted
	ClientCertificate string `json:"-" gorm:"column:client_certificate;type:text"` // Base64 encoded, encrypted
	ClientKey         string `json:"-" gorm:"column:client_key;type:text"`         // Base64 encoded, encrypted

	// Foreign key
	User User `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}

// KubeContextResponse represents the response returned to clients (with masked sensitive data)
type KubeContextResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	IsActive    bool      `json:"is_active"`
	ClusterName string    `json:"cluster_name"`
	Server      string    `json:"server"`
	ContextUser string    `json:"context_user"`
	HasToken    bool      `json:"has_token"`
	HasCert     bool      `json:"has_cert"`
	HasKey      bool      `json:"has_key"`
	HasCA       bool      `json:"has_ca"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts a KubeContext to KubeContextResponse with masked sensitive data
func (k *KubeContext) ToResponse() KubeContextResponse {
	return KubeContextResponse{
		ID:          k.ID,
		Name:        k.Name,
		IsActive:    k.IsActive,
		ClusterName: k.ClusterName,
		Server:      k.Server,
		ContextUser: k.ContextUser,
		HasToken:    k.Token != "",
		HasCert:     k.ClientCertificate != "",
		HasKey:      k.ClientKey != "",
		HasCA:       k.CertificateAuthorityData != "",
		CreatedAt:   k.CreatedAt,
		UpdatedAt:   k.UpdatedAt,
	}
}

// ImportKubeconfigRequest represents the request to import a kubeconfig
type ImportKubeconfigRequest struct {
	KubeconfigYAML string `json:"kubeconfig_yaml" binding:"required"`
}

// SetActiveContextRequest represents the request to set an active context
type SetActiveContextRequest struct {
	ContextID int `json:"context_id" binding:"required"`
}
