package com.dkhalife.tasks.data.sync

import android.content.Context
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import com.dkhalife.tasks.api.TaskWizardApi

class TaskSyncWorker(
    appContext: Context,
    workerParams: WorkerParameters,
    private val api: TaskWizardApi,
    private val engines: List<SyncEngine>
) : CoroutineWorker(appContext, workerParams) {

    override suspend fun doWork(): Result {
        return try {
            val response = api.getTasks()
            if (response.isSuccessful) {
                val tasks = response.body()?.tasks ?: emptyList()
                var anyFailed = false
                for (engine in engines) {
                    try {
                        engine.sync(applicationContext, tasks)
                    } catch (_: Exception) {
                        anyFailed = true
                    }
                }
                if (anyFailed) Result.retry() else Result.success()
            } else {
                Result.retry()
            }
        } catch (_: Exception) {
            Result.retry()
        }
    }
}
