package com.dkhalife.tasks.data

object AppPreferences {
    const val PREFS_NAME = "task_wizard_prefs"
    const val KEY_THEME_MODE = "theme_mode"
    const val KEY_TASK_GROUPING = "task_grouping"
    const val KEY_EXPANDED_GROUPS = "expanded_groups"
    const val KEY_CALENDAR_SYNC = "calendar_sync"
    const val KEY_SWIPE_ENABLED = "swipe_actions_enabled"
    const val KEY_SWIPE_START_TO_END_ACTION = "swipe_start_to_end_action"
    const val KEY_SWIPE_END_TO_START_ACTION = "swipe_end_to_start_action"
    const val KEY_SWIPE_DELETE_CONFIRMATION = "swipe_delete_confirmation"
    const val KEY_TELEMETRY_ENABLED = "telemetry_enabled"
    const val KEY_DEBUG_LOGGING_ENABLED = "debug_logging_enabled"
    const val KEY_DEVICE_IDENTIFIER = "device_identifier"
}

enum class ThemeMode {
    LIGHT, DARK, SYSTEM
}

enum class TaskGrouping {
    DUE_DATE, LABEL
}

enum class SwipeAction {
    COMPLETE, DELETE, SKIP, NONE
}

data class SwipeSettings(
    val enabled: Boolean = true,
    val startToEndAction: SwipeAction = SwipeAction.COMPLETE,
    val endToStartAction: SwipeAction = SwipeAction.DELETE,
    val deleteConfirmationEnabled: Boolean = false
)
