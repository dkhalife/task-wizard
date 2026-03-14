package auth

import (
	"net/http/httptest"
	"testing"

	"dkhalife.com/tasks/core/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCurrentIdentity_Present(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	expected := &models.SignedInIdentity{
		UserID:  42,
		TokenID: 0,
		Type:    models.IdentityTypeUser,
		Scopes:  models.AllUserScopes(),
	}
	c.Set(IdentityKey, expected)

	result := CurrentIdentity(c)
	assert.NotNil(t, result)
	assert.Equal(t, 42, result.UserID)
	assert.Equal(t, models.IdentityTypeUser, result.Type)
}

func TestCurrentIdentity_Missing(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	result := CurrentIdentity(c)
	assert.Nil(t, result)
}
