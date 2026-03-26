package com.dkhalife.tasks.data.calendar

import android.content.ContentResolver
import android.content.SharedPreferences
import android.graphics.Color
import androidx.core.content.edit
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import com.dkhalife.tasks.data.AppPreferences
import java.util.concurrent.TimeUnit
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class CalendarRepository @Inject constructor(
    private val sharedPreferences: SharedPreferences,
    private val calendarProviderClient: CalendarProviderClient
) {

    fun isCalendarSyncEnabled(): Boolean {
        return sharedPreferences.getBoolean(AppPreferences.KEY_CALENDAR_SYNC, false)
    }

    fun enableCalendarSync(contentResolver: ContentResolver, workManager: WorkManager) {
        val existingId = calendarProviderClient.getCalendarId(contentResolver, ACCOUNT_NAME)
        if (existingId == null) {
            calendarProviderClient.createCalendar(
                contentResolver, ACCOUNT_NAME, CALENDAR_DISPLAY_NAME, CALENDAR_COLOR
            )
        }

        val syncRequest = PeriodicWorkRequestBuilder<CalendarSyncWorker>(
            SYNC_INTERVAL_MINUTES, TimeUnit.MINUTES
        ).build()

        workManager.enqueueUniquePeriodicWork(
            WORK_TAG,
            ExistingPeriodicWorkPolicy.UPDATE,
            syncRequest
        )

        sharedPreferences.edit { putBoolean(AppPreferences.KEY_CALENDAR_SYNC, true) }
    }

    fun disableCalendarSync(contentResolver: ContentResolver, workManager: WorkManager) {
        workManager.cancelUniqueWork(WORK_TAG)

        val calendarId = calendarProviderClient.getCalendarId(contentResolver, ACCOUNT_NAME)
        if (calendarId != null) {
            calendarProviderClient.deleteCalendar(contentResolver, calendarId)
        }

        sharedPreferences.edit { putBoolean(AppPreferences.KEY_CALENDAR_SYNC, false) }
    }

    companion object {
        const val ACCOUNT_NAME = "com.dkhalife.tasks.calendar"
        private const val CALENDAR_DISPLAY_NAME = "Task Wizard"
        private val CALENDAR_COLOR = Color.parseColor("#4A90D9")
        private const val SYNC_INTERVAL_MINUTES = 15L
        private const val WORK_TAG = "calendar_sync"
    }
}
