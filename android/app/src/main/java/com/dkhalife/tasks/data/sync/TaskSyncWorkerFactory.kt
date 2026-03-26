package com.dkhalife.tasks.data.sync

import android.content.Context
import androidx.work.ListenableWorker
import androidx.work.WorkerFactory
import androidx.work.WorkerParameters
import com.dkhalife.tasks.api.TaskWizardApi
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TaskSyncWorkerFactory @Inject constructor(
    private val api: TaskWizardApi,
    private val engines: List<@JvmSuppressWildcards SyncEngine>
) : WorkerFactory() {

    override fun createWorker(
        appContext: Context,
        workerClassName: String,
        workerParameters: WorkerParameters
    ): ListenableWorker? {
        if (workerClassName == TaskSyncWorker::class.java.name) {
            return TaskSyncWorker(appContext, workerParameters, api, engines)
        }
        return null
    }
}
