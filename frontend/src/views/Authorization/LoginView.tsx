import { Logo } from '@/Logo'
import { Sheet } from '@mui/joy'
import {
  Container,
  Box,
  Typography,
  Button,
} from '@mui/joy'
import React from 'react'
import { setTitle } from '@/utils/dom'
import { NavigationPaths, WithNavigate } from '@/utils/navigation'
import { connect } from 'react-redux'
import { AppDispatch } from '@/store/store'
import { pushStatus } from '@/store/statusSlice'
import { StatusSeverity } from '@/models/status'
import { loginSilently, loginWithPopup } from '@/utils/msal'

type LoginViewProps = WithNavigate & {
  pushStatus: (message: string, severity: StatusSeverity, timeout?: number) => void
}

type LoginViewState = {
  attemptedSilentLogin: boolean
}

class LoginViewImpl extends React.Component<LoginViewProps, LoginViewState> {
  constructor(props: LoginViewProps) {
    super(props)

    this.state = {
      attemptedSilentLogin: false,
    }
  }

  async componentDidMount(): Promise<void> {
    setTitle('Login')

    if (!this.state.attemptedSilentLogin) {
      try {
        await loginSilently();
        this.props.navigate(NavigationPaths.HomeView())
      } catch {
        // Expected failure
      }

      this.setState({
        attemptedSilentLogin: true,
      })
    }
  }

  private handleLogin = async () => {
    try {
      await loginWithPopup();
      this.props.pushStatus('Login successful!', "success", 2000);
      this.props.navigate(NavigationPaths.HomeView());
    } catch (error) {
      this.props.pushStatus('Login failed. Please try again.', "error", 5000);
      console.error('Login error:', error);
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
            marginTop: 4,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
          }}
        >
          <Sheet
            component='form'
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

            <Button
              type='button'
              fullWidth
              size='lg'
              variant='solid'
              sx={{
                width: '100%',
                mt: 3,
                border: 'moccasin',
                borderRadius: '8px',
              }}
              onClick={this.handleLogin}
            >
              Sign In
            </Button>
          </Sheet>
        </Box>
      </Container>
    )
  }
}

const mapDispatchToProps = (dispatch: AppDispatch) => ({
  pushStatus: (message: string, severity: StatusSeverity, timeout?: number) =>
    dispatch(pushStatus({ message, severity, timeout })),
})

export const LoginView = connect(null, mapDispatchToProps)(LoginViewImpl)
