package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/services/logging"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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

func EncodeEmailAndCode(email, code string) string {
	data := email + ":" + code
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func DecodeEmailAndCode(encoded string) (string, string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", err
	}
	parts := string(data)
	split := strings.Split(parts, ":")
	if len(split) != 2 {
		return "", "", fmt.Errorf("invalid format")
	}
	return split[0], split[1], nil
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
