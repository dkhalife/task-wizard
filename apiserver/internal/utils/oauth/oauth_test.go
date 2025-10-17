package oauth

import (
	"context"
	"testing"
	"time"

	"dkhalife.com/tasks/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type OAuthTestSuite struct {
	suite.Suite
}

func TestOAuthTestSuite(t *testing.T) {
	suite.Run(t, new(OAuthTestSuite))
}

func (suite *OAuthTestSuite) TestNewOAuthProvider_Disabled() {
	cfg := &config.OAuthConfig{
		Enabled: false,
	}

	provider, err := NewOAuthProvider(cfg)

	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), provider)
}

func (suite *OAuthTestSuite) TestNewOAuthProvider_MissingClientID() {
	cfg := &config.OAuthConfig{
		Enabled:      true,
		ClientSecret: "secret",
		AuthorizeURL: "https://example.com/authorize",
		TokenURL:     "https://example.com/token",
		RedirectURL:  "https://example.com/callback",
		Scope:        "openid",
	}

	provider, err := NewOAuthProvider(cfg)

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), provider)
	assert.Contains(suite.T(), err.Error(), "client_id")
}

func (suite *OAuthTestSuite) TestNewOAuthProvider_MissingClientSecret() {
	cfg := &config.OAuthConfig{
		Enabled:      true,
		ClientID:     "client-id",
		AuthorizeURL: "https://example.com/authorize",
		TokenURL:     "https://example.com/token",
		RedirectURL:  "https://example.com/callback",
		Scope:        "openid",
	}

	provider, err := NewOAuthProvider(cfg)

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), provider)
	assert.Contains(suite.T(), err.Error(), "client_secret")
}

func (suite *OAuthTestSuite) TestNewOAuthProvider_Success() {
	cfg := &config.OAuthConfig{
		Enabled:      true,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		AuthorizeURL: "https://example.com/authorize",
		TokenURL:     "https://example.com/token",
		RedirectURL:  "https://example.com/callback",
		Scope:        "openid",
	}

	provider, err := NewOAuthProvider(cfg)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), provider)
	assert.Equal(suite.T(), cfg.ClientID, provider.config.ClientID)
	assert.Equal(suite.T(), cfg.ClientSecret, provider.config.ClientSecret)
}

func (suite *OAuthTestSuite) TestGenerateStateToken() {
	token1, err1 := GenerateStateToken()
	token2, err2 := GenerateStateToken()

	assert.Nil(suite.T(), err1)
	assert.Nil(suite.T(), err2)
	assert.NotEmpty(suite.T(), token1)
	assert.NotEmpty(suite.T(), token2)
	assert.NotEqual(suite.T(), token1, token2)
}

func (suite *OAuthTestSuite) TestGetAuthorizationURL() {
	cfg := &config.OAuthConfig{
		Enabled:      true,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		AuthorizeURL: "https://example.com/authorize",
		TokenURL:     "https://example.com/token",
		RedirectURL:  "https://example.com/callback",
		Scope:        "openid",
	}

	provider, _ := NewOAuthProvider(cfg)
	state := "test-state"

	url := provider.GetAuthorizationURL(state)

	assert.Contains(suite.T(), url, "https://example.com/authorize")
	assert.Contains(suite.T(), url, "client_id=client-id")
	assert.Contains(suite.T(), url, "state=test-state")
	assert.Contains(suite.T(), url, "redirect_uri=https%3A%2F%2Fexample.com%2Fcallback")
}

func (suite *OAuthTestSuite) TestExchangeCode_InvalidCode() {
	cfg := &config.OAuthConfig{
		Enabled:      true,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		AuthorizeURL: "https://example.com/authorize",
		TokenURL:     "https://example.com/token",
		RedirectURL:  "https://example.com/callback",
		Scope:        "openid",
	}

	provider, _ := NewOAuthProvider(cfg)
	ctx := context.Background()

	// This will fail because we're not hitting a real OAuth server
	token, err := provider.ExchangeCode(ctx, "invalid-code")

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), token)
}

func (suite *OAuthTestSuite) TestValidateToken_NoVerifier() {
	cfg := &config.OAuthConfig{
		Enabled:      true,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		AuthorizeURL: "https://example.com/authorize",
		TokenURL:     "https://example.com/token",
		RedirectURL:  "https://example.com/callback",
		Scope:        "openid",
		// No JwksURL, so no verifier
	}

	provider, _ := NewOAuthProvider(cfg)
	ctx := context.Background()

	claims, err := provider.ValidateToken(ctx, "some-token")

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), claims)
	assert.Contains(suite.T(), err.Error(), "not configured")
}

func (suite *OAuthTestSuite) TestClaimsIsExpired() {
	// Not expired
	claims := &Claims{
		Exp: time.Now().Add(1 * time.Hour).Unix(),
	}
	assert.False(suite.T(), claims.IsExpired())

	// Expired
	expiredClaims := &Claims{
		Exp: time.Now().Add(-1 * time.Hour).Unix(),
	}
	assert.True(suite.T(), expiredClaims.IsExpired())
}
