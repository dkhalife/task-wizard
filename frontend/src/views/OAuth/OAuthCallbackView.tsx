import { Box, CircularProgress, Container, Typography } from '@mui/joy'
import React from 'react'
import { NavigationPaths, WithNavigate } from '@/utils/navigation'
import { completeOAuth, validateOAuthState } from '@/utils/oauth'
import { setTitle } from '@/utils/dom'

type OAuthCallbackViewProps = WithNavigate

class OAuthCallbackViewImpl extends React.Component<OAuthCallbackViewProps> {
  componentDidMount(): void {
    setTitle('OAuth Callback')
    this.handleCallback()
  }

  private handleCallback = async () => {
    try {
      // Extract code and state from URL
      const urlParams = new URLSearchParams(window.location.search)
      const code = urlParams.get('code')
      const state = urlParams.get('state')
      const error = urlParams.get('error')
      const errorDescription = urlParams.get('error_description')

      // Check for errors from OAuth provider
      if (error) {
        console.error('OAuth error:', error, errorDescription)
        this.props.navigate(
          NavigationPaths.Login +
            '?error=' +
            encodeURIComponent(errorDescription || error),
        )
        return
      }

      // Validate required parameters
      if (!code || !state) {
        console.error('Missing code or state in OAuth callback')
        this.props.navigate(
          NavigationPaths.Login +
            '?error=' +
            encodeURIComponent('Invalid OAuth callback'),
        )
        return
      }

      // Validate state to prevent CSRF
      if (!validateOAuthState(state)) {
        console.error('Invalid OAuth state')
        this.props.navigate(
          NavigationPaths.Login +
            '?error=' +
            encodeURIComponent('Invalid OAuth state'),
        )
        return
      }

      // Exchange code for token
      const response = await completeOAuth(code)
      localStorage.setItem('ca_token', response.token)
      localStorage.setItem('ca_expiration', response.expiration)

      // Navigate to home or redirect URL
      const redirectUrl = localStorage.getItem('ca_redirect')
      if (redirectUrl) {
        localStorage.removeItem('ca_redirect')
        this.props.navigate(redirectUrl)
      } else {
        this.props.navigate(NavigationPaths.HomeView())
      }
    } catch (error) {
      console.error('Failed to complete OAuth flow:', error)
      this.props.navigate(
        NavigationPaths.Login +
          '?error=' +
          encodeURIComponent('Failed to complete authentication'),
      )
    }
  }

  render(): React.ReactNode {
    return (
      <Container
        component='main'
        maxWidth='xs'
      >
        <Box
          sx={{
            marginTop: 8,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
          }}
        >
          <CircularProgress size='lg' />
          <Typography
            level='h4'
            sx={{ mt: 2 }}
          >
            Completing sign in...
          </Typography>
        </Box>
      </Container>
    )
  }
}

export const OAuthCallbackView = OAuthCallbackViewImpl
