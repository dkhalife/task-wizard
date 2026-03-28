package com.dkhalife.tasks.data.sync

import android.content.Context
import com.dkhalife.tasks.api.TaskWizardApi
import com.dkhalife.tasks.data.calendar.CalendarSyncEngine
import com.dkhalife.tasks.data.widget.WidgetSyncEngine
import com.dkhalife.tasks.telemetry.TelemetryManager
import com.dkhalife.tasks.ws.WebSocketManager
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
    private val api: TaskWizardApi,
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
                    if (message.action in TASK_EVENTS) {
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
            val response = api.getTasks()
            if (!response.isSuccessful) {
                telemetryManager.logWarning(TAG, "Failed to fetch tasks for sync: ${response.code()}")
                return
            }
            val tasks = response.body()?.tasks ?: emptyList()
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
        private val TASK_EVENTS = setOf(
            "task_created",
            "task_updated",
            "task_deleted",
            "task_completed",
            "task_uncompleted",
            "task_skipped"
        )
    }
}
