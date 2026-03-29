import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit'
import {
  GetUserProfile,
  UpdateNotificationSettings,
  RequestAccountDeletion,
  CancelAccountDeletion,
} from '@/api/users'
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
  draftNotificationSettings: {
    provider: NotificationType
    triggers: NotificationTriggerOptions
  }
  deletionStatus: SyncState | null
  deletionError: string | null
}

const initialState: UserState = {
  profile: {
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
    deletion_requested_at: null,
  },
  status: 'loading',
  lastFetched: null,
  error: null,
  draftNotificationSettings: {
    provider: {
      provider: 'none'
    },
    triggers: {
      pre_due: false,
      due_date: false,
      overdue: false,
    },
  },
  deletionStatus: null,
  deletionError: null,
}

export const fetchUser = createAsyncThunk('user/fetchUser', async () => {
  const data = await GetUserProfile()
  return data.user
})

export const updateNotificationSettings = createAsyncThunk(
  'user/updateNotificationSettings',
  async (settings: { type: NotificationType, options: NotificationTriggerOptions}) => await UpdateNotificationSettings(settings.type, settings.options))

export const requestAccountDeletion = createAsyncThunk(
  'user/requestAccountDeletion',
  async () => {
    await RequestAccountDeletion()
    return new Date().toISOString()
  },
)

export const cancelAccountDeletion = createAsyncThunk(
  'user/cancelAccountDeletion',
  async () => {
    await CancelAccountDeletion()
  },
)

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
      // Keep draft in sync when updated via websocket
      state.draftNotificationSettings.provider = action.payload.provider
      state.draftNotificationSettings.triggers = action.payload.triggers
    },
    setNotificationSettingsDraft(
      state,
      action: PayloadAction<{
        provider: NotificationType
        triggers: NotificationTriggerOptions
      }>,
    ) {
      state.draftNotificationSettings.provider = action.payload.provider
      state.draftNotificationSettings.triggers = action.payload.triggers
    },
    accountDeletionRequested(state) {
      state.profile.deletion_requested_at = new Date().toISOString()
    },
    accountDeletionCancelled(state) {
      state.profile.deletion_requested_at = null
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
        // Initialize draft with the fetched profile's notification settings
        state.draftNotificationSettings.provider = action.payload.notifications.provider
        state.draftNotificationSettings.triggers = action.payload.notifications.triggers
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
      .addCase(requestAccountDeletion.pending, state => {
        state.deletionStatus = 'loading'
        state.deletionError = null
      })
      .addCase(requestAccountDeletion.fulfilled, (state, action) => {
        state.deletionStatus = 'succeeded'
        state.profile.deletion_requested_at = action.payload
      })
      .addCase(requestAccountDeletion.rejected, (state, action) => {
        state.deletionStatus = 'failed'
        state.deletionError = action.error.message ?? 'Failed to request account deletion.'
      })
      .addCase(cancelAccountDeletion.pending, state => {
        state.deletionStatus = 'loading'
        state.deletionError = null
      })
      .addCase(cancelAccountDeletion.fulfilled, state => {
        state.deletionStatus = 'succeeded'
        state.profile.deletion_requested_at = null
      })
      .addCase(cancelAccountDeletion.rejected, (state, action) => {
        state.deletionStatus = 'failed'
        state.deletionError = action.error.message ?? 'Failed to cancel account deletion.'
      })
  },
})

export const { setNotificationSettingsDraft } = userSlice.actions

export const userReducer = userSlice.reducer

const {
  notificationSettingsUpdated,
  accountDeletionRequested,
  accountDeletionCancelled,
} = userSlice.actions

const onNotificationSettingsUpdated = (
  data: WSEventPayloads['notification_settings_updated'],
) => {
  store.dispatch(notificationSettingsUpdated(data))
}

const onAccountDeletionRequested = () => {
  store.dispatch(accountDeletionRequested())
}

const onAccountDeletionCancelled = () => {
  store.dispatch(accountDeletionCancelled())
}

export const registerWebSocketListeners = (ws: WebSocketManager) => {
  ws.on('notification_settings_updated', onNotificationSettingsUpdated)
  ws.on('account_deletion_requested', onAccountDeletionRequested)
  ws.on('account_deletion_cancelled', onAccountDeletionCancelled)
}

export const unregisterWebSocketListeners = (ws: WebSocketManager) => {
  ws.off('notification_settings_updated', onNotificationSettingsUpdated)
  ws.off('account_deletion_requested', onAccountDeletionRequested)
  ws.off('account_deletion_cancelled', onAccountDeletionCancelled)
}

