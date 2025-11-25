package apis

import (
	"net/http"
	"strconv"

	authMW "dkhalife.com/tasks/core/internal/middleware/auth"
	"dkhalife.com/tasks/core/internal/models"
	tService "dkhalife.com/tasks/core/internal/services/tasks"
	auth "dkhalife.com/tasks/core/internal/utils/auth"
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

func (h *TasksAPIHandler) getCompletedTasks(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	limitStr := c.DefaultQuery("limit", "10")
	pageStr := c.DefaultQuery("page", "1")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID required",
		})
		return
	}

	endRecurrenceStr := c.DefaultQuery("endRecurrence", "false")
	endRecurrence, err := strconv.ParseBool(endRecurrenceStr)
	if err != nil {
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
	tasksRoutes.Use(auth.MiddlewareFunc())
	{
		tasksRoutes.GET("/", authMW.ScopeMiddleware(models.ApiTokenScopeTasksRead), h.getTasks)
		tasksRoutes.GET("/completed", authMW.ScopeMiddleware(models.ApiTokenScopeTasksRead), h.getCompletedTasks)
		tasksRoutes.PUT("/", authMW.ScopeMiddleware(models.ApiTokenScopeTasksWrite), h.editTask)
		tasksRoutes.POST("/", authMW.ScopeMiddleware(models.ApiTokenScopeTasksWrite), h.createTask)
		tasksRoutes.GET("/:id", authMW.ScopeMiddleware(models.ApiTokenScopeTasksRead), h.getTask)
		tasksRoutes.GET("/:id/history", authMW.ScopeMiddleware(models.ApiTokenScopeTasksRead), h.GetTaskHistory)
		tasksRoutes.POST("/:id/do", authMW.ScopeMiddleware(models.ApiTokenScopeTasksWrite), h.completeTask)
		tasksRoutes.POST("/:id/undo", authMW.ScopeMiddleware(models.ApiTokenScopeTasksWrite), h.uncompleteTask)
		tasksRoutes.POST("/:id/skip", authMW.ScopeMiddleware(models.ApiTokenScopeTasksWrite), h.skipTask)
		tasksRoutes.PUT("/:id/dueDate", authMW.ScopeMiddleware(models.ApiTokenScopeTasksWrite), h.updateDueDate)
		tasksRoutes.DELETE("/:id", authMW.ScopeMiddleware(models.ApiTokenScopeTasksWrite), h.deleteTask)
	}
}
