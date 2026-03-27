package users

import (
	"context"
	"net/http"

	"dkhalife.com/tasks/core/internal/models"
	repos "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/telemetry"
	"dkhalife.com/tasks/core/internal/ws"
	"github.com/gin-gonic/gin"
)

type UserService struct {
	r  repos.IUserRepo
	ws *ws.WSServer
}

func NewUserService(r repos.IUserRepo, ws *ws.WSServer) *UserService {
	return &UserService{r: r, ws: ws}
}

func (s *UserService) UpdateNotificationSettings(ctx context.Context, userID int, req models.NotificationUpdateRequest) (int, interface{}) {
	log := logging.FromContext(ctx)
	err := s.r.UpdateNotificationSettings(ctx, userID, req.Provider, req.Triggers)
	if err != nil {
		log.Errorf("failed to update notification target: %s", err.Error())
		telemetry.TrackError(nil, "notification_settings_update_failed", "user-service", err, nil)
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to update notification target",
		}
	}

	if req.Provider.Provider == models.NotificationProviderNone {
		err = s.r.DeleteNotificationsForUser(ctx, userID)
		if err != nil {
			log.Errorf("failed to delete existing notification: %s", err.Error())
			telemetry.TrackError(nil, "notification_delete_failed", "user-service", err, nil)
			return http.StatusInternalServerError, gin.H{
				"error": "Failed to delete existing notification",
			}
		}
	}

	s.ws.BroadcastToUser(userID, ws.WSResponse{
		Action: "notification_settings_updated",
		Data: gin.H{
			"provider": req.Provider,
			"triggers": req.Triggers,
		},
	})

	return http.StatusNoContent, gin.H{}
}
