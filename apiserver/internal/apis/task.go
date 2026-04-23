package apis

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	authMW "dkhalife.com/tasks/core/internal/middleware/auth"
	"dkhalife.com/tasks/core/internal/models"
	tService "dkhalife.com/tasks/core/internal/services/tasks"
	"dkhalife.com/tasks/core/internal/telemetry"
	auth "dkhalife.com/tasks/core/internal/utils/auth"
	middleware "dkhalife.com/tasks/core/internal/utils/middleware"
	"github.com/gin-gonic/gin"
)

type TasksAPIHandler struct {
	tService *tService.TaskService
}

func TasksAPI(tService *tService.TaskService) *TasksAPIHandler {
	return &TasksAPIHandler{
		tService: tService,
	}
}

func (h *TasksAPIHandler) getTasks(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	status, response := h.tService.GetUserTasks(c, currentIdentity.UserID)
	c.JSON(status, response)
}

func (h *TasksAPIHandler) getTasksDueBefore(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	raw := c.Query("before")
	if raw == "" {
		telemetry.TrackWarning(c, "task_invalid_param", "task-handler", "Missing 'before' query parameter", nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "'before' query parameter is required",
		})
		return
	}

	before, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		telemetry.TrackWarning(c, "task_invalid_param", "task-handler", "Invalid 'before' format: "+raw, nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "'before' must be in RFC 3339 / ISO 8601 format (e.g. 2025-01-15T00:00:00Z)",
		})
		return
	}

	status, response := h.tService.GetTasksDueBefore(c, currentIdentity.UserID, before.UTC())
	c.JSON(status, response)
}

func (h *TasksAPIHandler) getTasksByLabel(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	rawID := c.Param("labelId")
	labelID, err := strconv.Atoi(rawID)
	if err != nil {
		telemetry.TrackWarning(c, "task_invalid_param", "task-handler", "Invalid label ID: "+rawID, nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid label ID",
		})
		return
	}

	status, response := h.tService.GetTasksByLabel(c, currentIdentity.UserID, labelID)
	c.JSON(status, response)
}

func (h *TasksAPIHandler) searchTasks(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		telemetry.TrackWarning(c, "task_invalid_param", "task-handler", "Missing 'q' query parameter", nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "'q' query parameter is required",
		})
		return
	}

	status, response := h.tService.SearchTasksByTitle(c, currentIdentity.UserID, query)
	c.JSON(status, response)
}

func (h *TasksAPIHandler) getCompletedTasks(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	limitStr := c.DefaultQuery("limit", "10")
	pageStr := c.DefaultQuery("page", "1")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		telemetry.TrackWarning(c, "task_invalid_param", "task-handler", "Invalid limit: "+limitStr, nil)
		c.Status(http.StatusBadRequest)
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		telemetry.TrackWarning(c, "task_invalid_param", "task-handler", "Invalid page: "+pageStr, nil)
		c.Status(http.StatusBadRequest)
		return
	}

	status, response := h.tService.GetCompletedTasks(c, currentIdentity.UserID, limit, page)
	c.JSON(status, response)
}

func (h *TasksAPIHandler) getTask(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
		telemetry.TrackWarning(c, "task_invalid_param", "task-handler", "Invalid task ID: "+rawID, nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID",
		})
		return
	}

	status, response := h.tService.GetTask(c, currentIdentity.UserID, id)
	c.JSON(status, response)
}

func (h *TasksAPIHandler) createTask(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	var TaskReq models.CreateTaskReq
	if err := c.ShouldBindJSON(&TaskReq); err != nil {
		telemetry.TrackWarning(c, "task_bind_failed", "task-handler", err.Error(), nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	status, response := h.tService.CreateTask(c, currentIdentity.UserID, TaskReq)
	c.JSON(status, response)
}

func (h *TasksAPIHandler) editTask(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	var TaskReq models.UpdateTaskReq
	if err := c.ShouldBindJSON(&TaskReq); err != nil {
		telemetry.TrackWarning(c, "task_bind_failed", "task-handler", err.Error(), nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	status, response := h.tService.EditTask(c, currentIdentity.UserID, TaskReq)
	c.JSON(status, response)
}

func (h *TasksAPIHandler) deleteTask(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
		telemetry.TrackWarning(c, "task_invalid_param", "task-handler", "Invalid task ID: "+rawID, nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID",
		})
		return
	}

	status, response := h.tService.DeleteTask(c, currentIdentity.UserID, id)
	c.JSON(status, response)
}

func (h *TasksAPIHandler) skipTask(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)

	if err != nil {
		telemetry.TrackWarning(c, "task_invalid_param", "task-handler", "Invalid task ID: "+rawID, nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID",
		})
		return
	}

	status, response := h.tService.SkipTask(c, currentIdentity.UserID, id)
	c.JSON(status, response)
}

func (h *TasksAPIHandler) updateDueDate(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	var dueDateReq models.UpdateDueDateReq
	if err := c.ShouldBindJSON(&dueDateReq); err != nil {
		telemetry.TrackWarning(c, "task_bind_failed", "task-handler", err.Error(), nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
		telemetry.TrackWarning(c, "task_invalid_param", "task-handler", "Invalid task ID: "+rawID, nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID required",
		})
		return
	}

	status, response := h.tService.UpdateDueDate(c, currentIdentity.UserID, id, dueDateReq)
	c.JSON(status, response)
}

func (h *TasksAPIHandler) completeTask(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	completeTaskID := c.Param("id")
	id, err := strconv.Atoi(completeTaskID)
	if err != nil {
		telemetry.TrackWarning(c, "task_invalid_param", "task-handler", "Invalid task ID: "+completeTaskID, nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID required",
		})
		return
	}

	endRecurrenceStr := c.DefaultQuery("endRecurrence", "false")
	endRecurrence, err := strconv.ParseBool(endRecurrenceStr)
	if err != nil {
		telemetry.TrackWarning(c, "task_invalid_param", "task-handler", "Invalid endRecurrence: "+endRecurrenceStr, nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid endRecurrence value",
		})
		return
	}

	status, response := h.tService.CompleteTask(c, currentIdentity.UserID, id, endRecurrence)
	c.JSON(status, response)
}

func (h *TasksAPIHandler) uncompleteTask(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
		telemetry.TrackWarning(c, "task_invalid_param", "task-handler", "Invalid task ID: "+rawID, nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID",
		})
		return
	}

	status, response := h.tService.UncompleteTask(c, currentIdentity.UserID, id)
	c.JSON(status, response)
}

func (h *TasksAPIHandler) GetTaskHistory(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
		telemetry.TrackWarning(c, "task_invalid_param", "task-handler", "Invalid task ID: "+rawID, nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID required",
		})
		return
	}

	status, response := h.tService.GetTaskHistory(c, currentIdentity.UserID, id)
	c.JSON(status, response)
}

func TaskRoutes(router *gin.Engine, h *TasksAPIHandler, auth *authMW.AuthMiddleware) {
	tasksRoutes := router.Group("api/v1/tasks")
	tasksRoutes.Use(auth.MiddlewareFunc(), middleware.DeletionGuardMiddleware())
	{
		tasksRoutes.GET("/", authMW.ScopeMiddleware(models.ApiTokenScopeTaskRead), h.getTasks)
		tasksRoutes.GET("/due", authMW.ScopeMiddleware(models.ApiTokenScopeTaskRead), h.getTasksDueBefore)
		tasksRoutes.GET("/label/:labelId", authMW.ScopeMiddleware(models.ApiTokenScopeTaskRead), h.getTasksByLabel)
		tasksRoutes.GET("/search", authMW.ScopeMiddleware(models.ApiTokenScopeTaskRead), h.searchTasks)
		tasksRoutes.GET("/completed", authMW.ScopeMiddleware(models.ApiTokenScopeTaskRead), h.getCompletedTasks)
		tasksRoutes.PUT("/", authMW.ScopeMiddleware(models.ApiTokenScopeTaskWrite), h.editTask)
		tasksRoutes.POST("/", authMW.ScopeMiddleware(models.ApiTokenScopeTaskWrite), h.createTask)
		tasksRoutes.GET("/:id", authMW.ScopeMiddleware(models.ApiTokenScopeTaskRead), h.getTask)
		tasksRoutes.GET("/:id/history", authMW.ScopeMiddleware(models.ApiTokenScopeTaskRead), h.GetTaskHistory)
		tasksRoutes.POST("/:id/do", authMW.ScopeMiddleware(models.ApiTokenScopeTaskWrite), h.completeTask)
		tasksRoutes.POST("/:id/undo", authMW.ScopeMiddleware(models.ApiTokenScopeTaskWrite), h.uncompleteTask)
		tasksRoutes.POST("/:id/skip", authMW.ScopeMiddleware(models.ApiTokenScopeTaskWrite), h.skipTask)
		tasksRoutes.PUT("/:id/dueDate", authMW.ScopeMiddleware(models.ApiTokenScopeTaskWrite), h.updateDueDate)
		tasksRoutes.DELETE("/:id", authMW.ScopeMiddleware(models.ApiTokenScopeTaskWrite), h.deleteTask)
	}
}
