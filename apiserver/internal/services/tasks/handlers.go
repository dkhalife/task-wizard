package tasks

import (
	"context"
	"encoding/json"
	"net/http"

	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/ws"
	"github.com/gin-gonic/gin"
)

// TasksMessageHandler provides websocket handlers for task related messages.
type TasksMessageHandler struct {
	ts *TaskService
}

// NewTasksMessageHandler returns a new TasksMessageHandler.
func NewTasksMessageHandler(ts *TaskService) *TasksMessageHandler {
	return &TasksMessageHandler{ts: ts}
}

func (h *TasksMessageHandler) getUserTasks(ctx context.Context, userID int, _ ws.WSMessage) *ws.WSResponse {
	status, response := h.ts.GetUserTasks(ctx, userID)
	return &ws.WSResponse{
		Status: status,
		Data:   response,
	}
}

func (h *TasksMessageHandler) getCompletedTasks(ctx context.Context, userID int, msg ws.WSMessage) *ws.WSResponse {
	var req struct {
		Limit int `json:"limit"`
		Page  int `json:"page"`
	}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return &ws.WSResponse{
			Status: http.StatusBadRequest,
			Data: gin.H{
				"error": "Invalid request data",
			},
		}
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	status, response := h.ts.GetCompletedTasks(ctx, userID, req.Limit, req.Page)
	return &ws.WSResponse{Status: status, Data: response}
}

func (h *TasksMessageHandler) getTask(ctx context.Context, userID int, msg ws.WSMessage) *ws.WSResponse {
	var id int
	if err := json.Unmarshal(msg.Data, &id); err != nil {
		return &ws.WSResponse{
			Status: http.StatusBadRequest,
			Data: gin.H{
				"error": "Invalid task ID",
			},
		}
	}
	status, response := h.ts.GetTask(ctx, userID, id)
	return &ws.WSResponse{
		Status: status,
		Data:   response,
	}
}

func (h *TasksMessageHandler) createTask(ctx context.Context, userID int, msg ws.WSMessage) *ws.WSResponse {
	var req models.CreateTaskReq
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return &ws.WSResponse{
			Status: http.StatusBadRequest,
			Data: gin.H{
				"error": "Invalid request data",
			},
		}
	}
	status, response := h.ts.CreateTask(ctx, userID, req)
	return &ws.WSResponse{
		Status: status,
		Data:   response,
	}
}

func (h *TasksMessageHandler) updateTask(ctx context.Context, userID int, msg ws.WSMessage) *ws.WSResponse {
	var req models.UpdateTaskReq
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return &ws.WSResponse{
			Status: http.StatusBadRequest,
			Data: gin.H{
				"error": "Invalid request data",
			},
		}
	}
	status, response := h.ts.EditTask(ctx, userID, req)
	return &ws.WSResponse{
		Status: status,
		Data:   response,
	}
}

func (h *TasksMessageHandler) deleteTask(ctx context.Context, userID int, msg ws.WSMessage) *ws.WSResponse {
	var id int
	if err := json.Unmarshal(msg.Data, &id); err != nil {
		return &ws.WSResponse{
			Status: http.StatusBadRequest,
			Data: gin.H{
				"error": "Invalid task ID",
			},
		}
	}
	status, response := h.ts.DeleteTask(ctx, userID, id)
	return &ws.WSResponse{
		Status: status,
		Data:   response,
	}
}

func (h *TasksMessageHandler) skipTask(ctx context.Context, userID int, msg ws.WSMessage) *ws.WSResponse {
	var id int
	if err := json.Unmarshal(msg.Data, &id); err != nil {
		return &ws.WSResponse{
			Status: http.StatusBadRequest,
			Data: gin.H{
				"error": "Invalid task ID",
			},
		}
	}
	status, response := h.ts.SkipTask(ctx, userID, id)
	return &ws.WSResponse{
		Status: status,
		Data:   response,
	}
}

func (h *TasksMessageHandler) updateDueDate(ctx context.Context, userID int, msg ws.WSMessage) *ws.WSResponse {
	var req struct {
		ID int `json:"id"`
		models.UpdateDueDateReq
	}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return &ws.WSResponse{
			Status: http.StatusBadRequest,
			Data: gin.H{
				"error": "Invalid request data",
			},
		}
	}
	status, response := h.ts.UpdateDueDate(ctx, userID, req.ID, req.UpdateDueDateReq)
	return &ws.WSResponse{
		Status: status,
		Data:   response,
	}
}

func (h *TasksMessageHandler) completeTask(ctx context.Context, userID int, msg ws.WSMessage) *ws.WSResponse {
	var id int
	if err := json.Unmarshal(msg.Data, &id); err != nil {
		return &ws.WSResponse{
			Status: http.StatusBadRequest,
			Data: gin.H{
				"error": "Invalid task ID",
			},
		}
	}
	status, response := h.ts.CompleteTask(ctx, userID, id)
	return &ws.WSResponse{
		Status: status,
		Data:   response,
	}
}

func (h *TasksMessageHandler) uncompleteTask(ctx context.Context, userID int, msg ws.WSMessage) *ws.WSResponse {
	var id int
	if err := json.Unmarshal(msg.Data, &id); err != nil {
		return &ws.WSResponse{
			Status: http.StatusBadRequest,
			Data: gin.H{
				"error": "Invalid task ID",
			},
		}
	}
	status, response := h.ts.UncompleteTask(ctx, userID, id)
	return &ws.WSResponse{
		Status: status,
		Data:   response,
	}
}

func (h *TasksMessageHandler) getTaskHistory(ctx context.Context, userID int, msg ws.WSMessage) *ws.WSResponse {
	var id int
	if err := json.Unmarshal(msg.Data, &id); err != nil {
		return &ws.WSResponse{
			Status: http.StatusBadRequest,
			Data: gin.H{
				"error": "Invalid task ID",
			},
		}
	}
	status, response := h.ts.GetTaskHistory(ctx, userID, id)
	return &ws.WSResponse{
		Status: status,
		Data:   response,
	}
}

// TaskMessages registers websocket handlers for task actions.
func TaskMessages(wsServer *ws.WSServer, h *TasksMessageHandler) {
	wsServer.RegisterHandler("get_tasks", h.getUserTasks)
	wsServer.RegisterHandler("get_completed_tasks", h.getCompletedTasks)
	wsServer.RegisterHandler("get_task", h.getTask)
	wsServer.RegisterHandler("create_task", h.createTask)
	wsServer.RegisterHandler("update_task", h.updateTask)
	wsServer.RegisterHandler("delete_task", h.deleteTask)
	wsServer.RegisterHandler("skip_task", h.skipTask)
	wsServer.RegisterHandler("update_due_date", h.updateDueDate)
	wsServer.RegisterHandler("complete_task", h.completeTask)
	wsServer.RegisterHandler("uncomplete_task", h.uncompleteTask)
	wsServer.RegisterHandler("get_task_history", h.getTaskHistory)
}
