import { Container, Typography, Sheet, List, Button, Box } from '@mui/joy'
import { EventBusy } from '@mui/icons-material'
import React from 'react'
import { connect } from 'react-redux'
import { Loading } from '@/Loading'
import { ActivityCard } from './ActivityCard'
import { setTitle } from '@/utils/dom'
import { ActivityEntryUI, MakeActivityUI } from '@/utils/marshalling'
import { AppDispatch, RootState } from '@/store/store'
import { fetchActivity, loadMoreActivity, revertAction } from '@/store/activitySlice'
import { pushStatus } from '@/store/statusSlice'
import { Status } from '@/models/status'
import { SyncState } from '@/models/sync'
import { NavigationPaths, WithNavigate } from '@/utils/navigation'
import { playSound, SoundEffect } from '@/utils/sound'

type ActivityProps = {
  entries: ActivityEntryUI[]
  status: SyncState
  hasMore: boolean

  fetchActivity: () => Promise<unknown>
  loadMoreActivity: () => Promise<unknown>
  revertAction: (taskId: number, historyId: number) => Promise<unknown>
  pushStatus: (status: Status) => void
} & WithNavigate

class ActivityImpl extends React.Component<ActivityProps> {
  componentDidMount(): void {
    setTitle('Activity')
    void this.props.fetchActivity()
  }

  private onRevert = async (entry: ActivityEntryUI) => {
    try {
      await this.props.revertAction(entry.task_id, entry.id)

      playSound(SoundEffect.TaskComplete)
      this.props.pushStatus({
        message: 'Action reverted',
        severity: 'success',
        timeout: 3000,
      })
    } catch (e) {
      this.props.pushStatus({
        message: e instanceof Error ? e.message : 'Unable to revert action',
        severity: 'warning',
        timeout: 4000,
      })

      // Refresh so the feed reflects the latest revertable state.
      void this.props.fetchActivity()
    }
  }

  private onTitleClick = (taskId: number) => {
    this.props.navigate(NavigationPaths.TaskHistory(taskId))
  }

  render(): React.ReactNode {
    const { entries, status, hasMore } = this.props

    if (status === 'loading' && entries.length === 0) {
      return <Loading />
    }

    if (entries.length === 0) {
      return (
        <Container
          maxWidth='md'
          sx={{
            textAlign: 'center',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexDirection: 'column',
            height: '50vh',
          }}
        >
          <EventBusy sx={{ fontSize: '6rem', mb: 1 }} />
          <Typography
            level='h3'
            gutterBottom
          >
            No Recent Activity
          </Typography>
          <Typography>
            Completed and skipped tasks will show up here, with the option to
            revert your most recent action on each task.
          </Typography>
        </Container>
      )
    }

    return (
      <Container maxWidth='md'>
        <Typography
          level='h4'
          my={1.5}
        >
          Recent Activity
        </Typography>
        <Sheet sx={{ borderRadius: 'sm', p: 2, boxShadow: 'md' }}>
          <List sx={{ p: 0 }}>
            {entries.map(entry => (
              <ActivityCard
                key={`activity-${entry.id}`}
                entry={entry}
                onRevert={this.onRevert}
                onTitleClick={this.onTitleClick}
              />
            ))}
          </List>
        </Sheet>
        {hasMore && (
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'center',
              mt: 2,
            }}
          >
            <Button
              variant='soft'
              loading={status === 'loading'}
              onClick={() => this.props.loadMoreActivity()}
            >
              Load more
            </Button>
          </Box>
        )}
      </Container>
    )
  }
}

const mapStateToProps = (state: RootState) => ({
  entries: state.activity.items.map(MakeActivityUI),
  status: state.activity.status,
  hasMore: state.activity.hasMore,
})

const mapDispatchToProps = (dispatch: AppDispatch) => ({
  fetchActivity: () => dispatch(fetchActivity()),
  loadMoreActivity: () => dispatch(loadMoreActivity()),
  revertAction: (taskId: number, historyId: number) =>
    dispatch(revertAction({ taskId, historyId })).unwrap(),
  pushStatus: (status: Status) => dispatch(pushStatus(status)),
})

export const Activity = connect(
  mapStateToProps,
  mapDispatchToProps,
)(ActivityImpl)
