package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"taskwiz.app/core/config"
)

const testLeeway = 2 * time.Minute

func newDisabledAuthConfig(hostName string, allowInsecure bool) *config.Config {
	return &config.Config{
		Entra: config.EntraConfig{Enabled: false},
		Server: config.ServerConfig{
			HostName:            hostName,
			AllowInsecureNoAuth: allowInsecure,
		},
	}
}

func TestNewAuthMiddleware_DisabledWithHostNameAndNoOptIn_Fails(t *testing.T) {
	cfg := newDisabledAuthConfig("example.com", false)

	m, err := NewAuthMiddleware(cfg, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, m)
}

func TestNewAuthMiddleware_DisabledWithOptIn_Allowed(t *testing.T) {
	cfg := newDisabledAuthConfig("example.com", true)

	m, err := NewAuthMiddleware(cfg, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.False(t, m.enabled)
}

func TestNewAuthMiddleware_DisabledWithoutHostName_Allowed(t *testing.T) {
	cfg := newDisabledAuthConfig("", false)

	m, err := NewAuthMiddleware(cfg, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.False(t, m.enabled)
}

func TestValidateTemporalClaims_WithinWindow_Valid(t *testing.T) {
	now := time.Now()
	claims := accessTokenClaims{
		ExpiresAt: now.Add(10 * time.Minute).Unix(),
		NotBefore: now.Add(-10 * time.Minute).Unix(),
	}

	assert.NoError(t, validateTemporalClaims(claims, now, testLeeway))
}

func TestValidateTemporalClaims_ExpiredWithinLeeway_Valid(t *testing.T) {
	now := time.Now()
	claims := accessTokenClaims{
		ExpiresAt: now.Add(-1 * time.Minute).Unix(),
	}

	assert.NoError(t, validateTemporalClaims(claims, now, testLeeway))
}

func TestValidateTemporalClaims_ExpiredBeyondLeeway_Invalid(t *testing.T) {
	now := time.Now()
	claims := accessTokenClaims{
		ExpiresAt: now.Add(-5 * time.Minute).Unix(),
	}

	err := validateTemporalClaims(claims, now, testLeeway)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestValidateTemporalClaims_NotBeforeWithinLeeway_Valid(t *testing.T) {
	now := time.Now()
	claims := accessTokenClaims{
		ExpiresAt: now.Add(10 * time.Minute).Unix(),
		NotBefore: now.Add(1 * time.Minute).Unix(),
	}

	assert.NoError(t, validateTemporalClaims(claims, now, testLeeway))
}

func TestValidateTemporalClaims_NotBeforeBeyondLeeway_Invalid(t *testing.T) {
	now := time.Now()
	claims := accessTokenClaims{
		ExpiresAt: now.Add(10 * time.Minute).Unix(),
		NotBefore: now.Add(5 * time.Minute).Unix(),
	}

	err := validateTemporalClaims(claims, now, testLeeway)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet valid")
}

func TestValidateTemporalClaims_NotBeforeAbsent_Valid(t *testing.T) {
	now := time.Now()
	claims := accessTokenClaims{
		ExpiresAt: now.Add(10 * time.Minute).Unix(),
	}

	assert.NoError(t, validateTemporalClaims(claims, now, testLeeway))
}
