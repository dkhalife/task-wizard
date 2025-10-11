import { Request } from '../utils/api'
import { User } from '@/models/user'
import {
  NotificationTriggerOptions,
  NotificationType,
} from '@/models/notifications'

type UserResponse = {
  user: User
}

export const UpdatePassword = async (newPassword: string) =>
  await Request<void>(`/users/change_password`, 'PUT', {
    password: newPassword,
  })

export const GetUserProfile = async () =>
  await Request<UserResponse>(`/users/profile`)

export const UpdateNotificationSettings = async (
  provider: NotificationType,
  triggers: NotificationTriggerOptions,
) =>
  await Request<void>(`/users/notifications`, 'PUT', {
    provider,
    triggers,
  })
