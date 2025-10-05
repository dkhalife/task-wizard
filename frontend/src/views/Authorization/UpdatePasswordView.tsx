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
  Snackbar,
} from '@mui/joy'
import React, { ChangeEvent } from 'react'

type UpdatePasswordViewProps = WithNavigate

interface UpdatePasswordViewState {
  password: string
  passwordConfirm: string
  passwordError: string | null
  passwordConfirmationError: string | null
  updateStatusOk: boolean | null
}

export class UpdatePasswordView extends React.Component<
  UpdatePasswordViewProps,
  UpdatePasswordViewState
> {
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

  private onSnackbarClose = () => {
    this.props.navigate(NavigationPaths.Login)
  }

  private onCancel = () => {
    this.setState({
      updateStatusOk: null,
    })
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
      this.setState({
        updateStatusOk: true,
      })

      this.props.navigate(NavigationPaths.Login)
    } catch {
      this.setState({
        updateStatusOk: false,
      })
    }
  }

  render(): React.ReactNode {
    const {
      password,
      passwordConfirm,
      passwordError,
      passwordConfirmationError,
      updateStatusOk,
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
        <Snackbar
          open={updateStatusOk !== true}
          autoHideDuration={6000}
          onClose={this.onSnackbarClose}
        >
          Password update failed, try again later
        </Snackbar>
      </Container>
    )
  }
}
