import { Logo } from '@/Logo'
import { Sheet } from '@mui/joy'
import {
  Container,
  Box,
  Typography,
  Input,
  Button,
  Divider,
} from '@mui/joy'
import React, { ChangeEvent } from 'react'
import { doLogin } from '@/utils/auth'
import { setTitle } from '@/utils/dom'
import { NavigationPaths, WithNavigate } from '@/utils/navigation'
import { connect } from 'react-redux'
import { AppDispatch, RootState } from '@/store/store'
import { pushStatus } from '@/store/statusSlice'
import { StatusSeverity } from '@/models/status'
import {
  getOAuthConfig,
  getOAuthConfigFromEnv,
  initiateOAuth,
  isOAuthConfiguredViaEnv,
  storeOAuthState,
} from '@/utils/oauth'

type LoginViewProps = WithNavigate & {
  pushStatus: (message: string, severity: StatusSeverity, timeout?: number) => void
  useOAuth: boolean
}

interface LoginViewState {
  email: string
  password: string
  oauthAvailable: boolean
  oauthLoading: boolean
}

class LoginViewImpl extends React.Component<LoginViewProps, LoginViewState> {
  constructor(props: LoginViewProps) {
    super(props)

    this.state = {
      email: '',
      password: '',
      oauthAvailable: false,
      oauthLoading: false,
    }
  }

  async componentDidMount(): Promise<void> {
    setTitle('Login')

    // Check if OAuth is enabled via feature flag
    if (this.props.useOAuth) {
      // First check if OAuth is configured via environment variables
      if (isOAuthConfiguredViaEnv()) {
        this.setState({ oauthAvailable: true })
      } else {
        // Otherwise, check with the backend
        try {
          const config = await getOAuthConfig()
          if (config && config.enabled) {
            this.setState({ oauthAvailable: true })
          }
        } catch (error) {
          console.error('Failed to check OAuth configuration:', error)
        }
      }
    }
  }

  private handleSubmit = async (e: React.MouseEvent<HTMLAnchorElement> | React.FormEvent) => {
    e.preventDefault()

    try {
      const { email, password } = this.state
      await doLogin(email, password, this.props.navigate)
    } catch (error) {
      this.props.pushStatus((error as Error).message, 'error', 5000)
    }
  }

  private onEmailChange = (e: ChangeEvent<HTMLInputElement>) => {
    this.setState({ email: e.target.value })
  }

  private onPasswordChange = (e: ChangeEvent<HTMLInputElement>) => {
    this.setState({ password: e.target.value })
  }

  private handleOAuthLogin = async () => {
    this.setState({ oauthLoading: true })
    try {
      // Check if OAuth is configured via environment variables
      const envConfig = getOAuthConfigFromEnv()
      if (envConfig) {
        // Use environment variable configuration
        const state = crypto.getRandomValues(new Uint8Array(32)).reduce((acc, val) => acc + val.toString(16).padStart(2, '0'), '')
        storeOAuthState(state)
        
        const authUrl = `${envConfig.authorize_url}?client_id=${encodeURIComponent(envConfig.client_id)}&redirect_uri=${encodeURIComponent(envConfig.redirect_url)}&response_type=code&scope=${encodeURIComponent(envConfig.scope)}&state=${state}`
        window.location.href = authUrl
      } else {
        // Use backend-provided configuration
        const response = await initiateOAuth()
        storeOAuthState(response.state)
        window.location.href = response.authorization_url
      }
    } catch (error) {
      this.setState({ oauthLoading: false })
      this.props.pushStatus((error as Error).message, 'error', 5000)
    }
  }

  render(): React.ReactNode {
    const { navigate, useOAuth } = this.props
    const { oauthAvailable, oauthLoading } = this.state

    return (
      <Container
        component='main'
        maxWidth='xs'
      >
        <Box
          sx={{
            marginTop: 4,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
          }}
        >
          <Sheet
            component='form'
            onSubmit={this.handleSubmit}
            sx={{
              mt: 1,
              width: '100%',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              padding: 2,
              borderRadius: '8px',
              boxShadow: 'md',
            }}
          >
            <Logo />

            <Typography>Sign in to your account to continue</Typography>

            {useOAuth && oauthAvailable ? (
              <>
                <Button
                  type='button'
                  fullWidth
                  size='lg'
                  variant='solid'
                  loading={oauthLoading}
                  sx={{
                    width: '100%',
                    mt: 4,
                    border: 'moccasin',
                    borderRadius: '8px',
                  }}
                  onClick={this.handleOAuthLogin}
                >
                  Sign in with OAuth
                </Button>
                <Divider sx={{ my: 2 }}> or </Divider>
              </>
            ) : null}

            <Typography
              alignSelf={'start'}
              mt={useOAuth && oauthAvailable ? 0 : 4}
            >
              Email
            </Typography>
            <Input
              required
              fullWidth
              autoComplete='email'
              autoFocus
              onChange={this.onEmailChange}
            />
            <Typography alignSelf={'start'}>Password:</Typography>
            <Input
              required
              fullWidth
              type='password'
              onChange={this.onPasswordChange}
            />

            <Button
              type='submit'
              fullWidth
              size='lg'
              variant='solid'
              sx={{
                width: '100%',
                mt: 3,
                border: 'moccasin',
                borderRadius: '8px',
              }}
              onClick={this.handleSubmit}
            >
              Sign In
            </Button>
            <Button
              type='button'
              fullWidth
              size='lg'
              variant='plain'
              sx={{
                width: '100%',
                border: 'moccasin',
                borderRadius: '8px',
              }}
              onClick={() => navigate(NavigationPaths.ResetPassword)}
            >
              Forgot password?
            </Button>

            <Divider> or </Divider>
            <Button
              onClick={() => navigate(NavigationPaths.Register)}
              fullWidth
              variant='soft'
              size='lg'
              sx={{
                mt: 2,
              }}
            >
              Create new account
            </Button>
          </Sheet>
        </Box>
      </Container>
    )
  }
}

const mapStateToProps = (state: RootState) => ({
  useOAuth: state.featureFlags.useOAuth,
})

const mapDispatchToProps = (dispatch: AppDispatch) => ({
  pushStatus: (message: string, severity: StatusSeverity, timeout?: number) =>
    dispatch(pushStatus({ message, severity, timeout })),
})

export const LoginView = connect(
  mapStateToProps,
  mapDispatchToProps,
)(LoginViewImpl)
