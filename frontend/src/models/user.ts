import {
  NotificationTriggerOptions,
  NotificationType,
} from '@/models/notifications'

export interface User {
  notifications: {
    provider: NotificationType
    triggers: NotificationTriggerOptions
  }
  deletion_requested_at: string | null
}
