package users

import (
	"context"
	"net/http"

	"dkhalife.com/tasks/core/internal/models"
	repos "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/ws"
	"github.com/gin-gonic/gin"
)

type UserService struct {
	r  *repos.UserRepository
	ws *ws.WSServer
}

func NewUserService(r *repos.UserRepository, ws *ws.WSServer) *UserService {
	return &UserService{r: r, ws: ws}
}

func (s *UserService) GetAllAppTokens(ctx context.Context, userID int) (int, interface{}) {
	log := logging.FromContext(ctx)

	tokens, err := s.r.GetAllUserTokens(ctx, userID)
	if err != nil {
		log.Errorf("failed to get user tokens: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to get user tokens",
		}
	}

	return http.StatusOK, gin.H{
		"tokens": tokens,
	}
}

func (s *UserService) CreateAppToken(ctx context.Context, userID int, req models.CreateAppTokenRequest) (int, interface{}) {
	if len(req.Scopes) == 0 {
		return http.StatusBadRequest, gin.H{
			"error": "Tokens must have at least one scope",
		}
	}

	days := req.Expiration
	if days < 1 || days > 90 {
		return http.StatusBadRequest, gin.H{
			"error": "Expiration can at most be 90 days into the future",
		}
	}

	log := logging.FromContext(ctx)
	token, err := s.r.CreateAppToken(ctx, userID, req.Name, req.Scopes, req.Expiration)
	if err != nil {
		log.Errorf("failed to create token: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to create token",
		}
	}

	s.ws.BroadcastToUser(userID, ws.WSResponse{
		Action: "app_token_created",
		Data:   token,
	})

	return http.StatusCreated, gin.H{
		"token": token,
	}
}

func (s *UserService) DeleteAppToken(ctx context.Context, userID int, tokenID int) (int, interface{}) {
	log := logging.FromContext(ctx)

	err := s.r.DeleteAppToken(ctx, userID, tokenID)
	if err != nil {
		log.Errorf("failed to delete token: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Failed to delete the token",
		}
	}

	s.ws.BroadcastToUser(userID, ws.WSResponse{
		Action: "app_token_deleted",
		Data: gin.H{
			"id": tokenID,
		},
	})

	return http.StatusNoContent, nil
}
