package labels

import (
	"context"
	"net/http"

	"dkhalife.com/tasks/core/internal/models"
	repos "dkhalife.com/tasks/core/internal/repos/label"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/ws"
	"github.com/gin-gonic/gin"
)

type LabelService struct {
	r  *repos.LabelRepository
	ws *ws.WSServer
}

func NewLabelService(r *repos.LabelRepository, ws *ws.WSServer) *LabelService {
	return &LabelService{r: r, ws: ws}
}

func (s *LabelService) GetUserLabels(ctx context.Context, userID int) (int, interface{}) {
	labels, err := s.r.GetUserLabels(ctx, userID)
	if err != nil {
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to get labels",
		}
	}

	labelResponses := make([]gin.H, len(labels))
	for i, label := range labels {
		labelResponses[i] = gin.H{
			"id":    label.ID,
			"name":  label.Name,
			"color": label.Color,
		}
	}

	return http.StatusOK, gin.H{
		"labels": labelResponses,
	}
}

func (s *LabelService) CreateLabel(ctx context.Context, userId int, req models.CreateLabelReq) (int, interface{}) {
	log := logging.FromContext(ctx)

	label := &models.Label{
		Name:      req.Name,
		Color:     req.Color,
		CreatedBy: userId,
	}

	if err := s.r.CreateLabels(ctx, []*models.Label{label}); err != nil {
		log.Errorf("Failed to create label: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to create label",
		}
	}

	return http.StatusCreated, gin.H{
		"label": gin.H{
			"id":    label.ID,
			"name":  label.Name,
			"color": label.Color,
		},
	}
}

func (s *LabelService) UpdateLabel(ctx context.Context, userId int, req models.UpdateLabelReq) (int, interface{}) {
	log := logging.FromContext(ctx)

	label := &models.Label{
		Name:  req.Name,
		Color: req.Color,
		ID:    req.ID,
	}

	if !s.r.AreLabelsAssignableByUser(ctx, userId, []int{label.ID}) {
		return http.StatusForbidden, gin.H{
			"error": "You are not allowed to perform this update",
		}
	}

	if err := s.r.UpdateLabel(ctx, userId, label); err != nil {
		log.Errorf("Failed to update label: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error updating label",
		}
	}

	return http.StatusOK, gin.H{
		"label": gin.H{
			"id":    label.ID,
			"name":  label.Name,
			"color": label.Color,
		},
	}
}

func (s *LabelService) DeleteLabel(ctx context.Context, userID int, labelID int) (int, interface{}) {
	if err := s.r.DeleteLabel(ctx, userID, labelID); err != nil {
		log := logging.FromContext(ctx)
		log.Errorf("Failed to delete label: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error unassociating label from task",
		}
	}

	return http.StatusNoContent, nil
}
