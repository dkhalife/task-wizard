package com.dkhalife.tasks.data

/**
 * Generates unique negative placeholder ids for records created offline, so they never collide
 * with server-assigned positive ids. Persisted in SharedPreferences so ids survive process death.
 */
class LocalIdGenerator(private val prefs: android.content.SharedPreferences) {
    fun nextId(): Int {
        synchronized(this) {
            val current = prefs.getInt(AppPreferences.KEY_LOCAL_ID_COUNTER, -1)
            val next = if (current >= 0) -1 else current - 1
            prefs.edit().putInt(AppPreferences.KEY_LOCAL_ID_COUNTER, next).apply()
            return next
        }
    }
}
