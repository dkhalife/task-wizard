import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import { SyncState } from '@/models/sync'
import {
  registerWebSocketListeners as registerLabelsWs,
  unregisterWebSocketListeners as unregisterLabelsWs,
} from './labelsSlice'
import {
  registerWebSocketListeners as registerNotificationsWs,
  unregisterWebSocketListeners as unregisterNotificationsWs,
} from './userSlice'
import {
  registerWebSocketListeners as registerTasksWs,
  unregisterWebSocketListeners as unregisterTasksWs,
} from './tasksSlice'
import {
  registerActivityWebSocketListeners as registerActivityWs,
  unregisterActivityWebSocketListeners as unregisterActivityWs,
} from './activitySlice'
import WebSocketManager from '@/utils/websocket'

export interface WSState {
  status: SyncState
  error: string | null
  lastUpdated: number | null
}

const initialState: WSState = {
  status: 'loading',
  error: null,
  lastUpdated: null,
}

const wsSlice = createSlice({
  name: 'ws',
  initialState,
  reducers: {
    wsConnecting: (state) => {
      state.status = 'loading'
      state.error = null
      state.lastUpdated = null
    },
    wsConnected: (state) => {
      state.status = 'succeeded'
      state.error = null
      state.lastUpdated = Date.now()

      const mgr = WebSocketManager.getInstance()
      registerLabelsWs(mgr)
      registerNotificationsWs(mgr)
      registerTasksWs(mgr)
      registerActivityWs(mgr)
    },
    wsDisconnected: (state, action: PayloadAction<string | null>) => {
      state.status = 'failed'
      state.error = action.payload
      state.lastUpdated = Date.now()

      const mgr = WebSocketManager.getInstance()
      unregisterLabelsWs(mgr)
      unregisterNotificationsWs(mgr)
      unregisterTasksWs(mgr)
      unregisterActivityWs(mgr)
    },
  },
})

export const wsReducer = wsSlice.reducer
export const { wsConnecting, wsConnected, wsDisconnected } = wsSlice.actions
