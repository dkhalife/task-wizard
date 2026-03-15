package users

import (
	"context"
	"encoding/json"
	"net/http"

	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/ws"
	"github.com/gin-gonic/gin"
)

type UsersMessageHandler struct {
	us *UserService
}

func NewUsersMessageHandler(us *UserService) *UsersMessageHandler {
	return &UsersMessageHandler{us: us}
}

func (h *UsersMessageHandler) updateNotificationSettings(ctx context.Context, userID int, msg ws.WSMessage) *ws.WSResponse {
	var req models.NotificationUpdateRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return &ws.WSResponse{
			Status: http.StatusBadRequest,
			Data: gin.H{
				"error": "Invalid request data",
			},
		}
	}

	status, response := h.us.UpdateNotificationSettings(ctx, userID, req)
	return &ws.WSResponse{
		Status: status,
		Data:   response,
	}
}

func UserMessages(ws *ws.WSServer, h *UsersMessageHandler) {
	ws.RegisterHandler("update_notification_settings", h.updateNotificationSettings)
}
