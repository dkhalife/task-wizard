package com.dkhalife.tasks.data.calendar

import android.content.Context
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import com.dkhalife.tasks.api.TaskWizardApi

class CalendarSyncWorker(
    appContext: Context,
    workerParams: WorkerParameters,
    private val api: TaskWizardApi,
    private val calendarSyncEngine: CalendarSyncEngine,
    private val calendarProviderClient: CalendarProviderClient
) : CoroutineWorker(appContext, workerParams) {

    override suspend fun doWork(): Result {
        val calendarId = calendarProviderClient.getCalendarId(
            applicationContext.contentResolver,
            CalendarRepository.ACCOUNT_NAME
        )

        if (calendarId == null) {
            return Result.retry()
        }

        return try {
            val response = api.getTasks()
            if (response.isSuccessful) {
                val tasks = response.body()?.tasks ?: emptyList()
                calendarSyncEngine.sync(applicationContext.contentResolver, calendarId, tasks)
                Result.success()
            } else {
                Result.retry()
            }
        } catch (_: Exception) {
            Result.retry()
        }
    }
}
