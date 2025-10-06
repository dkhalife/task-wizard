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
} from '@mui/joy'
import React, { ChangeEvent } from 'react'
import { connect } from 'react-redux'
import { AppDispatch } from '@/store/store'
import { pushStatus } from '@/store/statusSlice'

type SignupViewProps = WithNavigate & {
  pushStatus: (message: string, severity: 'error' | 'success' | 'info' | 'warning', timeout?: number) => void
}

interface SignupViewState {
  password: string
  displayName: string
  email: string
  passwordError: string | null
  emailError: string | null
  displayNameError: string | null
  accountCreated: boolean
}

class SignupViewImpl extends React.Component<
  SignupViewProps,
  SignupViewState
> {
  private navigationTimeout?: NodeJS.Timeout

  constructor(props: SignupViewProps) {
    super(props)

    this.state = {
      password: '',
      displayName: '',
      email: '',
      passwordError: null,
      emailError: null,
      displayNameError: null,
      accountCreated: false,
    }
  }

  componentDidMount(): void {
    setTitle('Sign Up')
  }

  componentWillUnmount(): void {
    if (this.navigationTimeout) {
      clearTimeout(this.navigationTimeout)
    }
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
      this.props.pushStatus('Please check your email to verify your account.', 'success', 3000)
      this.navigationTimeout = setTimeout(() => this.props.navigate(NavigationPaths.Login), 3000)
    } catch (error) {
      this.props.pushStatus((error as Error).message, 'error', 5000)
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
      password,
      displayName,
      email,
      passwordError,
      displayNameError,
      emailError,
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
      </Container>
    )
  }
}

const mapDispatchToProps = (dispatch: AppDispatch) => ({
  pushStatus: (message: string, severity: 'error' | 'success' | 'info' | 'warning', timeout?: number) =>
    dispatch(pushStatus({ message, severity, timeout })),
})

export const SignupView = connect(null, mapDispatchToProps)(SignupViewImpl)
