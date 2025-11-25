package auth

import (
	"dkhalife.com/tasks/core/internal/models"
	"github.com/gin-gonic/gin"
)

var IdentityKey = "user_id"
var AppTokenKey = "token_id"

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

func ConvertScopesToStringArray(scopes []models.ApiTokenScope) []string {
	strScopes := make([]string, len(scopes))
	for i, scope := range scopes {
		strScopes[i] = string(scope)
	}

	return strScopes
}

func ConvertStringArrayToScopes(scopes []string) []models.ApiTokenScope {
	tokenScopes := make([]models.ApiTokenScope, len(scopes))
	for i, scope := range scopes {
		tokenScopes[i] = models.ApiTokenScope(scope)
	}

	return tokenScopes
}
