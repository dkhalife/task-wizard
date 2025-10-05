import { Box, CircularProgress, Container } from '@mui/joy'
import { Typography } from '@mui/joy'
import React from 'react'

type LoadingProps = object

interface LoadingState {
  subMessage: string
}

export class Loading extends React.Component<LoadingProps, LoadingState> {
  private timeout: number = -1

  constructor(props: LoadingProps) {
    super(props)

    this.state = {
      subMessage: '',
    }
  }

  componentDidMount() {
    this.timeout = window.setTimeout(() => {
      this.setState({
        subMessage:
          'This is taking longer than usual. There might be an issue.',
      })
    }, 5000)
  }

  componentWillUnmount() {
    if (this.timeout !== -1) {
      clearTimeout(this.timeout)
    }
  }

  render() {
    const { subMessage } = this.state

    return (
      <Container
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100%',
        }}
      >
        <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: '80vh',
          }}
        >
          <CircularProgress
            color='success'
            sx={{ '--CircularProgress-size': '50px' }}
          />
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              gap: 2,
              fontWeight: 700,
              fontSize: 24,
              mt: 2,
            }}
          >
            Loading...
          </Box>
          <Typography
            fontWeight={500}
            textAlign={'center'}
          >
            {subMessage}
          </Typography>
        </Box>
      </Container>
    )
  }
}
