import { App } from '@/App'
import { INVALID_TASK_ID, newTask, Task } from '@/models/task'
import { AppDispatch, RootState } from '@/store/store'
import { setDraft } from '@/store/tasksSlice'
import { getHomeView, getPathName, NavigationPaths } from '@/utils/navigation'
import { ForgotPasswordView } from '@/views/Authorization/ForgotPasswordView'
import { LoginView } from '@/views/Authorization/LoginView'
import { SignupView } from '@/views/Authorization/Signup'
import { UpdatePasswordView } from '@/views/Authorization/UpdatePasswordView'
import { TaskHistory } from '@/views/History/TaskHistory'
import { LabelView } from '@/views/Labels/LabelView'
import { NotFound } from '@/views/NotFound'
import { Settings } from '@/views/Settings/Settings'
import { MyTasks } from '@/views/Tasks/MyTasks'
import { TaskEdit } from '@/views/Tasks/TaskEdit'
import { TasksOverview } from '@/views/Tasks/TasksOverview'
import React from 'react'
import { connect } from 'react-redux'
import {
  BrowserRouter,
  matchPath,
  Navigate,
  Route,
  Routes,
} from 'react-router-dom'

type RouterContextProps = {
  tasks: Task[]

  setDraft: (draft: Task) => void
}

type RouterContextState = {
  navigateTo: string | null
  lastTaskId: number
}

class RouterContextImpl extends React.Component<RouterContextProps, RouterContextState> {
  constructor(props: RouterContextProps) {
    super(props)

    this.state = {
      lastTaskId: this.getTaskId(),
      navigateTo: null,
    }
  }
  private getTaskId = (): number => {
    const match = matchPath<'taskId', string>(
      {
        path: '/tasks/:taskId',
        caseSensitive: true,
        end: false,
      },
      getPathName(),
    )

    if (!match) {
      return INVALID_TASK_ID
    }

    const maybeTaskId = parseInt(match.params.taskId as string, 10)
    if (isNaN(maybeTaskId)) {
      return INVALID_TASK_ID
    }

    return maybeTaskId
  }

  private getMainRoute = () => {
    const home_view = getHomeView()

    if (home_view === 'my_tasks') {
      return <MyTasks navigate={this.navigate} />
    }

    return <TasksOverview navigate={this.navigate} />
  }

  private navigate = (path: string) => {
    this.setState({
      navigateTo: path,
    })
  }

  private updateDraftState = (taskId: number) => {
    const { setDraft } = this.props

    if (taskId !== INVALID_TASK_ID) {
      const task = this.props.tasks.find((task) => task.id === taskId)
      if (task) {
        setDraft(task)
      } else {
        this.navigate(NavigationPaths.HomeView())
      }
    } else {
      setDraft(newTask())
    }
  }

  componentDidMount(): void {
    this.updateDraftState(this.state.lastTaskId)
  }

  componentDidUpdate(): void {
    if (this.state.navigateTo !== null) {
      this.setState({
        navigateTo: null,
      })
    }

    const taskId = this.getTaskId()
    if (taskId !== this.state.lastTaskId) {
      this.updateDraftState(taskId)
      this.setState({
        lastTaskId: taskId,
      })
    } else if (taskId === INVALID_TASK_ID) {
      // Always reset draft when navigating to create page, even if taskId hasn't changed
      this.updateDraftState(taskId)
    }
  }

  render() {
    const { navigateTo } = this.state

    return (
      <BrowserRouter>
        <Routes>
          <Route
            path='/'
            element={<App navigate={this.navigate} />}
            errorElement={<NotFound />}
          >
            <Route
              path='/'
              element={this.getMainRoute()}
            />
            <Route
              path='/settings'
              element={<Settings />}
            />
            <Route
              path='/tasks'
              element={<TasksOverview navigate={this.navigate} />}
            />
            <Route
              path='/tasks/:taskId/edit'
              element={
                <TaskEdit
                  navigate={this.navigate}
                />
              }
            />
            <Route
              path='/tasks/create'
              element={
                <TaskEdit
                  navigate={this.navigate}
                />
              }
            />
            <Route
              path='/tasks/:taskId/history'
              element={<TaskHistory />}
            />
            <Route
              path='/my/tasks'
              element={<MyTasks navigate={this.navigate} />}
            />
            <Route
              path='/login'
              element={<LoginView navigate={this.navigate} />}
            />
            <Route
              path='/signup'
              element={<SignupView navigate={this.navigate} />}
            />
            <Route
              path='/forgot-password'
              element={<ForgotPasswordView navigate={this.navigate} />}
            />
            <Route
              path='/password/update'
              element={<UpdatePasswordView navigate={this.navigate} />}
            />
            <Route
              path='/labels/'
              element={<LabelView />}
            />
          </Route>
        </Routes>

        {navigateTo && <Navigate to={navigateTo} />}
      </BrowserRouter>
    )
  }
}

const mapStateToProps = (state: RootState) => ({
  tasks: state.tasks.items,
})

const mapDispatchToProps = (dispatch: AppDispatch) => ({
  setDraft: (task: Task) => dispatch(setDraft(task)),
})

export const RouterContext = connect(
  mapStateToProps,
  mapDispatchToProps,
)(RouterContextImpl)
