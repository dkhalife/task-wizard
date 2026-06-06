package app.taskwiz.data.sync

import android.content.Context
import app.taskwiz.data.calendar.CalendarSyncEngine
import app.taskwiz.data.widget.WidgetSyncEngine
import app.taskwiz.repo.TaskRepository
import app.taskwiz.telemetry.TelemetryManager
import app.taskwiz.ws.WebSocketManager
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.FlowPreview
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.debounce
import kotlinx.coroutines.launch
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class WebSocketSyncBridge @Inject constructor(
    @ApplicationContext private val context: Context,
    private val webSocketManager: WebSocketManager,
    private val taskRepository: TaskRepository,
    private val syncCoordinator: SyncCoordinator,
    private val widgetSyncEngine: WidgetSyncEngine,
    private val calendarSyncEngine: CalendarSyncEngine,
    private val telemetryManager: TelemetryManager
) {
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    private var collectJob: Job? = null

    @OptIn(FlowPreview::class)
    fun start() {
        collectJob?.cancel()
        collectJob = scope.launch {
            webSocketManager.messages
                .debounce(DEBOUNCE_MS)
                .collect { message ->
                    if (message.action in SYNC_EVENTS) {
                        syncAll()
                    }
                }
        }
    }

    fun stop() {
        collectJob?.cancel()
        collectJob = null
    }

    private suspend fun syncAll() {
        try {
            syncCoordinator.syncOnceBlocking()
            val tasks = taskRepository.tasks.value
            try {
                widgetSyncEngine.sync(context, tasks)
            } catch (e: Exception) {
                telemetryManager.logError(TAG, "Widget sync failed: ${e.message}", e)
            }
            try {
                calendarSyncEngine.sync(context, tasks)
            } catch (e: Exception) {
                telemetryManager.logError(TAG, "Calendar sync failed: ${e.message}", e)
            }
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "WebSocket sync failed: ${e.message}", e)
        }
    }

    companion object {
        private const val TAG = "WebSocketSyncBridge"
        private const val DEBOUNCE_MS = 500L
        private val SYNC_EVENTS = setOf(
            "task_created",
            "task_updated",
            "task_deleted",
            "task_completed",
            "task_uncompleted",
            "task_skipped",
            "label_created",
            "label_updated",
            "label_deleted",
        )
    }
}

