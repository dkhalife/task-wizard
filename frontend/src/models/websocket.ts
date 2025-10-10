import { Label } from './label'
import { APIToken } from './token'
import { NotificationTrigger, NotificationTriggerOptions, NotificationType } from './notifications'
import { Task } from './task'

export type WSAction =
  | 'get_user_labels'
  | 'create_label'
  | 'update_label'
  | 'delete_label'
  | 'get_tasks'
  | 'get_completed_tasks'
  | 'get_task'
  | 'create_task'
  | 'update_task'
  | 'delete_task'
  | 'skip_task'
  | 'update_due_date'
  | 'complete_task'
  | 'uncomplete_task'
  | 'get_task_history'
  | 'get_app_tokens'
  | 'create_app_token'
  | 'delete_app_token'
  | 'update_notification_settings'
  | 'get_user_profile'
  | 'update_password'

export interface WSActionPayloads {
  get_user_labels: void
  create_label: Omit<Label, 'id'>
  update_label: Label
  delete_label: { id: number }
  get_tasks: void
  get_completed_tasks: { limit?: number; page?: number }
  get_task: { id: number }
  create_task: Omit<Task, 'id'>
  update_task: Task
  delete_task: { id: number }
  skip_task: { id: number }
  update_due_date: { id: number; due_date: string }
  complete_task: { id: number }
  uncomplete_task: { id: number }
  get_task_history: { id: number }
  get_app_tokens: void
  create_app_token: { name: string; scopes: string[]; expiration: number }
  delete_app_token: { id: string }
  update_notification_settings: { provider: NotificationType; triggers: NotificationTriggerOptions }
  get_user_profile: void
  update_password: { password: string }
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
