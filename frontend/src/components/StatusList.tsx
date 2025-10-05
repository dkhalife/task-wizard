import React from 'react'
import { connect } from 'react-redux'
import { Snackbar } from '@mui/joy'
import { RootState, AppDispatch } from '@/store/store'
import { StatusSeverity, TrackedStatus } from '@/models/status'
import { dismissStatus } from '@/store/statusSlice'

interface StatusListProps {
  statuses: TrackedStatus[]
  dismiss: (id: string) => void
}

class StatusListImpl extends React.Component<StatusListProps> {
  private mapSeverity = (severity: StatusSeverity) => {
    switch (severity) {
      case 'error':
        return 'danger'
      case 'success':
        return 'success'
      case 'warning':
        return 'warning'
      default:
        return 'primary'
    }
  }

  render(): React.ReactNode {
    const { statuses, dismiss } = this.props

    return (
      <>
        {statuses.map(status => (
          <Snackbar
            key={status.id}
            open={true}
            color={this.mapSeverity(status.severity)}
            onClose={() => dismiss(status.id)}
          >
            {status.message}
          </Snackbar>
        ))}
      </>
    )
  }
}

const mapStateToProps = (state: RootState) => ({
  statuses: state.status.items,
})

const mapDispatchToProps = (dispatch: AppDispatch) => ({
  dismiss: (id: string) => dispatch(dismissStatus(id)),
})

export const StatusList = connect(
  mapStateToProps,
  mapDispatchToProps,
)(StatusListImpl)
