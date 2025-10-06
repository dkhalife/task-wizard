import { ChangePassword } from '@/api/auth'
import { Logo } from '@/Logo'
import { validatePassword } from '@/models/user'
import { getQuery, NavigationPaths, WithNavigate } from '@/utils/navigation'
import { Sheet } from '@mui/joy'
import {
  Container,
  Box,
  Typography,
  FormControl,
  Input,
  FormHelperText,
  Button,
} from '@mui/joy'
import React, { ChangeEvent } from 'react'
import { connect } from 'react-redux'
import { AppDispatch } from '@/store/store'
import { pushStatus } from '@/store/statusSlice'
import { StatusSeverity } from '@/models/status'

type UpdatePasswordViewProps = WithNavigate & {
  pushStatus: (message: string, severity: StatusSeverity, timeout?: number) => void
}

interface UpdatePasswordViewState {
  password: string
  passwordConfirm: string
  passwordError: string | null
  passwordConfirmationError: string | null
}

class UpdatePasswordViewImpl extends React.Component<
  UpdatePasswordViewProps,
  UpdatePasswordViewState
> {
  private navigationTimeout?: NodeJS.Timeout

  constructor(props: UpdatePasswordViewProps) {
    super(props)

    this.state = {
      password: '',
      passwordConfirm: '',
      passwordError: null,
      passwordConfirmationError: null,
    }
  }

  componentWillUnmount(): void {
    if (this.navigationTimeout) {
      clearTimeout(this.navigationTimeout)
    }
  }

  private handlePasswordChange = (e: ChangeEvent<HTMLInputElement>) => {
    const password = e.target.value

    if (!validatePassword(password)) {
      this.setState({
        password,
        passwordError: 'Password must be at least 8 characters',
      })
    } else {
      this.setState({
        password,
        passwordError: null,
      })
    }
  }

  private onCancel = () => {
    this.props.navigate(NavigationPaths.Login)
  }

  private handlePasswordConfirmChange = (e: ChangeEvent<HTMLInputElement>) => {
    const { password } = this.state
    if (e.target.value !== password) {
      this.setState({
        passwordConfirm: e.target.value,
        passwordConfirmationError: 'Passwords do not match',
      })
    } else {
      this.setState({
        passwordConfirm: e.target.value,
        passwordConfirmationError: null,
      })
    }
  }

  private handleSubmit = async () => {
    const { password, passwordError, passwordConfirmationError } = this.state

    if (passwordError != null || passwordConfirmationError != null) {
      return
    }

    try {
      const verificationCode = getQuery('c')
      await ChangePassword(verificationCode, password)
      this.props.pushStatus('Password updated successfully', 'success', 3000)
      this.navigationTimeout = setTimeout(() => this.props.navigate(NavigationPaths.Login), 3000)
    } catch {
      this.props.pushStatus('Password update failed, try again later', 'error', 6000)
    }
  }

  render(): React.ReactNode {
    const {
      password,
      passwordConfirm,
      passwordError,
      passwordConfirmationError,
    } = this.state

    return (
      <Container
        component='main'
        maxWidth='xs'
      >
        <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            marginTop: 4,
          }}
        >
          <Sheet
            component='form'
            sx={{
              mt: 1,
              width: '100%',
              display: 'flex',
              flexDirection: 'column',
              padding: 2,
              borderRadius: '8px',
              boxShadow: 'md',
            }}
          >
            <Box
              sx={{
                display: 'flex',
                alignItems: 'center',
                flexDirection: 'column',
              }}
            >
              <Logo />
              <Typography mb={4}>
                Please enter your new password below
              </Typography>
            </Box>

            <FormControl error>
              <Input
                placeholder='Password'
                type='password'
                value={password}
                onChange={this.handlePasswordChange}
                error={passwordError !== null}
              />
              <FormHelperText>{passwordError}</FormHelperText>
            </FormControl>

            <FormControl error>
              <Input
                placeholder='Confirm Password'
                type='password'
                value={passwordConfirm}
                onChange={this.handlePasswordConfirmChange}
                error={passwordConfirmationError !== null}
              />
              <FormHelperText>{passwordConfirmationError}</FormHelperText>
            </FormControl>

            <Button
              fullWidth
              size='lg'
              sx={{
                mt: 5,
                mb: 1,
              }}
              onClick={this.handleSubmit}
            >
              Save Password
            </Button>
            <Button
              fullWidth
              size='lg'
              variant='soft'
              onClick={this.onCancel}
            >
              Cancel
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

export const UpdatePasswordView = connect(null, mapDispatchToProps)(UpdatePasswordViewImpl)
