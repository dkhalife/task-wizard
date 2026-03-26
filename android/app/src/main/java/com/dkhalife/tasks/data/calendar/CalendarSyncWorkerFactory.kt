package com.dkhalife.tasks.data.calendar

import android.content.Context
import androidx.work.ListenableWorker
import androidx.work.WorkerFactory
import androidx.work.WorkerParameters
import com.dkhalife.tasks.api.TaskWizardApi
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class CalendarSyncWorkerFactory @Inject constructor(
    private val api: TaskWizardApi,
    private val calendarSyncEngine: CalendarSyncEngine,
    private val calendarProviderClient: CalendarProviderClient
) : WorkerFactory() {

    override fun createWorker(
        appContext: Context,
        workerClassName: String,
        workerParameters: WorkerParameters
    ): ListenableWorker? {
        if (workerClassName == CalendarSyncWorker::class.java.name) {
            return CalendarSyncWorker(
                appContext, workerParameters, api, calendarSyncEngine, calendarProviderClient
            )
        }
        return null
    }
}
