package com.dkhalife.tasks.data.sync

import android.content.Context
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import com.dkhalife.tasks.repo.TaskRepository
import com.dkhalife.tasks.telemetry.TelemetryManager

class TaskSyncWorker(
    appContext: Context,
    workerParams: WorkerParameters,
    private val syncCoordinator: SyncCoordinator,
    private val taskRepository: TaskRepository,
    private val engines: List<SyncEngine>,
    private val telemetryManager: TelemetryManager
) : CoroutineWorker(appContext, workerParams) {

    override suspend fun doWork(): Result {
        return try {
            syncCoordinator.syncOnceBlocking()
            val tasks = taskRepository.tasks.value
            var anyFailed = false
            for (engine in engines) {
                try {
                    engine.sync(applicationContext, tasks)
                } catch (e: Exception) {
                    telemetryManager.logError(TAG, "Sync engine ${engine::class.simpleName} failed: ${e.message}", e)
                    anyFailed = true
                }
            }
            if (anyFailed) Result.retry() else Result.success()
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Task sync failed: ${e.message}", e)
            Result.retry()
        }
    }

    companion object {
        private const val TAG = "TaskSyncWorker"
    }
}

