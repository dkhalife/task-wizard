package auth

import (
	"github.com/gin-gonic/gin"
	"taskwiz.app/core/internal/models"
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
