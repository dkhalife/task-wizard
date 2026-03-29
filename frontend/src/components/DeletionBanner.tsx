import React from 'react'
import { Alert, Button, Typography } from '@mui/joy'
import { WarningAmberRounded } from '@mui/icons-material'
import { connect } from 'react-redux'
import { Link } from 'react-router-dom'
import { AppDispatch, RootState } from '@/store/store'
import { cancelAccountDeletion } from '@/store/userSlice'
import { NavigationPaths } from '@/utils/navigation'

type DeletionBannerProps = {
  deletionRequestedAt: string | null
  deletionStatus: 'loading' | 'succeeded' | 'failed' | null
  cancelDeletion: () => void
}

class DeletionBannerImpl extends React.Component<DeletionBannerProps> {
  private formatDeletionTime(iso: string): string {
    const requested = new Date(iso)
    const deleteAt = new Date(requested.getTime() + 24 * 60 * 60 * 1000)
    return deleteAt.toLocaleString()
  }

  render(): React.ReactNode {
    const { deletionRequestedAt, deletionStatus, cancelDeletion } = this.props
    if (!deletionRequestedAt) return null

    return (
      <Alert
        color='warning'
        variant='solid'
        startDecorator={<WarningAmberRounded />}
        endDecorator={
          <Button
            size='sm'
            variant='outlined'
            color='warning'
            loading={deletionStatus === 'loading'}
            onClick={cancelDeletion}
          >
            Cancel Deletion
          </Button>
        }
        sx={{ borderRadius: 0 }}
      >
        <Typography level='body-sm'>
          Your account is scheduled for deletion on{' '}
          <strong>{this.formatDeletionTime(deletionRequestedAt)}</strong>.
          Writes are disabled. Visit{' '}
          <Link to={NavigationPaths.Settings} style={{ color: 'inherit' }}>
            Settings
          </Link>{' '}
          to manage this.
        </Typography>
      </Alert>
    )
  }
}

const mapStateToProps = (state: RootState) => ({
  deletionRequestedAt: state.user.profile.deletion_requested_at,
  deletionStatus: state.user.deletionStatus,
})

const mapDispatchToProps = (dispatch: AppDispatch) => ({
  cancelDeletion: () => dispatch(cancelAccountDeletion()),
})

export const DeletionBanner = connect(
  mapStateToProps,
  mapDispatchToProps,
)(DeletionBannerImpl)
