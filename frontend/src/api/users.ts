import { Request } from '../utils/api'
import { User } from '@/models/user'
import {
  NotificationTriggerOptions,
  NotificationType,
} from '@/models/notifications'
import { transport } from './transport'

type UserResponse = {
  user: User
}

export const GetUserProfile = async () =>
  await Request<UserResponse>(`/users/profile`)

export const UpdateNotificationSettings = async (
  provider: NotificationType,
  triggers: NotificationTriggerOptions,
) =>
  await transport({
    http: () =>
      Request<void>(`/users/notifications`, 'PUT', {
        provider,
        triggers,
      }),
    ws: (ws) => ws.request('update_notification_settings', { provider, triggers }),
  })

export const RequestAccountDeletion = async () =>
  await Request<void>(`/users/deletion`, 'POST')

export const CancelAccountDeletion = async () =>
  await Request<void>(`/users/deletion`, 'DELETE')
