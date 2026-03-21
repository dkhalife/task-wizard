package com.dkhalife.tasks.ws

object WebSocketActions {
    const val GET_TASKS = "get_tasks"
    const val GET_COMPLETED_TASKS = "get_completed_tasks"
    const val GET_TASK = "get_task"
    const val CREATE_TASK = "create_task"
    const val UPDATE_TASK = "update_task"
    const val DELETE_TASK = "delete_task"
    const val SKIP_TASK = "skip_task"
    const val UPDATE_DUE_DATE = "update_due_date"
    const val COMPLETE_TASK = "complete_task"
    const val UNCOMPLETE_TASK = "uncomplete_task"
    const val GET_TASK_HISTORY = "get_task_history"

    const val GET_USER_LABELS = "get_user_labels"
    const val CREATE_LABEL = "create_label"
    const val UPDATE_LABEL = "update_label"
    const val DELETE_LABEL = "delete_label"

    const val UPDATE_NOTIFICATION_SETTINGS = "update_notification_settings"

    const val TASK_UPDATED = "task_updated"
}
