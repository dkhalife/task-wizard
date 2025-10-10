package kubeconfig

import (
	"context"
	"net/http"

	"dkhalife.com/tasks/core/internal/models"
	kubeconfigRepo "dkhalife.com/tasks/core/internal/repos/kubeconfig"
	"dkhalife.com/tasks/core/internal/utils/kubeconfig"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type KubeconfigService struct {
	repo kubeconfigRepo.IKubeconfigRepo
}

func NewKubeconfigService(repo kubeconfigRepo.IKubeconfigRepo) *KubeconfigService {
	return &KubeconfigService{
		repo: repo,
	}
}

// ImportKubeconfig parses and imports contexts from a kubeconfig YAML
func (s *KubeconfigService) ImportKubeconfig(ctx context.Context, userID int, kubeconfigYAML string) (int, interface{}) {
	// Parse kubeconfig
	parsedContexts, err := kubeconfig.ParseKubeconfig(kubeconfigYAML)
	if err != nil {
		if validationErr, ok := err.(kubeconfig.ValidationError); ok {
			return http.StatusBadRequest, gin.H{
				"error":   "Invalid kubeconfig",
				"field":   validationErr.Field,
				"message": validationErr.Message,
			}
		}
		return http.StatusBadRequest, gin.H{
			"error":   "Failed to parse kubeconfig",
			"message": err.Error(),
		}
	}

	// Create contexts
	var createdContexts []models.KubeContextResponse
	for _, parsed := range parsedContexts {
		kubeContext := &models.KubeContext{
			UserID:                   userID,
			Name:                     parsed.Name,
			ClusterName:              parsed.ClusterName,
			Server:                   parsed.Server,
			CertificateAuthorityData: parsed.CertificateAuthorityData,
			ContextUser:              parsed.UserName,
			Token:                    parsed.Token,
			ClientCertificate:        parsed.ClientCertificateData,
			ClientKey:                parsed.ClientKeyData,
			IsActive:                 false,
		}

		if err := s.repo.CreateContext(ctx, kubeContext); err != nil {
			return http.StatusInternalServerError, gin.H{
				"error":   "Failed to save context",
				"context": parsed.Name,
				"message": err.Error(),
			}
		}

		createdContexts = append(createdContexts, kubeContext.ToResponse())
	}

	return http.StatusCreated, gin.H{
		"contexts": createdContexts,
		"count":    len(createdContexts),
	}
}

// ListContexts retrieves all contexts for a user
func (s *KubeconfigService) ListContexts(ctx context.Context, userID int) (int, interface{}) {
	contexts, err := s.repo.GetAllContexts(ctx, userID)
	if err != nil {
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve contexts",
		}
	}

	// Convert to response format
	var responses []models.KubeContextResponse
	for _, kc := range contexts {
		responses = append(responses, kc.ToResponse())
	}

	return http.StatusOK, gin.H{
		"contexts": responses,
	}
}

// GetContext retrieves a specific context
func (s *KubeconfigService) GetContext(ctx context.Context, userID, contextID int) (int, interface{}) {
	kubeContext, err := s.repo.GetContextByID(ctx, userID, contextID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return http.StatusNotFound, gin.H{
				"error": "Context not found",
			}
		}
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve context",
		}
	}

	return http.StatusOK, kubeContext.ToResponse()
}

// SetActiveContext sets a context as active
func (s *KubeconfigService) SetActiveContext(ctx context.Context, userID, contextID int) (int, interface{}) {
	err := s.repo.SetActiveContext(ctx, userID, contextID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return http.StatusNotFound, gin.H{
				"error": "Context not found",
			}
		}
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to set active context",
		}
	}

	return http.StatusOK, gin.H{
		"message": "Active context updated",
	}
}

// DeleteContext deletes a context
func (s *KubeconfigService) DeleteContext(ctx context.Context, userID, contextID int) (int, interface{}) {
	err := s.repo.DeleteContext(ctx, userID, contextID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return http.StatusNotFound, gin.H{
				"error": "Context not found",
			}
		}
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to delete context",
		}
	}

	return http.StatusOK, gin.H{
		"message": "Context deleted",
	}
}

// GetActiveContext retrieves the active context for a user
func (s *KubeconfigService) GetActiveContext(ctx context.Context, userID int) (int, interface{}) {
	kubeContext, err := s.repo.GetActiveContext(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return http.StatusNotFound, gin.H{
				"error": "No active context found",
			}
		}
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve active context",
		}
	}

	return http.StatusOK, kubeContext.ToResponse()
}
