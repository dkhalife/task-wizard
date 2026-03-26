package com.dkhalife.tasks.data.calendar

import android.content.ContentResolver
import android.content.SharedPreferences
import android.graphics.Color
import androidx.core.content.edit
import androidx.work.Constraints
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.NetworkType
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import com.dkhalife.tasks.data.AppPreferences
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
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

    suspend fun enableCalendarSync(contentResolver: ContentResolver, workManager: WorkManager): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            val existingId = calendarProviderClient.getCalendarId(contentResolver, ACCOUNT_NAME)
            if (existingId == null) {
                calendarProviderClient.createCalendar(
                    contentResolver, ACCOUNT_NAME, CALENDAR_DISPLAY_NAME, CALENDAR_COLOR
                )
            }

            val constraints = Constraints.Builder()
                .setRequiredNetworkType(NetworkType.CONNECTED)
                .build()

            val syncRequest = PeriodicWorkRequestBuilder<CalendarSyncWorker>(
                SYNC_INTERVAL_MINUTES, TimeUnit.MINUTES
            )
                .setConstraints(constraints)
                .addTag(WORK_TAG)
                .build()

            workManager.enqueueUniquePeriodicWork(
                WORK_NAME,
                ExistingPeriodicWorkPolicy.UPDATE,
                syncRequest
            )

            sharedPreferences.edit { putBoolean(AppPreferences.KEY_CALENDAR_SYNC, true) }
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun disableCalendarSync(contentResolver: ContentResolver, workManager: WorkManager): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            workManager.cancelUniqueWork(WORK_NAME)

            val calendarId = calendarProviderClient.getCalendarId(contentResolver, ACCOUNT_NAME)
            if (calendarId != null) {
                calendarProviderClient.deleteCalendar(contentResolver, calendarId, ACCOUNT_NAME)
            }

            sharedPreferences.edit { putBoolean(AppPreferences.KEY_CALENDAR_SYNC, false) }
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    companion object {
        const val ACCOUNT_NAME = "com.dkhalife.tasks.calendar"
        private const val CALENDAR_DISPLAY_NAME = "Task Wizard"
        private val CALENDAR_COLOR = Color.parseColor("#4A90D9")
        private const val SYNC_INTERVAL_MINUTES = 15L
        private const val WORK_NAME = "calendar_sync"
        private const val WORK_TAG = "calendar_sync"
    }
}
