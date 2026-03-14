import { Logo } from '@/Logo'
import { Sheet } from '@mui/joy'
import {
  Container,
  Box,
  Typography,
  Button,
} from '@mui/joy'
import React from 'react'
import { loginWithRedirect } from '@/utils/msal'
import { setTitle } from '@/utils/dom'
import { WithNavigate } from '@/utils/navigation'
import { connect } from 'react-redux'
import { AppDispatch } from '@/store/store'
import { pushStatus } from '@/store/statusSlice'
import { StatusSeverity } from '@/models/status'

type LoginViewProps = WithNavigate & {
  pushStatus: (message: string, severity: StatusSeverity, timeout?: number) => void
}

class LoginViewImpl extends React.Component<LoginViewProps> {
  componentDidMount(): void {
    setTitle('Login')
  }

  private handleLogin = async () => {
    try {
      await loginWithRedirect()
    } catch (error) {
      this.props.pushStatus((error as Error).message, 'error', 5000)
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
