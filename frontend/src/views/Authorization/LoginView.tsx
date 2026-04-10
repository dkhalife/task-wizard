import { Loading } from '@/Loading'
import { Logo } from '@/Logo'
import { Sheet } from '@mui/joy'
import {
  Container,
  Box,
  Typography,
  Button,
} from '@mui/joy'
import React from 'react'
import { Link } from 'react-router-dom'
import { hasCachedAccounts, initializeMsal, loginSilently, loginWithRedirect } from '@/utils/msal'
import { setTitle } from '@/utils/dom'
import { getQuery, NavigationPaths, WithNavigate } from '@/utils/navigation'
import { connect } from 'react-redux'
import { AppDispatch } from '@/store/store'
import { pushStatus } from '@/store/statusSlice'
import { StatusSeverity } from '@/models/status'

type LoginViewProps = WithNavigate & {
  pushStatus: (message: string, severity: StatusSeverity, timeout?: number) => void
}

type LoginViewState = {
  authReady: boolean
}

class LoginViewImpl extends React.Component<LoginViewProps, LoginViewState> {
  constructor(props: LoginViewProps) {
    super(props)
    this.state = { authReady: false }
  }

  private getReturnPath = (): string => {
    const returnTo = getQuery('return_to')
    return returnTo || NavigationPaths.HomeView()
  }

  async componentDidMount(): Promise<void> {
    setTitle('Login')
    await initializeMsal()
    const silentOk = await loginSilently()
    if (silentOk) {
      this.props.navigate(this.getReturnPath())
      return
    }
    try {
      if (await hasCachedAccounts()) {
        await loginWithRedirect()
        return
      }
    } catch (error) {
      this.props.pushStatus((error as Error).message, 'error', 5000)
    }
    this.setState({ authReady: true })
  }

  private handleLogin = async () => {
    try {
      await loginWithRedirect()
    } catch (error) {
      this.props.pushStatus((error as Error).message, 'error', 5000)
    }
  }

  render(): React.ReactNode {
    if (!this.state.authReady) {
      return <Loading />
    }

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

            <Typography sx={{ mt: 2 }}>Sign in to your account to continue</Typography>

            <Button
              fullWidth
              size='lg'
              variant='solid'
              sx={{
                width: '100%',
                mt: 3,
                borderRadius: '8px',
              }}
              onClick={this.handleLogin}
            >
              Sign in
            </Button>
          </Sheet>
          <Typography
            level='body-xs'
            sx={{ mt: 2, textAlign: 'center' }}
          >
            <Link
              to={NavigationPaths.Privacy}
              style={{ color: 'inherit' }}
            >
              Privacy Policy
            </Link>
          </Typography>
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
