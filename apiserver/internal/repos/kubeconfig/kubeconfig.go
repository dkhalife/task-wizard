package repos

import (
	"context"
	"fmt"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/utils/encryption"
	"gorm.io/gorm"
)

type IKubeconfigRepo interface {
	CreateContext(ctx context.Context, kubeContext *models.KubeContext) error
	GetContextByID(ctx context.Context, userID, contextID int) (*models.KubeContext, error)
	GetAllContexts(ctx context.Context, userID int) ([]models.KubeContext, error)
	SetActiveContext(ctx context.Context, userID, contextID int) error
	DeleteContext(ctx context.Context, userID, contextID int) error
	GetActiveContext(ctx context.Context, userID int) (*models.KubeContext, error)
}

type KubeconfigRepo struct {
	db        *gorm.DB
	encryptor *encryption.Encryptor
}

var _ IKubeconfigRepo = (*KubeconfigRepo)(nil)

func NewKubeconfigRepository(db *gorm.DB, cfg *config.Config) IKubeconfigRepo {
	return &KubeconfigRepo{
		db:        db,
		encryptor: encryption.NewEncryptor(cfg.Jwt.Secret),
	}
}

// CreateContext creates a new kubeconfig context with encrypted sensitive data
func (r *KubeconfigRepo) CreateContext(ctx context.Context, kubeContext *models.KubeContext) error {
	// Encrypt sensitive fields
	var err error
	if kubeContext.Token != "" {
		kubeContext.Token, err = r.encryptor.Encrypt(kubeContext.Token)
		if err != nil {
			return fmt.Errorf("failed to encrypt token: %w", err)
		}
	}

	if kubeContext.ClientCertificate != "" {
		kubeContext.ClientCertificate, err = r.encryptor.Encrypt(kubeContext.ClientCertificate)
		if err != nil {
			return fmt.Errorf("failed to encrypt client certificate: %w", err)
		}
	}

	if kubeContext.ClientKey != "" {
		kubeContext.ClientKey, err = r.encryptor.Encrypt(kubeContext.ClientKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt client key: %w", err)
		}
	}

	if kubeContext.CertificateAuthorityData != "" {
		kubeContext.CertificateAuthorityData, err = r.encryptor.Encrypt(kubeContext.CertificateAuthorityData)
		if err != nil {
			return fmt.Errorf("failed to encrypt CA data: %w", err)
		}
	}

	return r.db.WithContext(ctx).Create(kubeContext).Error
}

// GetContextByID retrieves a kubeconfig context by ID with decrypted sensitive data
func (r *KubeconfigRepo) GetContextByID(ctx context.Context, userID, contextID int) (*models.KubeContext, error) {
	var kubeContext models.KubeContext
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", contextID, userID).
		First(&kubeContext).Error
	if err != nil {
		return nil, err
	}

	// Decrypt sensitive fields
	if err := r.decryptContext(&kubeContext); err != nil {
		return nil, err
	}

	return &kubeContext, nil
}

// GetAllContexts retrieves all kubeconfig contexts for a user
func (r *KubeconfigRepo) GetAllContexts(ctx context.Context, userID int) ([]models.KubeContext, error) {
	var contexts []models.KubeContext
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&contexts).Error
	if err != nil {
		return nil, err
	}

	// Decrypt sensitive fields for all contexts
	for i := range contexts {
		if err := r.decryptContext(&contexts[i]); err != nil {
			return nil, err
		}
	}

	return contexts, nil
}

// SetActiveContext sets a context as active and deactivates all others for the user
func (r *KubeconfigRepo) SetActiveContext(ctx context.Context, userID, contextID int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First, verify the context exists and belongs to the user
		var kubeContext models.KubeContext
		if err := tx.Where("id = ? AND user_id = ?", contextID, userID).First(&kubeContext).Error; err != nil {
			return err
		}

		// Deactivate all contexts for this user
		if err := tx.Model(&models.KubeContext{}).
			Where("user_id = ?", userID).
			Update("is_active", false).Error; err != nil {
			return err
		}

		// Activate the specified context
		return tx.Model(&models.KubeContext{}).
			Where("id = ? AND user_id = ?", contextID, userID).
			Update("is_active", true).Error
	})
}

// DeleteContext deletes a kubeconfig context
func (r *KubeconfigRepo) DeleteContext(ctx context.Context, userID, contextID int) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", contextID, userID).
		Delete(&models.KubeContext{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// GetActiveContext retrieves the active kubeconfig context for a user
func (r *KubeconfigRepo) GetActiveContext(ctx context.Context, userID int) (*models.KubeContext, error) {
	var kubeContext models.KubeContext
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		First(&kubeContext).Error
	if err != nil {
		return nil, err
	}

	// Decrypt sensitive fields
	if err := r.decryptContext(&kubeContext); err != nil {
		return nil, err
	}

	return &kubeContext, nil
}

// decryptContext decrypts all sensitive fields in a KubeContext
func (r *KubeconfigRepo) decryptContext(kubeContext *models.KubeContext) error {
	var err error

	if kubeContext.Token != "" {
		kubeContext.Token, err = r.encryptor.Decrypt(kubeContext.Token)
		if err != nil {
			return fmt.Errorf("failed to decrypt token: %w", err)
		}
	}

	if kubeContext.ClientCertificate != "" {
		kubeContext.ClientCertificate, err = r.encryptor.Decrypt(kubeContext.ClientCertificate)
		if err != nil {
			return fmt.Errorf("failed to decrypt client certificate: %w", err)
		}
	}

	if kubeContext.ClientKey != "" {
		kubeContext.ClientKey, err = r.encryptor.Decrypt(kubeContext.ClientKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt client key: %w", err)
		}
	}

	if kubeContext.CertificateAuthorityData != "" {
		kubeContext.CertificateAuthorityData, err = r.encryptor.Decrypt(kubeContext.CertificateAuthorityData)
		if err != nil {
			return fmt.Errorf("failed to decrypt CA data: %w", err)
		}
	}

	return nil
}
