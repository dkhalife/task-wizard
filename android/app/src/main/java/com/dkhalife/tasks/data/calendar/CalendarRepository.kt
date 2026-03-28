package com.dkhalife.tasks.data.calendar

import android.accounts.Account
import android.accounts.AccountManager
import android.content.ContentResolver
import android.content.Context
import android.content.SharedPreferences
import android.graphics.Color
import androidx.core.content.edit
import androidx.work.WorkManager
import com.dkhalife.tasks.auth.AuthManager
import com.dkhalife.tasks.data.AppPreferences
import com.dkhalife.tasks.data.sync.TaskSyncScheduler
import com.dkhalife.tasks.telemetry.TelemetryManager
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
    private val taskSyncScheduler: TaskSyncScheduler,
    private val authManager: AuthManager,
    private val telemetryManager: TelemetryManager
) {

    fun isCalendarSyncEnabled(): Boolean {
        return sharedPreferences.getBoolean(AppPreferences.KEY_CALENDAR_SYNC, false)
    }

    fun getAccountName(): String {
        return authManager.getAccountName() ?: FALLBACK_ACCOUNT_NAME
    }

    private fun ensureAccount(context: Context) {
        val accountName = getAccountName()
        val accountManager = AccountManager.get(context)
        val account = Account(accountName, CalendarProviderClient.ACCOUNT_TYPE)
        val existing = accountManager.getAccountsByType(CalendarProviderClient.ACCOUNT_TYPE)
        if (existing.none { it.name == accountName }) {
            if (!accountManager.addAccountExplicitly(account, null, null)) {
                throw IllegalStateException("Failed to register calendar account")
            }
        }
    }

    suspend fun enableCalendarSync(context: Context, contentResolver: ContentResolver, workManager: WorkManager): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            ensureAccount(context)

            val accountName = getAccountName()
            val existingId = calendarProviderClient.getCalendarId(contentResolver, accountName)
            if (existingId == null) {
                calendarProviderClient.createCalendar(
                    contentResolver, accountName, CALENDAR_DISPLAY_NAME, CALENDAR_COLOR
                )
            }

            taskSyncScheduler.ensureScheduled(workManager)
            taskSyncScheduler.triggerImmediate(workManager)

            sharedPreferences.edit { putBoolean(AppPreferences.KEY_CALENDAR_SYNC, true) }
            Result.success(Unit)
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to enable calendar sync: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun disableCalendarSync(context: Context, contentResolver: ContentResolver, workManager: WorkManager): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            val accountName = getAccountName()
            val calendarId = calendarProviderClient.getCalendarId(contentResolver, accountName)
            if (calendarId != null) {
                calendarProviderClient.deleteCalendar(contentResolver, calendarId, accountName)
            }

            val accountManager = AccountManager.get(context)
            val account = Account(accountName, CalendarProviderClient.ACCOUNT_TYPE)
            if (!accountManager.removeAccountExplicitly(account)) {
                throw IllegalStateException("Failed to remove calendar account")
            }

            sharedPreferences.edit { putBoolean(AppPreferences.KEY_CALENDAR_SYNC, false) }
            taskSyncScheduler.cancelIfUnneeded(workManager, appContext)
            Result.success(Unit)
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to disable calendar sync: ${e.message}", e)
            Result.failure(e)
        }
    }

    companion object {
        private const val TAG = "CalendarRepository"
        private const val FALLBACK_ACCOUNT_NAME = "Task Wizard"
        internal const val CALENDAR_DISPLAY_NAME = "Task Wizard"
        internal val CALENDAR_COLOR = Color.parseColor("#4A90D9")
    }
}
