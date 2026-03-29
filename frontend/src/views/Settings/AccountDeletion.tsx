import React from 'react'
import {
  Box,
  Typography,
  Divider,
  Button,
  Modal,
  ModalDialog,
  ModalClose,
  DialogTitle,
  DialogContent,
  DialogActions,
  Alert,
} from '@mui/joy'
import { WarningAmberRounded } from '@mui/icons-material'
import { connect } from 'react-redux'
import { AppDispatch, RootState } from '@/store/store'
import {
  requestAccountDeletion,
  cancelAccountDeletion,
} from '@/store/userSlice'

type AccountDeletionProps = {
  deletionRequestedAt: string | null
  deletionStatus: 'loading' | 'succeeded' | 'failed' | null
  deletionError: string | null
  requestDeletion: () => void
  cancelDeletion: () => void
}

type AccountDeletionState = {
  confirmOpen: boolean
}

class AccountDeletionImpl extends React.Component<
  AccountDeletionProps,
  AccountDeletionState
> {
  constructor(props: AccountDeletionProps) {
    super(props)
    this.state = { confirmOpen: false }
  }

  private openConfirm = () => this.setState({ confirmOpen: true })
  private closeConfirm = () => this.setState({ confirmOpen: false })

  private onConfirmDelete = () => {
    this.closeConfirm()
    this.props.requestDeletion()
  }

  private formatDeletionTime(iso: string): string {
    const requested = new Date(iso)
    const deleteAt = new Date(requested.getTime() + 24 * 60 * 60 * 1000)
    return deleteAt.toLocaleString()
  }

  render(): React.ReactNode {
    const { deletionRequestedAt, deletionStatus, deletionError, cancelDeletion } =
      this.props
    const { confirmOpen } = this.state
    const isPending = !!deletionRequestedAt
    const isLoading = deletionStatus === 'loading'

    return (
      <Box sx={{ mt: 2 }}>
        <Typography level='h3' color='danger'>
          Account Deletion
        </Typography>
        <Divider />

        {deletionError && (
          <Alert color='danger' sx={{ mt: 1 }}>
            {deletionError}
          </Alert>
        )}

        {isPending ? (
          <Box sx={{ mt: 1 }}>
            <Alert
              color='warning'
              startDecorator={<WarningAmberRounded />}
              sx={{ mb: 1 }}
            >
              Your account is scheduled for permanent deletion on{' '}
              <strong>{this.formatDeletionTime(deletionRequestedAt!)}</strong>.
              All your tasks, labels, and data will be permanently erased. You
              can cancel below to restore full access.
            </Alert>
            <Button
              color='neutral'
              variant='outlined'
              loading={isLoading}
              onClick={cancelDeletion}
            >
              Cancel Deletion
            </Button>
          </Box>
        ) : (
          <Box sx={{ mt: 1 }}>
            <Typography sx={{ mb: 1 }}>
              Permanently delete your account and all associated data. This
              action initiates a 24-hour grace period during which you can
              cancel. After 24 hours, all data will be irrecoverably deleted.
            </Typography>
            <Button
              color='danger'
              variant='outlined'
              loading={isLoading}
              onClick={this.openConfirm}
            >
              Delete My Account
            </Button>
          </Box>
        )}

        <Modal open={confirmOpen} onClose={this.closeConfirm}>
          <ModalDialog variant='outlined' role='alertdialog'>
            <ModalClose />
            <DialogTitle>
              <WarningAmberRounded color='error' />
              Confirm Account Deletion
            </DialogTitle>
            <DialogContent>
              Are you sure you want to delete your account? Your account will be
              locked immediately and all data will be permanently deleted after
              24 hours. You can cancel during this period.
            </DialogContent>
            <DialogActions>
              <Button color='danger' onClick={this.onConfirmDelete}>
                Yes, delete my account
              </Button>
              <Button
                variant='plain'
                color='neutral'
                onClick={this.closeConfirm}
              >
                Cancel
              </Button>
            </DialogActions>
          </ModalDialog>
        </Modal>
      </Box>
    )
  }
}

const mapStateToProps = (state: RootState) => ({
  deletionRequestedAt: state.user.profile.deletion_requested_at,
  deletionStatus: state.user.deletionStatus,
  deletionError: state.user.deletionError,
})

const mapDispatchToProps = (dispatch: AppDispatch) => ({
  requestDeletion: () => dispatch(requestAccountDeletion()),
  cancelDeletion: () => dispatch(cancelAccountDeletion()),
})

export const AccountDeletion = connect(
  mapStateToProps,
  mapDispatchToProps,
)(AccountDeletionImpl)
