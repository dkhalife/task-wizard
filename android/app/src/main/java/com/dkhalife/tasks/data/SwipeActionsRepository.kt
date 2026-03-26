package com.dkhalife.tasks.data

import android.content.SharedPreferences
import androidx.core.content.edit
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class SwipeActionsRepository @Inject constructor(
    private val sharedPreferences: SharedPreferences
) {
    fun getSettings(): SwipeSettings = SwipeSettings(
        enabled = sharedPreferences.getBoolean(AppPreferences.KEY_SWIPE_ENABLED, true),
        startToEndAction = getAction(AppPreferences.KEY_SWIPE_START_TO_END_ACTION, SwipeAction.COMPLETE),
        endToStartAction = getAction(AppPreferences.KEY_SWIPE_END_TO_START_ACTION, SwipeAction.DELETE),
        deleteConfirmationEnabled = sharedPreferences.getBoolean(AppPreferences.KEY_SWIPE_DELETE_CONFIRMATION, false)
    )

    fun setEnabled(enabled: Boolean) {
        sharedPreferences.edit { putBoolean(AppPreferences.KEY_SWIPE_ENABLED, enabled) }
    }

    fun setStartToEndAction(action: SwipeAction) {
        sharedPreferences.edit { putString(AppPreferences.KEY_SWIPE_START_TO_END_ACTION, action.name) }
    }

    fun setEndToStartAction(action: SwipeAction) {
        sharedPreferences.edit { putString(AppPreferences.KEY_SWIPE_END_TO_START_ACTION, action.name) }
    }

    fun setDeleteConfirmationEnabled(enabled: Boolean) {
        sharedPreferences.edit { putBoolean(AppPreferences.KEY_SWIPE_DELETE_CONFIRMATION, enabled) }
    }

    private fun getAction(key: String, default: SwipeAction): SwipeAction {
        val name = sharedPreferences.getString(key, null)
        return SwipeAction.entries.find { it.name == name } ?: default
    }
}
