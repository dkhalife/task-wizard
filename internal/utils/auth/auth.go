package auth

import (
	"crypto/rand"
	"encoding/base64"

	"dkhalife.com/tasks/core/internal/services/logging"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	uModel "dkhalife.com/tasks/core/internal/models/user"
)

const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()_+-=[]{}|;':,.<>?/~"

var IdentityKey = "id"

func EncodePassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func Matches(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func GenerateEmailResetToken(c *gin.Context) (string, error) {
	logger := logging.FromContext(c)
	// Define the length of the token (in bytes). For example, 32 bytes will result in a 44-character base64-encoded token.
	tokenLength := 32

	// Generate a random byte slice.
	tokenBytes := make([]byte, tokenLength)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		logger.Errorw("password.GenerateEmailResetToken failed to generate random bytes", "err", err)
		return "", err
	}

	// Encode the byte slice to a base64 string.
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	return token, nil
}

func CurrentUser(c *gin.Context) (*uModel.User, bool) {
	data, ok := c.Get(IdentityKey)
	if !ok {
		return nil, false
	}
	acc, ok := data.(*uModel.User)
	return acc, ok
}

func MustCurrentUser(c *gin.Context) *uModel.User {
	acc, ok := CurrentUser(c)
	if ok {
		return acc
	}
	panic("no account in gin.Context")
}
