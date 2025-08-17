import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit'
import { GetUserProfile, UpdateNotificationSettings } from '@/api/users'
import { User } from '@/models/user'
import { NotificationTriggerOptions, NotificationType } from '@/models/notifications'
import { SyncState } from '@/models/sync'
import { WSEventPayloads } from '@/models/websocket'
import WebSocketManager from '@/utils/websocket'
import { store } from './store'

export interface UserState {
  profile: User
  status: SyncState
  lastFetched: number | null
  error: string | null
}

const initialState: UserState = {
  profile: {
    display_name: '',
    notifications: {
      provider: {
        provider: 'none'
      },
      triggers: {
        pre_due: false,
        due_date: false,
        overdue: false,
      },
    },
  },
  status: 'loading',
  lastFetched: null,
  error: null,
}

export const fetchUser = createAsyncThunk('user/fetchUser', async () => {
  const data = await GetUserProfile()
  return data.user
})

export const updateNotificationSettings = createAsyncThunk(
  'user/updateNotificationSettings',
  async (settings: { type: NotificationType, options: NotificationTriggerOptions}) => await UpdateNotificationSettings(settings.type, settings.options))

const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    notificationSettingsUpdated(
      state,
      action: PayloadAction<{
        provider: NotificationType
        triggers: NotificationTriggerOptions
      }>,
    ) {
      state.profile.notifications.provider = action.payload.provider
      state.profile.notifications.triggers = action.payload.triggers
    },
  },
  extraReducers: builder => {
    builder
      .addCase(fetchUser.pending, state => {
        state.status = 'loading'
        state.error = null
      })
      .addCase(fetchUser.fulfilled, (state, action) => {
        state.status = 'succeeded'
        state.profile = action.payload
        state.lastFetched = Date.now()
        state.error = null
      })
      .addCase(fetchUser.rejected, (state, action) => {
        state.status = 'failed'
        state.error = action.error.message ?? null
      })
      .addCase(updateNotificationSettings.pending, state => {
        state.status = 'loading'
      })
      .addCase(updateNotificationSettings.fulfilled, (state, action) => {
        state.status = 'succeeded'
        userSlice.caseReducers.notificationSettingsUpdated(state, {
          payload: {
            provider: action.meta.arg.type,
            triggers: action.meta.arg.options,
          },
          type: 'user/notificationSettingsUpdated',
        })
        state.error = null
      })
      .addCase(updateNotificationSettings.rejected, (state, action) => {
        state.status = 'failed'
        state.error =
          action.error.message ??
          'An unknown error occurred while updating notification settings.'
      })
  },
})

export const userReducer = userSlice.reducer

const { notificationSettingsUpdated } = userSlice.actions

const onNotificationSettingsUpdated = (
  data: WSEventPayloads['notification_settings_updated'],
) => {
  store.dispatch(notificationSettingsUpdated(data))
}

export const registerWebSocketListeners = (ws: WebSocketManager) => {
  ws.on('notification_settings_updated', onNotificationSettingsUpdated)
}

export const unregisterWebSocketListeners = (ws: WebSocketManager) => {
  ws.off('notification_settings_updated', onNotificationSettingsUpdated)
}
