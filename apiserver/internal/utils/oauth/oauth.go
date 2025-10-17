package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"dkhalife.com/tasks/core/config"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OAuthProvider wraps OAuth2 and OIDC functionality
type OAuthProvider struct {
	config   *oauth2.Config
	verifier *oidc.IDTokenVerifier
	cfg      *config.OAuthConfig
}

// NewOAuthProvider creates a new OAuth provider
func NewOAuthProvider(cfg *config.OAuthConfig) (*OAuthProvider, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	// Validate required fields
	if cfg.ClientID == "" {
		return nil, errors.New("oauth client_id is required")
	}
	if cfg.ClientSecret == "" {
		return nil, errors.New("oauth client_secret is required")
	}
	if cfg.AuthorizeURL == "" {
		return nil, errors.New("oauth authorize_url is required")
	}
	if cfg.TokenURL == "" {
		return nil, errors.New("oauth token_url is required")
	}
	if cfg.RedirectURL == "" {
		return nil, errors.New("oauth redirect_url is required")
	}
	if cfg.Scope == "" {
		return nil, errors.New("oauth scope is required")
	}

	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthorizeURL,
			TokenURL: cfg.TokenURL,
		},
		Scopes: []string{cfg.Scope},
	}

	// Setup OIDC verifier if JWKS URL is provided
	var verifier *oidc.IDTokenVerifier
	if cfg.JwksURL != "" {
		ctx := context.Background()
		keySet := oidc.NewRemoteKeySet(ctx, cfg.JwksURL)
		verifier = oidc.NewVerifier(cfg.ClientID, keySet, &oidc.Config{
			ClientID:          cfg.ClientID,
			SkipExpiryCheck:   false,
			SkipIssuerCheck:   true, // We may need to be flexible with issuer
			SkipClientIDCheck: false,
		})
	}

	return &OAuthProvider{
		config:   oauth2Config,
		verifier: verifier,
		cfg:      cfg,
	}, nil
}

// GenerateStateToken generates a random state token for OAuth2 flow
func GenerateStateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetAuthorizationURL returns the OAuth2 authorization URL
func (p *OAuthProvider) GetAuthorizationURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges an authorization code for tokens
func (p *OAuthProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

// ValidateToken validates an OAuth2 access token
func (p *OAuthProvider) ValidateToken(ctx context.Context, token string) (*Claims, error) {
	if p.verifier == nil {
		return nil, errors.New("token validation not configured")
	}

	// Verify the ID token
	idToken, err := p.verifier.Verify(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return &claims, nil
}

// Claims represents the claims in an OAuth2 token
type Claims struct {
	Email             string   `json:"email"`
	EmailVerified     bool     `json:"email_verified"`
	Name              string   `json:"name"`
	PreferredUsername string   `json:"preferred_username"`
	Sub               string   `json:"sub"`
	Roles             []string `json:"roles"`
	Scope             string   `json:"scp"`
	Aud               string   `json:"aud"`
	Exp               int64    `json:"exp"`
	Iat               int64    `json:"iat"`
	Iss               string   `json:"iss"`
}

// IsExpired checks if the token is expired
func (c *Claims) IsExpired() bool {
	return time.Now().Unix() > c.Exp
}
