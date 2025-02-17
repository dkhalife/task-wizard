package label

import (
	"net/http"
	"strconv"

	lModel "dkhalife.com/tasks/core/internal/models/label"
	lRepo "dkhalife.com/tasks/core/internal/repos/label"
	auth "dkhalife.com/tasks/core/internal/utils/auth"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type LabelReq struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

type UpdateLabelReq struct {
	ID int `json:"id" binding:"required"`
	LabelReq
}

type Handler struct {
	lRepo *lRepo.LabelRepository
}

func NewHandler(lRepo *lRepo.LabelRepository) *Handler {
	return &Handler{
		lRepo: lRepo,
	}
}

func (h *Handler) getLabels(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to fetch labels",
		})
		return
	}

	labels, err := h.lRepo.GetUserLabels(c, currentUser.ID)
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

func (h *Handler) createLabel(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to create labels",
		})
		return
	}

	var req LabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	label := &lModel.Label{
		Name:      req.Name,
		Color:     req.Color,
		CreatedBy: currentUser.ID,
	}
	if err := h.lRepo.CreateLabels(c, []*lModel.Label{label}); err != nil {
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

func (h *Handler) updateLabel(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to update label",
		})
		return
	}

	var req UpdateLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	label := &lModel.Label{
		Name:  req.Name,
		Color: req.Color,
		ID:    req.ID,
	}

	if !h.lRepo.AreLabelsAssignableByUser(c, currentUser.ID, []int{label.ID}, nil) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not allowed to perform this update",
		})
	}

	if err := h.lRepo.UpdateLabel(c, currentUser.ID, label); err != nil {
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

func (h *Handler) deleteLabel(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to delete label",
		})
		return
	}

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

	if err := h.lRepo.DeassignLabelFromAllTaskAndDelete(c, currentUser.ID, labelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error unassociating label from task",
		})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func Routes(r *gin.Engine, h *Handler, auth *jwt.GinJWTMiddleware) {

	labelRoutes := r.Group("api/v1/labels")
	labelRoutes.Use(auth.MiddlewareFunc())
	{
		labelRoutes.GET("", h.getLabels)
		labelRoutes.POST("", h.createLabel)
		labelRoutes.PUT("", h.updateLabel)
		labelRoutes.DELETE("/:id", h.deleteLabel)
	}
}
