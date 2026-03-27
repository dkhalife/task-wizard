package com.dkhalife.tasks.data

import android.content.SharedPreferences
import androidx.core.content.edit
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TelemetryRepository @Inject constructor(
    private val sharedPreferences: SharedPreferences
) {
    fun isTelemetryEnabled(): Boolean {
        return sharedPreferences.getBoolean(AppPreferences.KEY_TELEMETRY_ENABLED, false)
    }

    fun setTelemetryEnabled(enabled: Boolean) {
        sharedPreferences.edit { putBoolean(AppPreferences.KEY_TELEMETRY_ENABLED, enabled) }
    }

    fun isDebugLoggingEnabled(): Boolean {
        return sharedPreferences.getBoolean(AppPreferences.KEY_DEBUG_LOGGING_ENABLED, false)
    }

    fun setDebugLoggingEnabled(enabled: Boolean) {
        sharedPreferences.edit { putBoolean(AppPreferences.KEY_DEBUG_LOGGING_ENABLED, enabled) }
    }
}
