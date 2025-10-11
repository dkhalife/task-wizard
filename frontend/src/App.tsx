import { NavBar } from './views/Navigation/NavBar'
import { Outlet } from 'react-router-dom'
import { isTokenValid } from './utils/api'
import React from 'react'
import { WithNavigate } from './utils/navigation'
import { CssBaseline, CssVarsProvider } from '@mui/joy'
import { preloadSounds } from './utils/sound'
import WebSocketManager from './utils/websocket'
import { fetchLabels } from './store/labelsSlice'
import { AppDispatch, RootState, store } from './store/store'
import { connect } from 'react-redux'
import { fetchUser } from './store/userSlice'
import { fetchTokens } from './store/tokensSlice'
import { StatusList } from './components/StatusList'
import { fetchTasks, initGroups } from './store/tasksSlice'
import { FIVE_MINUTES_MS } from '@/constants/time'

type AppProps = {
  refreshStaleData: boolean

  fetchLabels: () => Promise<any>
  fetchUser: () => Promise<any>
  fetchTasks: () => Promise<any>
  initGroups: () => void
  fetchTokens: () => Promise<any>
} & WithNavigate

class AppImpl extends React.Component<AppProps> {
  private onVisibilityChange = () => {
    if (!document.hidden) {
      this.refreshStaleData()
    }
  }

  private refreshStaleData = async () => {
    if (!this.props.refreshStaleData) {
      return
    }

    if (!isTokenValid()) {
      return
    }

    const state = store.getState()
    const now = Date.now()

    let groupsOutdated = false

    if (!state.user.lastFetched || now - state.user.lastFetched > FIVE_MINUTES_MS) {
      await this.props.fetchUser()
    }

    if (!state.labels.lastFetched || now - state.labels.lastFetched > FIVE_MINUTES_MS) {
      await this.props.fetchLabels()
      groupsOutdated = true
    }

    if (!state.tasks.lastFetched || now - state.tasks.lastFetched > FIVE_MINUTES_MS) {
      await this.props.fetchTasks()
      groupsOutdated = true
    }

    if (!state.tokens.lastFetched || now - state.tokens.lastFetched > FIVE_MINUTES_MS) {
      await this.props.fetchTokens()
    }

    if (groupsOutdated) {
      await this.props.initGroups()
    }
  }

  async componentDidMount(): Promise<void> {
    if (isTokenValid()) {
      preloadSounds();
      WebSocketManager.getInstance();

      await this.props.fetchUser()
      await this.props.fetchLabels()
      await this.props.fetchTasks()
      await this.props.fetchTokens()
      await this.props.initGroups()
    }

    document.addEventListener('visibilitychange', this.onVisibilityChange)
  }

  componentWillUnmount(): void {
    document.removeEventListener('visibilitychange', this.onVisibilityChange)
  }

  render() {
    const { navigate } = this.props

    return (
      <div style={{ minHeight: '100vh' }}>
        <CssBaseline />
        <CssVarsProvider
          modeStorageKey='themeMode'
          attribute='data-theme'
          defaultMode='system'
          colorSchemeNode={document.body}
        >
          <NavBar navigate={navigate} />
          <Outlet />
          <StatusList />
        </CssVarsProvider>
      </div>
    )
  }
}

const mapStateToProps = (state: RootState) => ({
  refreshStaleData: state.featureFlags.refreshStaleData,
})

const mapDispatchToProps = (dispatch: AppDispatch) => ({
  fetchUser: () => dispatch(fetchUser()),
  fetchLabels: () => dispatch(fetchLabels()),
  fetchTasks: () => dispatch(fetchTasks()),
  initGroups: () => dispatch(initGroups()),
  fetchTokens: () => dispatch(fetchTokens()),
})

export const App = connect(
  mapStateToProps,
  mapDispatchToProps,
)(AppImpl)
