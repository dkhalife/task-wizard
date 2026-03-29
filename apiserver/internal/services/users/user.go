package users

import (
	"context"
	"net/http"
	"time"

	"dkhalife.com/tasks/core/internal/models"
	repos "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/telemetry"
	"dkhalife.com/tasks/core/internal/ws"
	"github.com/gin-gonic/gin"
)

const deletionGracePeriod = 24 * time.Hour

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
		telemetry.TrackError(ctx, "notification_settings_update_failed", "user-service", err, nil)
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to update notification target",
		}
	}

	if req.Provider.Provider == models.NotificationProviderNone {
		err = s.r.DeleteNotificationsForUser(ctx, userID)
		if err != nil {
			log.Errorf("failed to delete existing notification: %s", err.Error())
			telemetry.TrackError(ctx, "notification_delete_failed", "user-service", err, nil)
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

func (s *UserService) RequestDeletion(ctx context.Context, userID int) (int, interface{}) {
	log := logging.FromContext(ctx)
	if err := s.r.RequestDeletion(ctx, userID); err != nil {
		log.Errorf("failed to request account deletion: %s", err.Error())
		telemetry.TrackError(ctx, "account_deletion_request_failed", "user-service", err, nil)
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to request account deletion",
		}
	}

	s.ws.BroadcastToUser(userID, ws.WSResponse{
		Action: "account_deletion_requested",
		Data:   gin.H{},
	})

	return http.StatusNoContent, gin.H{}
}

func (s *UserService) CancelDeletion(ctx context.Context, userID int) (int, interface{}) {
	log := logging.FromContext(ctx)
	if err := s.r.CancelDeletion(ctx, userID); err != nil {
		log.Errorf("failed to cancel account deletion: %s", err.Error())
		telemetry.TrackError(ctx, "account_deletion_cancel_failed", "user-service", err, nil)
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to cancel account deletion",
		}
	}

	s.ws.BroadcastToUser(userID, ws.WSResponse{
		Action: "account_deletion_cancelled",
		Data:   gin.H{},
	})

	return http.StatusNoContent, gin.H{}
}

func (s *UserService) ProcessDeletions(ctx context.Context) error {
	log := logging.FromContext(ctx)

	users, err := s.r.FindUsersForDeletion(ctx, deletionGracePeriod)
	if err != nil {
		return err
	}

	for _, user := range users {
		log.Infof("deleting account for user %d (requested at %s)", user.ID, user.DeletionRequestedAt)
		if err := s.r.DeleteUser(ctx, user.ID); err != nil {
			log.Errorf("failed to delete user %d: %s", user.ID, err.Error())
			telemetry.TrackError(ctx, "account_deletion_failed", "user-service", err, map[string]string{
				"user_id": string(rune(user.ID)),
			})
		}
	}

	return nil
}

