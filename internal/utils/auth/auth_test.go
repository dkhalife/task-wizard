package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodePasswordAndMatches(t *testing.T) {
	password := "securepassword"
	hashedPassword, err := EncodePassword(password)
	assert.NoError(t, err)
	assert.NoError(t, Matches(hashedPassword, password))
	assert.Error(t, Matches(hashedPassword, "wrongpassword"))
}

func TestEncodeAndDecodeEmailAndCode(t *testing.T) {
	email := "user@example.com"
	code := "reset-code"

	encoded := EncodeEmailAndCode(email, code)
	decodedEmail, decodedCode, err := DecodeEmailAndCode(encoded)

	assert.NoError(t, err)
	assert.Equal(t, email, decodedEmail)
	assert.Equal(t, code, decodedCode)
}

func TestGenerateEmailResetToken(t *testing.T) {
	token, err := GenerateEmailResetToken(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}
