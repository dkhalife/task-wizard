package com.dkhalife.tasks.data

object AppPreferences {
    const val PREFS_NAME = "task_wizard_prefs"
    const val KEY_THEME_MODE = "theme_mode"
}

enum class ThemeMode {
    LIGHT, DARK, SYSTEM
}
