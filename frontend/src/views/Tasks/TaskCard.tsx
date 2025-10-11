import {
  TASK_UPDATE_EVENT,
  getDueDateChipColor,
  getDueDateChipText,
  getRecurrentChipText,
} from '@/models/task'
import { getTextColorFromBackgroundColor } from '@/utils/colors'
import { playSound, SoundEffect } from '@/utils/sound'
import {
  TimesOneMobiledata,
  Repeat,
  Check,
  NotificationsActive,
} from '@mui/icons-material'
import {
  Box,
  Chip,
  Card,
  Grid,
  Typography,
  IconButton,
} from '@mui/joy'
import React from 'react'
import { NavigationPaths, WithNavigate } from '@/utils/navigation'
import { connect } from 'react-redux'
import { completeTask } from '@/store/tasksSlice'
import { AppDispatch } from '@/store/store'
import { TaskUI } from '@/utils/marshalling'

type TaskCardProps = WithNavigate & {
  task: TaskUI

  completeTask: (taskId: number) => Promise<any>
  onTaskUpdate: (event: TASK_UPDATE_EVENT) => void
  onContextMenu: (event: React.MouseEvent<HTMLDivElement>, task: TaskUI) => void
}

class TaskCardImpl extends React.Component<TaskCardProps> {
  private getFrequencyIcon = (task: TaskUI) => {
    if (task.frequency.type === 'once') {
      return <TimesOneMobiledata />
    } else {
      return <Repeat />
    }
  }

  private handleTaskCompletion = async () => {
    const { task, onTaskUpdate } = this.props
    await this.props.completeTask(task.id)

    // Play the task completion sound
    playSound(SoundEffect.TaskComplete)
    
    onTaskUpdate('completed')
  }

  private hasAnyNotificationsActive = () => {
    const { task } = this.props
    if (!task.notification.enabled) {
      return false
    }

    const notifications = task.notification
    return (
      notifications.due_date || notifications.overdue || notifications.pre_due
    )
  }

  render(): React.ReactNode {
    const { task, navigate } = this.props

    const notificationsActive = this.hasAnyNotificationsActive()

    return (
      <Box key={task.id + '-box'}>
        <Chip
          variant='soft'
          sx={{
            position: 'relative',
            top: 10,
            zIndex: 1,
            left: 10,
          }}
          color={getDueDateChipColor(task.next_due_date)}
        >
          {getDueDateChipText(task.next_due_date)}
        </Chip>

        <Chip
          variant='soft'
          sx={{
            position: 'relative',
            top: 10,
            zIndex: 1,
            ml: 0.4,
            left: 10,
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            {this.getFrequencyIcon(task)}
            {getRecurrentChipText(task.next_due_date, task.frequency)}
          </div>
        </Chip>

        {notificationsActive && (
          <Chip
            variant='soft'
            sx={{
              position: 'relative',
              top: 10,
              zIndex: 1,
              ml: 0.4,
              left: 10,
            }}
          >
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <NotificationsActive />
            </div>
          </Chip>
        )}

        <Card
          variant='plain'
          sx={{
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'space-between',
            p: 2,
            boxShadow: 'sm',
            borderRadius: 20,
            key: `${task.id}-card`,
          }}
        >
          <Grid container sx={{
            display: 'flex',
            flexDirection: 'row',
          }}>
            <Grid
              sx={{
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'center',
              }}
            >
              <Box
                display='flex'
                justifyContent='center'
                alignItems='center'
              >
                <IconButton
                  variant='solid'
                  color='success'
                  onClick={this.handleTaskCompletion}
                  sx={{
                    borderRadius: '50%',
                    minWidth: 50,
                    height: 50,
                    zIndex: 1,
                  }}
                >
                  <div
                    style={{
                      position: 'relative',
                      alignItems: 'center',
                    }}
                  >
                    <Check />
                  </div>
                </IconButton>
              </Box>
            </Grid>
            <Grid
              style={{
                cursor: 'pointer',
                flex: 1,
              }}
              onContextMenu={(e) => this.props.onContextMenu(e, task)}
              onClick={() => navigate(NavigationPaths.TaskEdit(task.id))}
            >
              <Box
                display='flex'
                justifyContent='start'
                alignItems='center'
              >
                <Box
                  sx={{
                    display: 'flex',
                    flex: 1,
                    flexDirection: 'column',
                    ml: 2,
                  }}
                >
                  <Typography level='title-md'>{task.title}</Typography>
                  <Box key={`${task.id}-labels`}>
                    {task.labels?.map((l, index) => {
                      return (
                        <Chip
                          variant='solid'
                          key={`taskcard-${task.id}-label-${l.id}`}
                          sx={{
                            position: 'relative',
                            ml: index === 0 ? 0 : 0.5,
                            top: 2,
                            zIndex: 1,
                            backgroundColor: `${l?.color}`,
                            color: getTextColorFromBackgroundColor(l?.color),
                          }}
                        >
                          {l?.name}
                        </Chip>
                      )
                    })}
                  </Box>
                </Box>
              </Box>
            </Grid>
          </Grid>
        </Card>
      </Box>
    )
  }
}

const mapDispatchToProps = (dispatch: AppDispatch) => ({
  completeTask: (taskId: number) => dispatch(completeTask(taskId)),
})

export const TaskCard = connect(
  null,
  mapDispatchToProps,
)(TaskCardImpl)
