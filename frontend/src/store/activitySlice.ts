import { createAsyncThunk, createSlice } from '@reduxjs/toolkit'
import { GetActivity, UncompleteTask } from '@/api/tasks'
import { ActivityEntry } from '@/models/activity'
import { SyncState } from '@/models/sync'
import WebSocketManager from '@/utils/websocket'
import { store } from './store'

export const ACTIVITY_PAGE_SIZE = 20

export interface ActivityState {
  items: ActivityEntry[]
  loaded: boolean
  hasMore: boolean
  status: SyncState
  error: string | null
}

const initialState: ActivityState = {
  items: [],
  loaded: false,
  hasMore: false,
  status: 'loading',
  error: null,
}

export const fetchActivity = createAsyncThunk(
  'activity/fetchActivity',
  async () => {
    const data = await GetActivity(0, ACTIVITY_PAGE_SIZE)
    return data.activity
  },
)

export const loadMoreActivity = createAsyncThunk(
  'activity/loadMoreActivity',
  async (_, thunkAPI) => {
    const state = thunkAPI.getState() as { activity: ActivityState }
    const items = state.activity.items
    const beforeId = items.length > 0 ? items[items.length - 1].id : 0
    const data = await GetActivity(beforeId, ACTIVITY_PAGE_SIZE)
    return data.activity
  },
)

export const revertAction = createAsyncThunk(
  'activity/revertAction',
  async (
    { taskId, historyId }: { taskId: number; historyId: number },
    thunkAPI,
  ) => {
    await UncompleteTask(taskId, historyId)

    // When connected over WebSocket, the server broadcasts task_uncompleted and
    // the feed refreshes via the listener. Over plain HTTP there is no such
    // event, so refresh here to reflect the revert.
    if (!WebSocketManager.getInstance().isConnected()) {
      await thunkAPI.dispatch(fetchActivity())
    }
  },
)

const activitySlice = createSlice({
  name: 'activity',
  initialState,
  reducers: {},
  extraReducers: builder => {
    builder
      .addCase(fetchActivity.pending, state => {
        state.status = 'loading'
        state.error = null
      })
      .addCase(fetchActivity.fulfilled, (state, action) => {
        state.status = 'succeeded'
        state.items = action.payload
        state.hasMore = action.payload.length === ACTIVITY_PAGE_SIZE
        state.loaded = true
        state.error = null
      })
      .addCase(fetchActivity.rejected, (state, action) => {
        state.status = 'failed'
        state.error = action.error.message ?? null
      })
      .addCase(loadMoreActivity.pending, state => {
        state.status = 'loading'
        state.error = null
      })
      .addCase(loadMoreActivity.fulfilled, (state, action) => {
        state.status = 'succeeded'
        const existingIds = new Set(state.items.map(e => e.id))
        const newEntries = action.payload.filter(e => !existingIds.has(e.id))
        state.items.push(...newEntries)
        state.hasMore = action.payload.length === ACTIVITY_PAGE_SIZE
        state.error = null
      })
      .addCase(loadMoreActivity.rejected, (state, action) => {
        state.status = 'failed'
        state.error = action.error.message ?? null
      })
  },
})

export const activityReducer = activitySlice.reducer

const onActivityChanged = () => {
  if (!store.getState().activity.loaded) {
    return
  }
  store.dispatch(fetchActivity())
}

export const registerActivityWebSocketListeners = (ws: WebSocketManager) => {
  ws.on('task_completed', onActivityChanged)
  ws.on('task_uncompleted', onActivityChanged)
  ws.on('task_skipped', onActivityChanged)
  ws.on('task_deleted', onActivityChanged)
}

export const unregisterActivityWebSocketListeners = (ws: WebSocketManager) => {
  ws.off('task_completed', onActivityChanged)
  ws.off('task_uncompleted', onActivityChanged)
  ws.off('task_skipped', onActivityChanged)
  ws.off('task_deleted', onActivityChanged)
}
