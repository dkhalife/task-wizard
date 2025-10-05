export type NotificationDisabled = {
  enabled: false
}
export type NotificationMqtt = object
export type NotificationTrigger = 'due_date' | 'pre_due' | 'overdue'
export type NotificationTriggerOptions = Record<NotificationTrigger, boolean>
export type NotificationEnabled = {
  enabled: true
} & NotificationTriggerOptions &
  NotificationMqtt
export type Notification = NotificationDisabled | NotificationEnabled
export type NotificationProvider = 'none' | 'webhook' | 'gotify'
export type NotificationTypeNone = {
  provider: 'none'
}

export type WebhookMethod = 'GET' | 'POST'
export type NotificationTypeWebhook = {
  provider: 'webhook'
  url: string
  method: WebhookMethod
}

export type NotificationTypeGotify = {
  provider: 'gotify'
  url: string
  token: string
}

export type NotificationType =
  | NotificationTypeNone
  | NotificationTypeWebhook
  | NotificationTypeGotify

export const getDefaultTypeForProvider = (
  provider: NotificationProvider,
): NotificationType => {
  switch (provider) {
    default:
    case 'none':
      return {
        provider: 'none',
      }

    case 'webhook':
      return {
        provider: 'webhook',
        url: '',
        method: 'GET',
      }

    case 'gotify':
      return {
        provider: 'gotify',
        url: '',
        token: '',
      }
  }
}

export type NotificationTriggerOption = {
  type: NotificationTrigger
  title: string
  description: string
}

export const NotificationTriggerOptions: NotificationTriggerOption[] = [
  {
    title: 'Due Date/Time',
    description: 'After the due date and time has passed',
    type: 'due_date',
  },
  {
    title: 'Pre-due',
    description: 'A few hours before the due date',
    type: 'pre_due',
  },
  {
    title: 'Overdue',
    description: 'When left uncompleted at least one day past its due date',
    type: 'overdue',
  },
]
