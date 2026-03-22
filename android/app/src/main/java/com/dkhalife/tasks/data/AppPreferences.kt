package com.dkhalife.tasks.data

object AppPreferences {
    const val PREFS_NAME = "task_wizard_prefs"
    const val KEY_THEME_MODE = "theme_mode"
    const val KEY_TASK_GROUPING = "task_grouping"
    const val KEY_EXPANDED_GROUPS = "expanded_groups"
}

enum class ThemeMode {
    LIGHT, DARK, SYSTEM
}

enum class TaskGrouping {
    DUE_DATE, LABEL
}
