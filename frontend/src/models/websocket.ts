import { Label } from './label'
import { APIToken } from './token'
import { NotificationTrigger, NotificationTriggerOptions, NotificationType } from './notifications'
import { Task } from './task'

export type WSAction =
  | 'get_user_labels'
  | 'create_label'
  | 'update_label'
  | 'delete_label'

export interface WSActionPayloads {
  get_user_labels: void
  create_label: Omit<Label, 'id'>
  update_label: Label
  delete_label: { id: number }
}

export type WSEvent =
  | 'label_created'
  | 'label_updated'
  | 'label_deleted'
  | 'app_token_created'
  | 'app_token_deleted'
  | 'notification_settings_updated'
  | 'task_created'
  | 'task_updated'
  | 'task_deleted'
  | 'task_completed'
  | 'task_uncompleted'
  | 'task_skipped'
  | 'notification'

export interface WSEventPayloads {
  label_created: { label: Label }
  label_updated: { label: Label }
  label_deleted: { id: number }
  app_token_created: APIToken
  app_token_deleted: { id: string }
  notification_settings_updated: {
    provider: NotificationType
    triggers: NotificationTriggerOptions
  }
  task_created: Task
  task_updated: Task
  task_deleted: { id: number }
  task_completed: Task
  task_uncompleted: Task
  task_skipped: Task
  notification: { task_id: number, type: NotificationTrigger }
}

export interface WSRequest<T extends WSAction = WSAction> {
  requestId: string
  action: T
  data?: WSActionPayloads[T]
}

export interface WSResponse<T extends WSEvent = WSEvent> {
  action: T
  status: number
  requestId?: string
  data: WSEventPayloads[T]
}
