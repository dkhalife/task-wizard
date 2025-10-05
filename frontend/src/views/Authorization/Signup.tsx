import { SignUp } from '@/api/auth'
import { Logo } from '@/Logo'
import { validateEmail, validatePassword } from '@/models/user'
import { setTitle } from '@/utils/dom'
import { NavigationPaths, WithNavigate } from '@/utils/navigation'
import {
  Container,
  Box,
  Typography,
  Input,
  FormControl,
  FormHelperText,
  Button,
  Sheet,
  Divider,
  Snackbar,
} from '@mui/joy'
import React, { ChangeEvent } from 'react'

type SignupViewProps = WithNavigate

interface SignupViewState {
  password: string
  displayName: string
  email: string
  passwordError: string | null
  emailError: string | null
  displayNameError: string | null
  error: string | null
  accountCreated: boolean
}

export class SignupView extends React.Component<
  SignupViewProps,
  SignupViewState
> {
  constructor(props: SignupViewProps) {
    super(props)

    this.state = {
      password: '',
      displayName: '',
      email: '',
      passwordError: null,
      emailError: null,
      displayNameError: null,
      error: null,
      accountCreated: false,
    }
  }

  componentDidMount(): void {
    setTitle('Sign Up')
  }

  private handleSignUpValidation = () => {
    // Reset errors before validation
    const newState: SignupViewState = {
      ...this.state,
      passwordError: null,
      displayNameError: null,
      emailError: null,
    }

    const { password, displayName, email } = this.state

    let isValid = true

    if (!validateEmail(email)) {
      newState.emailError = 'Invalid email address'
      isValid = false
    }

    if (!validatePassword(password)) {
      newState.passwordError = 'Password must be at least 8 characters'
      isValid = false
    }

    if (!displayName.trim()) {
      newState.displayNameError = 'Display name is required'
      isValid = false
    }

    // display name should only contain letters and spaces and numbers:
    if (!/^[a-zA-Z0-9 ]+$/.test(displayName)) {
      newState.displayNameError =
        'Display name can only contain letters, numbers and spaces'
      isValid = false
    }

    this.setState(newState)
    return isValid
  }

  private handleSubmit = async (e: React.MouseEvent) => {
    e.preventDefault()

    if (!this.handleSignUpValidation()) {
      return
    }

    const { password, displayName, email } = this.state
    try {
      await SignUp(password, displayName, email)
      this.setState({
        accountCreated: true,
      })
    } catch (error) {
      this.setState({
        error: (error as Error).message,
      })
    }
  }

  private onEmailChange = (e: ChangeEvent<HTMLInputElement>) => {
    this.setState({
      emailError: null,
      email: e.target.value.trim(),
    })
  }

  private onPasswordChange = (e: ChangeEvent<HTMLInputElement>) => {
    this.setState({
      passwordError: null,
      password: e.target.value,
    })
  }

  private onDisplayNameChange = (e: ChangeEvent<HTMLInputElement>) => {
    this.setState({
      displayNameError: null,
      displayName: e.target.value.trim(),
    })
  }

  render(): React.ReactNode {
    const {
      accountCreated,
      password,
      displayName,
      email,
      passwordError,
      displayNameError,
      emailError,
      error,
    } = this.state
    const { navigate } = this.props

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
              <Typography>Create an account to get started!</Typography>
            </Box>
            <Typography alignSelf={'start'}>Email</Typography>
            <Input
              required
              fullWidth
              autoComplete='email'
              value={email}
              onChange={this.onEmailChange}
            />
            <FormControl>
              <FormHelperText>{emailError}</FormHelperText>
            </FormControl>
            <Typography alignSelf={'start'}>Password:</Typography>
            <Input
              required
              fullWidth
              type='password'
              value={password}
              onChange={this.onPasswordChange}
            />
            <FormControl>
              <FormHelperText>{passwordError}</FormHelperText>
            </FormControl>
            <Typography alignSelf={'start'}>Display Name:</Typography>
            <Input
              required
              fullWidth
              value={displayName}
              onChange={this.onDisplayNameChange}
            />
            <FormControl>
              <FormHelperText>{displayNameError}</FormHelperText>
            </FormControl>
            <Button
              size='lg'
              fullWidth
              variant='solid'
              sx={{ mt: 3, mb: 1 }}
              onClick={this.handleSubmit}
            >
              Sign Up
            </Button>
            <Divider> or </Divider>
            <Button
              size='lg'
              onClick={() => navigate(NavigationPaths.Login)}
              fullWidth
              variant='soft'
            >
              Login
            </Button>
          </Sheet>
        </Box>
        <Snackbar
          open={error !== null}
          onClose={() => this.setState({ error: null })}
          autoHideDuration={5000}
        >
          {error}
        </Snackbar>
        <Snackbar
          color='success'
          variant='solid'
          open={accountCreated}
          onClose={() => navigate(NavigationPaths.Login)}
          autoHideDuration={3000}
        >
          Please check your email to verify your account.
        </Snackbar>
      </Container>
    )
  }
}
