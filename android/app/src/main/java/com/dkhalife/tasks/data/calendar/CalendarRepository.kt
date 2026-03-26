package com.dkhalife.tasks.data.calendar

import android.content.ContentResolver
import android.content.Context
import android.content.SharedPreferences
import android.graphics.Color
import androidx.core.content.edit
import androidx.work.WorkManager
import com.dkhalife.tasks.data.AppPreferences
import com.dkhalife.tasks.data.sync.TaskSyncScheduler
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class CalendarRepository @Inject constructor(
    @ApplicationContext private val appContext: Context,
    private val sharedPreferences: SharedPreferences,
    private val calendarProviderClient: CalendarProviderClient,
    private val taskSyncScheduler: TaskSyncScheduler
) {

    fun isCalendarSyncEnabled(): Boolean {
        return sharedPreferences.getBoolean(AppPreferences.KEY_CALENDAR_SYNC, false)
    }

    suspend fun enableCalendarSync(contentResolver: ContentResolver, workManager: WorkManager): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            val existingId = calendarProviderClient.getCalendarId(contentResolver, ACCOUNT_NAME)
            if (existingId == null) {
                calendarProviderClient.createCalendar(
                    contentResolver, ACCOUNT_NAME, CALENDAR_DISPLAY_NAME, CALENDAR_COLOR
                )
            }

            taskSyncScheduler.ensureScheduled(workManager)

            sharedPreferences.edit { putBoolean(AppPreferences.KEY_CALENDAR_SYNC, true) }
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun disableCalendarSync(contentResolver: ContentResolver, workManager: WorkManager): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            val calendarId = calendarProviderClient.getCalendarId(contentResolver, ACCOUNT_NAME)
            if (calendarId != null) {
                calendarProviderClient.deleteCalendar(contentResolver, calendarId, ACCOUNT_NAME)
            }

            sharedPreferences.edit { putBoolean(AppPreferences.KEY_CALENDAR_SYNC, false) }
            taskSyncScheduler.cancelIfUnneeded(workManager, appContext)
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    companion object {
        const val ACCOUNT_NAME = "com.dkhalife.tasks.calendar"
        private const val CALENDAR_DISPLAY_NAME = "Task Wizard"
        private val CALENDAR_COLOR = Color.parseColor("#4A90D9")
    }
}
