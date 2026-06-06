package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"taskwiz.app/core/config"
)

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
