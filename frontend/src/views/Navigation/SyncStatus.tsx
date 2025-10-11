import { CheckCircleOutline, ErrorOutline } from '@mui/icons-material'
import { CircularProgress, IconButton, Modal, ModalDialog, Typography, Box, Divider } from '@mui/joy'
import React from 'react'
import { connect } from 'react-redux'
import { RootState } from '@/store/store'
import { SyncState } from '@/models/sync'

interface SyncStatusProps {
  websocketsEnabled: boolean

  style?: React.CSSProperties
  userStatus: SyncState
  userError: string | null
  tasksStatus: SyncState
  tasksError: string | null
  labelsStatus: SyncState
  labelsError: string | null
  tokensStatus: SyncState
  tokensError: string | null
  wsStatus: SyncState
  wsError: string | null
}

interface SyncStatusState {
  modalOpen: boolean
}

function getIcon(status: SyncState): React.ReactNode {
  switch (status) {
    case 'failed':
      return <ErrorOutline color='error' />
    case 'succeeded':
      return <CheckCircleOutline color='success' />
    default:
      return <CircularProgress size='sm' />
  }
}

class SyncStatusImpl extends React.Component<SyncStatusProps, SyncStatusState> {
  constructor(props: SyncStatusProps) {
    super(props)
    this.state = {
      modalOpen: false,
    }
  }

  private handleIconClick = () => {
    this.setState({
      modalOpen: true,
    })
  }

  private handleModalClose = () => {
    this.setState({
      modalOpen: false,
    })
  }

  private getStatuses = () => {
    const {
      userStatus,
      tasksStatus,
      labelsStatus,
      tokensStatus,
      userError,
      tasksError,
      labelsError,
      tokensError,
      wsStatus,
      wsError,
    } = this.props

    const statuses = [
      { name: 'User', status: userStatus, error: userError },
      { name: 'Tasks', status: tasksStatus, error: tasksError },
      { name: 'Labels', status: labelsStatus, error: labelsError },
      { name: 'Tokens', status: tokensStatus, error: tokensError },
    ]

    if (this.props.websocketsEnabled) {
      statuses.push({ name: 'WebSocket', status: wsStatus, error: wsError })
    }

    return statuses
  }

  private renderStatusDetails = () => {
    const statuses = this.getStatuses()
    const hasErrors = statuses.some(s => s.error)

    return (
      <Box>
        {statuses.map((s) => (
          <Box key={s.name} sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
            <Typography level="body-sm" sx={{ minWidth: '60px' }}>
              {s.name}:
            </Typography>
            {getIcon(s.status)}
          </Box>
        ))}
        {hasErrors && (
          <>
            <Divider sx={{ my: 2 }} />
            <Typography level="title-sm" sx={{ mb: 1 }}>
              Errors:
            </Typography>
            {statuses.filter(s => s.error).map(s => (
              <Box key={`error-${s.name}`} sx={{ mb: 1 }}>
                <Typography level="body-sm" color="danger">
                  <strong>{s.name}:</strong> {s.error}
                </Typography>
              </Box>
            ))}
          </>
        )}
      </Box>
    )
  }
  render(): React.ReactNode {
    const statuses = this.getStatuses()
    const anyFailed = statuses.some(s => s.status === 'failed')
    const allSucceeded = statuses.every(s => s.status === 'succeeded')
    const icon = getIcon(anyFailed ? 'failed' : allSucceeded ? 'succeeded' : 'loading')

    return (
      <>
        <IconButton 
          variant='plain' 
          style={this.props.style}
          onClick={this.handleIconClick}
        >
          {icon}
        </IconButton>
        
        <Modal
          open={this.state.modalOpen}
          onClose={this.handleModalClose}
        >
          <ModalDialog
            sx={{
              minWidth: '120px',
              maxWidth: '90vw',
            }}
          >
            <Typography level="title-lg" sx={{ mb: 1 }}>
              Sync Status
            </Typography>
            {this.renderStatusDetails()}
          </ModalDialog>
        </Modal>
      </>
    )
  }
}

const mapStateToProps = (state: RootState) => ({
  websocketsEnabled: state.featureFlags.useWebsockets,
  userStatus: state.user.status,
  userError: state.user.error,
  tasksStatus: state.tasks.status,
  tasksError: state.tasks.error,
  labelsStatus: state.labels.status,
  labelsError: state.labels.error,
  tokensStatus: state.tokens.status,
  tokensError: state.tokens.error,
  wsStatus: state.ws.status,
  wsError: state.ws.error,
})

export const SyncStatus = connect(mapStateToProps)(SyncStatusImpl)
