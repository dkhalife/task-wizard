import { deleteTask, completeTask, updateDueDate, filterTasks, filterCompletedTasks, fetchCompletedTasks } from '@/store/tasksSlice'
import { getDueDateChipColor, getDueDateChipText } from '@/models/task'
import {
  CancelRounded,
  SearchRounded,
  CheckBox,
  Edit,
  History,
  CalendarMonth,
  Delete,
  Visibility,
  VisibilityOff,
} from '@mui/icons-material'
import {
  Container,
  Grid,
  Input,
  Button,
  Table,
  Chip,
  ButtonGroup,
  IconButton,
} from '@mui/joy'
import React, { ChangeEvent } from 'react'
import { DateModal } from '@/views/Modals/Inputs/DateModal'
import { NavigationPaths, WithNavigate } from '@/utils/navigation'
import { setTitle } from '@/utils/dom'
import { getTextColorFromBackgroundColor } from '@/utils/colors'
import { ConfirmationModal } from '../Modals/Inputs/ConfirmationModal'
import { moveFocusToJoyInput } from '@/utils/joy'
import { playSound, SoundEffect } from '@/utils/sound'
import { AppDispatch, RootState } from '@/store/store'
import { connect } from 'react-redux'
import { TaskUI, MakeTaskUI, MarshallDate } from '@/utils/marshalling'
import { pushStatus } from '@/store/statusSlice'
import { Status } from '@/models/status'

type TasksOverviewProps = {
  tasks: TaskUI[]
  completedTasks: TaskUI[]

  search: string
  completedSearch: string
  filterTasks: (searchQuery: string) => void
  filterCompletedTasks: (searchQuery: string) => void
  fetchCompletedTasks: () => Promise<any>
  completeTask: (taskId: number) => Promise<any>
  deleteTask: (taskId: number) => Promise<any>
  updateDueDate: (taskId: number, dueDate: string) => Promise<any>
  pushStatus: (status: Status) => void
} & WithNavigate

interface TasksOverviewState {
  showCompleted: boolean
}

class TasksOverviewImpl extends React.Component<TasksOverviewProps, TasksOverviewState> {
  private dateModalRef = React.createRef<DateModal>()
  private confirmationModalRef = React.createRef<ConfirmationModal>()
  private searchInputRef = React.createRef<HTMLInputElement>()
  private completedSearchInputRef = React.createRef<HTMLInputElement>()

  constructor(props: TasksOverviewProps) {
    super(props)
    this.state = {
      showCompleted: false,
    }
  }

  componentDidMount(): void {
    setTitle('Tasks Overview')

    this.registerKeyboardShortcuts()
  }

  componentWillUnmount(): void {
    this.unregisterKeyboardShortcuts()
  }

  private registerKeyboardShortcuts = () => {
    document.addEventListener('keydown', this.onKeyDown)
  }

  private unregisterKeyboardShortcuts = () => {
    document.removeEventListener('keydown', this.onKeyDown)
  }

  private onKeyDown = (event: KeyboardEvent) => {
    // Ctrl + F => Search
    if (event.key === 'f' && event.ctrlKey) {
      moveFocusToJoyInput(this.searchInputRef)
      event.preventDefault()
    }

    if (
      event.key === '+' &&
      document.activeElement?.tagName.toLowerCase() !== 'input'
    ) {
      this.onAddTaskClicked()
      event.preventDefault()
      event.stopPropagation()
    }
  }

  private onCompleteTaskClicked = (task: TaskUI) => async () => {
    await this.props.completeTask(task.id)

    playSound(SoundEffect.TaskComplete)
    this.props.pushStatus({
      message: 'Task completed',
      severity: 'success',
      timeout: 3000,
    })
  }

  private onSearchTermsChanged = (e: ChangeEvent<HTMLInputElement>) => {
    this.props.filterTasks(e.target.value)
  }

  private onCompletedSearchTermsChanged = (e: ChangeEvent<HTMLInputElement>) => {
    this.props.filterCompletedTasks(e.target.value)
  }

  private onClearSearchClicked = () => {
    this.props.filterTasks('')
  }

  private onClearCompletedSearchClicked = () => {
    this.props.filterCompletedTasks('')
  }

  private onToggleCompletedClicked = async () => {
    const { showCompleted } = this.state
    const { completedTasks } = this.props
    
    if (!showCompleted && completedTasks.length === 0) {
      await this.props.fetchCompletedTasks()
    }
    this.setState({ showCompleted: !showCompleted })
  }

  private onRescheduleClicked = async (task: TaskUI) => {
    this.dateModalRef.current?.open(task.next_due_date, async (newDate: Date | null) => {
      if (newDate === null) {
        return
      }

      await this.props.updateDueDate(task.id, MarshallDate(newDate))

      this.props.pushStatus({
        message: 'Task rescheduled',
        severity: 'success',
        timeout: 3000,
      })
    })
  }

  private onViewHistoryClicked = (task: TaskUI) => {
    const { navigate } = this.props
    navigate(NavigationPaths.TaskHistory(task.id))
  }

  private onAddTaskClicked = () => {
    const { navigate } = this.props
    navigate(NavigationPaths.TaskCreate)
  }

  private onDeleteTaskClicked = async (task: TaskUI) => {
    this.confirmationModalRef.current?.open(async () => {
      await this.props.deleteTask(task.id)

      this.props.pushStatus({
        message: 'Task deleted',
        severity: 'success',
        timeout: 3000,
      })
    })
  }

  render(): React.ReactNode {
    const { search, completedSearch, tasks, completedTasks, navigate } = this.props
    const { showCompleted } = this.state

    return (
      <Container>
        <Grid container>
          <Grid
            sm={6}
            alignSelf={'flex-start'}
            minWidth={100}
            display='flex'
            gap={2}
          >
            <Input
              ref={this.searchInputRef}
              placeholder='Search'
              value={search}
              onChange={this.onSearchTermsChanged}
              endDecorator={
                search !== '' ? (
                  <Button onClick={this.onClearSearchClicked}>
                    <CancelRounded />
                  </Button>
                ) : (
                  <Button>
                    <SearchRounded />
                  </Button>
                )
              }
            />
          </Grid>
          <Grid
            sm={6}
            justifyContent={'flex-end'}
            display={'flex'}
            gap={2}
          >
            <Button
              onClick={this.onToggleCompletedClicked}
              startDecorator={showCompleted ? <VisibilityOff /> : <Visibility />}
            >
              {showCompleted ? 'Hide' : 'Show'} Completed
            </Button>
            <Button onClick={this.onAddTaskClicked}>New Task</Button>
          </Grid>
        </Grid>

        <Table
          sx={{
            '& tbody tr:hover': {
              backgroundColor: 'var(--joy-palette-background-level2)',
            },
          }}
        >
          <thead>
            <tr>
              <th>Due</th>
              <th>Title</th>
              <th>Labels</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {tasks.map((task: TaskUI) => (
              <tr key={task.id}>
                <td>
                  <Chip color={getDueDateChipColor(task.next_due_date)}>
                    {getDueDateChipText(task.next_due_date)}
                  </Chip>
                </td>
                <td
                  style={{
                    cursor: 'pointer',
                  }}
                  onClick={() => navigate(NavigationPaths.TaskEdit(task.id))}
                >
                  {task.title || '--'}
                </td>
                <td>
                  {task.labels.map((l, index) => {
                    return (
                      <Chip
                        variant='solid'
                        key={`taskcard-${task.id}-label-${l.id}`}
                        color='primary'
                        sx={{
                          position: 'relative',
                          ml: index === 0 ? 0 : 0.5,
                          top: 2,
                          zIndex: 1,
                          backgroundColor: `${l.color} !important`,
                          color: getTextColorFromBackgroundColor(l.color),
                        }}
                      >
                        {l?.name}
                      </Chip>
                    )
                  })}
                </td>
                <td>
                  <ButtonGroup>
                    <IconButton
                      variant='outlined'
                      size='sm'
                      onClick={this.onCompleteTaskClicked(task)}
                      aria-setsize={2}
                      sx={{
                        '&:hover': {
                          color: 'var(--joy-palette-success-500)',
                          '& svg': { color: 'var(--joy-palette-success-500)' },
                        },
                      }}
                    >
                      <CheckBox />
                    </IconButton>
                    <IconButton
                      variant='outlined'
                      size='sm'
                      onClick={() => this.onRescheduleClicked(task)}
                      aria-setsize={2}
                      sx={{
                        '&:hover': {
                          color: 'var(--joy-palette-primary-500)',
                          '& svg': { color: 'var(--joy-palette-primary-500)' },
                        },
                      }}
                    >
                      <CalendarMonth />
                    </IconButton>
                    <IconButton
                      variant='outlined'
                      size='sm'
                      onClick={() => this.onViewHistoryClicked(task)}
                      aria-setsize={2}
                      sx={{
                        '&:hover': {
                          color: 'var(--joy-palette-secondary-500)',
                          '& svg': { color: 'var(--joy-palette-secondary-500)' },
                        },
                      }}
                    >
                      <History />
                    </IconButton>
                    <IconButton
                      variant='outlined'
                      size='sm'
                      onClick={() =>
                        navigate(NavigationPaths.TaskEdit(task.id))
                      }
                      sx={{
                        '&:hover': {
                          color: 'var(--joy-palette-warning-500)',
                          '& svg': { color: 'var(--joy-palette-warning-500)' },
                        },
                      }}
                    >
                      <Edit />
                    </IconButton>
                    <IconButton
                      variant='outlined'
                      size='sm'
                      onClick={() => this.onDeleteTaskClicked(task)}
                      aria-setsize={2}
                      sx={{
                        '&:hover': {
                          color: 'var(--joy-palette-danger-500)',
                          '& svg': { color: 'var(--joy-palette-danger-500)' },
                        },
                      }}
                    >
                      <Delete />
                    </IconButton>
                  </ButtonGroup>
                </td>
              </tr>
            ))}
          </tbody>
        </Table>

        {showCompleted && (
          <>
            <Grid
              container
              sx={{ mt: 4, mb: 2 }}
            >
              <Grid
                sm={6}
                alignSelf={'flex-start'}
                minWidth={100}
                display='flex'
                gap={2}
              >
                <Input
                  ref={this.completedSearchInputRef}
                  placeholder='Search Completed Tasks'
                  value={completedSearch}
                  onChange={this.onCompletedSearchTermsChanged}
                  endDecorator={
                    completedSearch !== '' ? (
                      <Button onClick={this.onClearCompletedSearchClicked}>
                        <CancelRounded />
                      </Button>
                    ) : (
                      <Button>
                        <SearchRounded />
                      </Button>
                    )
                  }
                />
              </Grid>
            </Grid>

            <Table
              sx={{
                '& tbody tr:hover': {
                  backgroundColor: 'var(--joy-palette-background-level2)',
                },
              }}
            >
              <thead>
                <tr>
                  <th>Due</th>
                  <th>Title</th>
                  <th>Labels</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {completedTasks.map((task: TaskUI) => (
                  <tr key={task.id}>
                    <td>
                      <Chip color={getDueDateChipColor(task.next_due_date)}>
                        {getDueDateChipText(task.next_due_date)}
                      </Chip>
                    </td>
                    <td
                      style={{
                        cursor: 'pointer',
                      }}
                      onClick={() => navigate(NavigationPaths.TaskEdit(task.id))}
                    >
                      {task.title || '--'}
                    </td>
                    <td>
                      {task.labels.map((l, index) => {
                        return (
                          <Chip
                            variant='solid'
                            key={`taskcard-${task.id}-label-${l.id}`}
                            color='primary'
                            sx={{
                              position: 'relative',
                              ml: index === 0 ? 0 : 0.5,
                              top: 2,
                              zIndex: 1,
                              backgroundColor: `${l.color} !important`,
                              color: getTextColorFromBackgroundColor(l.color),
                            }}
                          >
                            {l?.name}
                          </Chip>
                        )
                      })}
                    </td>
                    <td>
                      <ButtonGroup>
                        <IconButton
                          variant='outlined'
                          size='sm'
                          onClick={() => this.onViewHistoryClicked(task)}
                          aria-setsize={2}
                          sx={{
                            '&:hover': {
                              color: 'var(--joy-palette-secondary-500)',
                              '& svg': { color: 'var(--joy-palette-secondary-500)' },
                            },
                          }}
                        >
                          <History />
                        </IconButton>
                      </ButtonGroup>
                    </td>
                  </tr>
                ))}
              </tbody>
            </Table>
          </>
        )}
        <DateModal
          ref={this.dateModalRef}
          title={`Change due date`}
        />
        <ConfirmationModal
          ref={this.confirmationModalRef}
          title='Delete Task'
          confirmText='Delete'
          cancelText='Cancel'
          message='Are you sure you want to delete this task?'
        />
      </Container>
    )
  }
}

const mapStateToProps = (state: RootState) => {
  const search = state.tasks.searchQuery
  const completedSearch = state.tasks.completedSearchQuery
  const labels = state.labels.items
  const tasks = state.tasks.filteredItems.map(task => MakeTaskUI(task, labels))
  const completedTasks = state.tasks.filteredCompletedItems.map(task => MakeTaskUI(task, labels))

  return {
    search,
    completedSearch,
    tasks,
    completedTasks,
  }
}

const mapDispatchToProps = (dispatch: AppDispatch) => ({
  filterTasks: (searchQuery: string) => dispatch(filterTasks(searchQuery)),
  filterCompletedTasks: (searchQuery: string) => dispatch(filterCompletedTasks(searchQuery)),
  fetchCompletedTasks: () => dispatch(fetchCompletedTasks()),
  completeTask: (taskId: number) => dispatch(completeTask(taskId)),
  deleteTask: (taskId: number) => dispatch(deleteTask(taskId)),
  updateDueDate: (taskId: number, dueDate: string) => dispatch(updateDueDate({ taskId, dueDate })),
  pushStatus: (status: Status) => dispatch(pushStatus(status)),
})

export const TasksOverview = connect(
  mapStateToProps,
  mapDispatchToProps,
)(TasksOverviewImpl)
