import {
  NotificationTriggerOptions,
  NotificationType,
} from '@/models/notifications'

export interface User {
  display_name: string
  notifications: {
    provider: NotificationType
    triggers: NotificationTriggerOptions
  }
}

export const validateEmail = (email: string) => {
  return /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i.test(email)
}

export const validatePassword = (password: string) => {
  return password.length >= 8
}
