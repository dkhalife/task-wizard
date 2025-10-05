package auth

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AuthTestSuite struct {
	suite.Suite
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

func (s *AuthTestSuite) TestEncodePasswordAndMatches() {
	password := "securepassword"
	hashedPassword, err := EncodePassword(password)
	s.Require().NoError(err)
	s.NoError(Matches(hashedPassword, password))
	s.Error(Matches(hashedPassword, "wrongpassword"))
}

func (s *AuthTestSuite) TestEncodeAndDecodeEmailAndCode() {
	email := "user@example.com"
	code := "reset-code"

	encoded := EncodeEmailAndCode(email, code)
	decodedEmail, decodedCode, err := DecodeEmailAndCode(encoded)

	s.Require().NoError(err)
	s.Equal(email, decodedEmail)
	s.Equal(code, decodedCode)
}

func (s *AuthTestSuite) TestGenerateEmailResetToken() {
	token, err := GenerateEmailResetToken(nil)
	s.NoError(err)
	s.NotEmpty(token)
}
