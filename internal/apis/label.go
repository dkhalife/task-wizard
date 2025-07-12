package apis

import (
	"net/http"
	"strconv"

	authMW "dkhalife.com/tasks/core/internal/middleware/auth"
	models "dkhalife.com/tasks/core/internal/models"
	lService "dkhalife.com/tasks/core/internal/services/labels"
	auth "dkhalife.com/tasks/core/internal/utils/auth"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type LabelsAPIHandler struct {
	ls *lService.LabelService
}

func LabelsAPI(ls *lService.LabelService) *LabelsAPIHandler {
	return &LabelsAPIHandler{
		ls: ls,
	}
}

func (h *LabelsAPIHandler) getLabels(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)
	status, response := h.ls.GetUserLabels(c, currentIdentity.UserID)
	c.JSON(status, response)
}

func (h *LabelsAPIHandler) createLabel(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	var req models.CreateLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	status, response := h.ls.CreateLabel(c, currentIdentity.UserID, req)
	c.JSON(status, response)
}

func (h *LabelsAPIHandler) updateLabel(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	var req models.UpdateLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	status, response := h.ls.UpdateLabel(c, currentIdentity.UserID, req)
	c.JSON(status, response)
}

func (h *LabelsAPIHandler) deleteLabel(c *gin.Context) {
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

	status, response := h.ls.DeleteLabel(c, currentIdentity.UserID, labelID)
	c.JSON(status, response)
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
