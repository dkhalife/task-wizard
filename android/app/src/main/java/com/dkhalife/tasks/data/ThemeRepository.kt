package com.dkhalife.tasks.data

import android.content.SharedPreferences
import androidx.core.content.edit
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ThemeRepository @Inject constructor(
    private val sharedPreferences: SharedPreferences
) {
    fun getThemeMode(): ThemeMode {
        val name = sharedPreferences.getString(AppPreferences.KEY_THEME_MODE, null)
        return ThemeMode.entries.find { it.name == name } ?: ThemeMode.SYSTEM
    }

    fun setThemeMode(mode: ThemeMode) {
        sharedPreferences.edit { putString(AppPreferences.KEY_THEME_MODE, mode.name) }
    }
}
