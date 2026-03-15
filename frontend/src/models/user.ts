import {
  NotificationTriggerOptions,
  NotificationType,
} from '@/models/notifications'

export interface User {
  notifications: {
    provider: NotificationType
    triggers: NotificationTriggerOptions
  }
}
