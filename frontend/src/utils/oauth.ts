import { Request } from './api'

export interface OAuthConfig {
  enabled: boolean
  client_id: string
  authorize_url: string
  scope: string
  redirect_url: string
}

export interface OAuthInitiateResponse {
  authorization_url: string
  state: string
}

export interface OAuthTokenResponse {
  token: string
  expiration: string
}

// Get OAuth configuration from the backend
export const getOAuthConfig = async (): Promise<OAuthConfig | null> => {
  try {
    const config = await Request<OAuthConfig>(
      '/oauth/config',
      'GET',
      undefined,
      false,
    )
    return config
  } catch (error) {
    console.error('Failed to get OAuth config:', error)
    return null
  }
}

// Initiate OAuth flow - get authorization URL
export const initiateOAuth = async (): Promise<OAuthInitiateResponse> => {
  return await Request<OAuthInitiateResponse>(
    '/oauth/authorize',
    'GET',
    undefined,
    false,
  )
}

// Complete OAuth flow by exchanging code for token
export const completeOAuth = async (
  code: string,
): Promise<OAuthTokenResponse> => {
  return await Request<OAuthTokenResponse>(
    `/oauth/callback?code=${encodeURIComponent(code)}`,
    'GET',
    undefined,
    false,
  )
}

// Generate a random string for state parameter
export const generateState = (): string => {
  const array = new Uint8Array(32)
  crypto.getRandomValues(array)
  return Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join(
    '',
  )
}

// Store OAuth state in session storage
export const storeOAuthState = (state: string): void => {
  sessionStorage.setItem('oauth_state', state)
}

// Retrieve and validate OAuth state from session storage
export const validateOAuthState = (state: string): boolean => {
  const storedState = sessionStorage.getItem('oauth_state')
  sessionStorage.removeItem('oauth_state')
  return storedState === state
}

// Check if OAuth is configured via environment variables
export const isOAuthConfiguredViaEnv = (): boolean => {
  const enabled = import.meta.env.VITE_OAUTH_ENABLED === 'true'
  const clientId = import.meta.env.VITE_OAUTH_CLIENT_ID
  const authority = import.meta.env.VITE_OAUTH_AUTHORITY
  const scope = import.meta.env.VITE_OAUTH_SCOPE

  return enabled && !!clientId && !!authority && !!scope
}

// Get OAuth configuration from environment variables
export const getOAuthConfigFromEnv = (): OAuthConfig | null => {
  if (!isOAuthConfiguredViaEnv()) {
    return null
  }

  return {
    enabled: true,
    client_id: import.meta.env.VITE_OAUTH_CLIENT_ID,
    authorize_url: import.meta.env.VITE_OAUTH_AUTHORITY,
    scope: import.meta.env.VITE_OAUTH_SCOPE,
    redirect_url: import.meta.env.VITE_OAUTH_REDIRECT_URI || window.location.origin + '/oauth/callback',
  }
}
