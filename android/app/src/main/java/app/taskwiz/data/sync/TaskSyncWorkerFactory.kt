package app.taskwiz.data.sync

import android.content.Context
import androidx.work.ListenableWorker
import androidx.work.WorkerFactory
import androidx.work.WorkerParameters
import app.taskwiz.repo.TaskRepository
import app.taskwiz.telemetry.TelemetryManager
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TaskSyncWorkerFactory @Inject constructor(
    private val syncCoordinator: SyncCoordinator,
    private val taskRepository: TaskRepository,
    private val engines: List<@JvmSuppressWildcards SyncEngine>,
    private val telemetryManager: TelemetryManager
) : WorkerFactory() {

    override fun createWorker(
        appContext: Context,
        workerClassName: String,
        workerParameters: WorkerParameters
    ): ListenableWorker? {
        if (workerClassName == TaskSyncWorker::class.java.name) {
            return TaskSyncWorker(appContext, workerParameters, syncCoordinator, taskRepository, engines, telemetryManager)
        }
        return null
    }
}

