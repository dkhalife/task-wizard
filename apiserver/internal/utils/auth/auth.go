package auth

import (
	"dkhalife.com/tasks/core/internal/models"
	"github.com/gin-gonic/gin"
)

var IdentityKey = "user_id"

func CurrentIdentity(c *gin.Context) *models.SignedInIdentity {
	data, ok := c.Get(IdentityKey)
	if !ok {
		return nil
	}

	acc, ok := data.(*models.SignedInIdentity)
	if !ok {
		return nil
	}

	return acc
}
