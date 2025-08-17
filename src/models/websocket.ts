export type WSAction =
  | 'get_user_labels'
  | 'create_label'
  | 'update_label'
  | 'delete_label'

export interface WSRequest {
  requestId: string
  action: WSAction
  data?: unknown
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

export interface WSResponse {
  action: WSEvent
  status: number
  requestId?: string
  data?: any
}
