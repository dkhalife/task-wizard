package labels

import (
	"context"
	"net/http"
	"strings"

	"dkhalife.com/tasks/core/internal/models"
	repos "dkhalife.com/tasks/core/internal/repos/label"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/telemetry"
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
		log := logging.FromContext(ctx)
		log.Errorf("Failed to get labels: %s", err.Error())
		telemetry.TrackError(ctx, "label_get_failed", "label-service", err, nil)
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

	exists, err := s.r.LabelExistsByName(ctx, userId, req.Name, 0)
	if err != nil {
		log.Errorf("Failed to check label existence: %s", err.Error())
		telemetry.TrackError(ctx, "label_check_failed", "label-service", err, nil)
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to create label",
		}
	}

	if exists {
		return http.StatusConflict, gin.H{
			"error": "A label with this name already exists",
		}
	}

	label := &models.Label{
		Name:      req.Name,
		Color:     req.Color,
		CreatedBy: userId,
	}

	if err := s.r.CreateLabels(ctx, []*models.Label{label}); err != nil {
		if isDuplicateKeyError(err) {
			return http.StatusConflict, gin.H{
				"error": "A label with this name already exists",
			}
		}

		log.Errorf("Failed to create label: %s", err.Error())
		telemetry.TrackError(ctx, "label_create_failed", "label-service", err, nil)
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to create label",
		}
	}

	newLabel := gin.H{
		"label": gin.H{
			"id":    label.ID,
			"name":  label.Name,
			"color": label.Color,
		},
	}

	s.ws.BroadcastToUser(userId, ws.WSResponse{
		Action: "label_created",
		Data:   newLabel,
	})

	return http.StatusCreated, gin.H{
		"label": label.ID,
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
		telemetry.TrackWarning(ctx, "label_forbidden", "label-service", "User not allowed to update label", nil)
		return http.StatusForbidden, gin.H{
			"error": "You are not allowed to perform this update",
		}
	}

	exists, err := s.r.LabelExistsByName(ctx, userId, req.Name, req.ID)
	if err != nil {
		log.Errorf("Failed to check label existence: %s", err.Error())
		telemetry.TrackError(ctx, "label_check_failed", "label-service", err, nil)
		return http.StatusInternalServerError, gin.H{
			"error": "Error updating label",
		}
	}

	if exists {
		return http.StatusConflict, gin.H{
			"error": "A label with this name already exists",
		}
	}

	if err := s.r.UpdateLabel(ctx, userId, label); err != nil {
		if isDuplicateKeyError(err) {
			return http.StatusConflict, gin.H{
				"error": "A label with this name already exists",
			}
		}

		log.Errorf("Failed to update label: %s", err.Error())
		telemetry.TrackError(ctx, "label_update_failed", "label-service", err, nil)
		return http.StatusInternalServerError, gin.H{
			"error": "Error updating label",
		}
	}

	updatedLabel := gin.H{
		"label": gin.H{
			"id":    label.ID,
			"name":  label.Name,
			"color": label.Color,
		},
	}

	s.ws.BroadcastToUser(userId, ws.WSResponse{
		Action: "label_updated",
		Data:   updatedLabel,
	})

	return http.StatusOK, updatedLabel
}

func (s *LabelService) DeleteLabel(ctx context.Context, userID int, labelID int) (int, interface{}) {
	if err := s.r.DeleteLabel(ctx, userID, labelID); err != nil {
		log := logging.FromContext(ctx)
		log.Errorf("Failed to delete label: %s", err.Error())
		telemetry.TrackError(ctx, "label_delete_failed", "label-service", err, nil)
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to delete label",
		}
	}

	s.ws.BroadcastToUser(userID, ws.WSResponse{
		Action: "label_deleted",
		Data: gin.H{
			"id": labelID,
		},
	})
	return http.StatusNoContent, nil
}

func isDuplicateKeyError(err error) bool {
	msg := err.Error()
	// SQLite: "UNIQUE constraint failed: ..."
	// MySQL: "Error 1062 (23000): Duplicate entry ..."
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "Duplicate entry") ||
		strings.Contains(msg, "Error 1062")
}
