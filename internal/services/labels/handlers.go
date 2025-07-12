package labels

import (
	"context"
	"encoding/json"
	"net/http"

	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/ws"
	"github.com/gin-gonic/gin"
)

type LabelsMessageHandler struct {
	ls *LabelService
}

func NewLabelsMessageHandler(ls *LabelService) *LabelsMessageHandler {
	return &LabelsMessageHandler{
		ls: ls,
	}
}

func (h *LabelsMessageHandler) getUserLabels(ctx context.Context, userID int, _ ws.WSMessage) *ws.WSResponse {
	status, response := h.ls.GetUserLabels(ctx, userID)

	return &ws.WSResponse{
		Status: status,
		Data:   response,
	}
}

func (h *LabelsMessageHandler) createLabel(ctx context.Context, userID int, msg ws.WSMessage) *ws.WSResponse {
	var req models.CreateLabelReq
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return &ws.WSResponse{
			Status: http.StatusBadRequest,
			Data: &gin.H{
				"error": "Invalid request data",
			},
		}
	}

	status, response := h.ls.CreateLabel(ctx, userID, req)
	return &ws.WSResponse{
		Status: status,
		Data:   response,
	}
}

func (h *LabelsMessageHandler) updateLabel(ctx context.Context, userID int, msg ws.WSMessage) *ws.WSResponse {
	var req models.UpdateLabelReq
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return &ws.WSResponse{
			Status: http.StatusBadRequest,
			Data: &gin.H{
				"error": "Invalid request data",
			},
		}
	}

	status, response := h.ls.UpdateLabel(ctx, userID, req)
	return &ws.WSResponse{
		Status: status,
		Data:   response,
	}
}

func (h *LabelsMessageHandler) deleteLabel(ctx context.Context, userID int, msg ws.WSMessage) *ws.WSResponse {
	var labelID int
	if err := json.Unmarshal(msg.Data, &labelID); err != nil {
		return &ws.WSResponse{
			Status: http.StatusBadRequest,
			Data: &gin.H{
				"error": "Invalid label ID",
			},
		}
	}

	status, response := h.ls.DeleteLabel(ctx, userID, labelID)
	return &ws.WSResponse{
		Status: status,
		Data:   response,
	}
}

func LabelMessages(ws *ws.WSServer, h *LabelsMessageHandler) {
	ws.RegisterHandler("get_user_labels", h.getUserLabels)
	ws.RegisterHandler("create_label", h.createLabel)
	ws.RegisterHandler("update_label", h.updateLabel)
	ws.RegisterHandler("delete_label", h.deleteLabel)
}
