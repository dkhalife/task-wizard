package com.dkhalife.tasks.data

import android.content.SharedPreferences
import androidx.core.content.edit
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TaskListSettingsRepository @Inject constructor(
    private val sharedPreferences: SharedPreferences
) {
    fun isInlineCompleteEnabled(): Boolean {
        return sharedPreferences.getBoolean(AppPreferences.KEY_INLINE_COMPLETE_ENABLED, true)
    }

    fun setInlineCompleteEnabled(enabled: Boolean) {
        sharedPreferences.edit { putBoolean(AppPreferences.KEY_INLINE_COMPLETE_ENABLED, enabled) }
    }
}
