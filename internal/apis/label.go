package apis

import (
	"net/http"
	"strconv"

	authMW "dkhalife.com/tasks/core/internal/middleware/auth"
	models "dkhalife.com/tasks/core/internal/models"
	lRepo "dkhalife.com/tasks/core/internal/repos/label"
	"dkhalife.com/tasks/core/internal/services/logging"
	auth "dkhalife.com/tasks/core/internal/utils/auth"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type CreateLabelReq struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

type UpdateLabelReq struct {
	ID int `json:"id" binding:"required"`
	CreateLabelReq
}

type LabelsAPIHandler struct {
	lRepo *lRepo.LabelRepository
}

func LabelsAPI(lRepo *lRepo.LabelRepository) *LabelsAPIHandler {
	return &LabelsAPIHandler{
		lRepo: lRepo,
	}
}

func (h *LabelsAPIHandler) getLabels(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)
	labels, err := h.lRepo.GetUserLabels(c, currentIdentity.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get labels",
		})
		return
	}

	labelResponses := make([]gin.H, len(labels))
	for i, label := range labels {
		labelResponses[i] = gin.H{
			"id":    label.ID,
			"name":  label.Name,
			"color": label.Color,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"labels": labelResponses,
	})
}

func (h *LabelsAPIHandler) createLabel(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	log := logging.FromContext(c)

	var req CreateLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	label := &models.Label{
		Name:      req.Name,
		Color:     req.Color,
		CreatedBy: currentIdentity.UserID,
	}

	if err := h.lRepo.CreateLabels(c, []*models.Label{label}); err != nil {
		log.Errorf("Failed to create label: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create label",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"label": gin.H{
			"id":    label.ID,
			"name":  label.Name,
			"color": label.Color,
		},
	})
}

func (h *LabelsAPIHandler) updateLabel(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	log := logging.FromContext(c)

	var req UpdateLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	label := &models.Label{
		Name:  req.Name,
		Color: req.Color,
		ID:    req.ID,
	}

	if !h.lRepo.AreLabelsAssignableByUser(c, currentIdentity.UserID, []int{label.ID}) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not allowed to perform this update",
		})
		return
	}

	if err := h.lRepo.UpdateLabel(c, currentIdentity.UserID, label); err != nil {
		log.Errorf("Failed to update label: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error updating label",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"label": gin.H{
			"id":    label.ID,
			"name":  label.Name,
			"color": label.Color,
		},
	})
}

func (h *LabelsAPIHandler) deleteLabel(c *gin.Context) {
	log := logging.FromContext(c)
	currentIdentity := auth.CurrentIdentity(c)

	labelIDRaw := c.Param("id")
	if labelIDRaw == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Label ID is required",
		})
		return
	}

	labelID, err := strconv.Atoi(labelIDRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid label ID",
		})
		return
	}

	if err := h.lRepo.DeleteLabel(c, currentIdentity.UserID, labelID); err != nil {
		log.Errorf("Failed to delete label: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error unassociating label from task",
		})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func LabelRoutes(r *gin.Engine, h *LabelsAPIHandler, authGate *jwt.GinJWTMiddleware) {
	labelRoutes := r.Group("api/v1/labels")
	labelRoutes.Use(authGate.MiddlewareFunc())
	{
		labelRoutes.GET("", authMW.ScopeMiddleware(models.ApiTokenScopeLabelRead), h.getLabels)
		labelRoutes.POST("", authMW.ScopeMiddleware(models.ApiTokenScopeLabelWrite), h.createLabel)
		labelRoutes.PUT("", authMW.ScopeMiddleware(models.ApiTokenScopeLabelWrite), h.updateLabel)
		labelRoutes.DELETE("/:id", authMW.ScopeMiddleware(models.ApiTokenScopeLabelWrite), h.deleteLabel)
	}
}
