package com.dkhalife.tasks.ui.widget.actions

import android.content.Context
import androidx.glance.GlanceId
import androidx.glance.action.ActionParameters
import androidx.glance.appwidget.action.ActionCallback
import androidx.glance.appwidget.state.updateAppWidgetState
import androidx.work.WorkManager
import com.dkhalife.tasks.data.widget.WidgetSyncEngine
import dagger.hilt.android.EntryPointAccessors

class CompleteTaskAction : ActionCallback {

    override suspend fun onAction(
        context: Context,
        glanceId: GlanceId,
        parameters: ActionParameters
    ) {
        val taskId = parameters[PARAM_TASK_ID] ?: return

        val entryPoint = EntryPointAccessors.fromApplication(
            context.applicationContext,
            WidgetEntryPoint::class.java
        )

        val widgetSyncEngine = entryPoint.widgetSyncEngine()
        val gson = entryPoint.gson()
        val telemetryManager = entryPoint.telemetryManager()

        var originalJson: String? = null
        updateAppWidgetState(context, glanceId) { prefs ->
            originalJson = prefs[WidgetSyncEngine.KEY_TASKS_JSON]
        }

        val originalTasks = WidgetSyncEngine.deserializeTasks(gson, originalJson)
        if (originalTasks.isNotEmpty()) {
            val optimisticTasks = originalTasks.filter { it.id != taskId }
            try {
                widgetSyncEngine.sync(context, optimisticTasks)
            } catch (_: Exception) {
                // Optimistic update is best-effort; continue with API call regardless
            }
        }

        val workManager = WorkManager.getInstance(context)
        try {
            val result = entryPoint.taskRepository().completeTask(taskId)
            if (result.isSuccess) {
                entryPoint.taskSyncScheduler().ensureScheduled(workManager)
                entryPoint.taskSyncScheduler().triggerImmediate(workManager)
            } else if (originalTasks.isNotEmpty()) {
                widgetSyncEngine.sync(context, originalTasks)
            }
        } catch (e: Exception) {
            if (originalTasks.isNotEmpty()) {
                try {
                    widgetSyncEngine.sync(context, originalTasks)
                } catch (revertEx: Exception) {
                    telemetryManager.logWarning(TAG, "Failed to revert optimistic widget update after API error: ${revertEx.message}", revertEx)
                }
            }
        }
    }

    companion object {
        private const val TAG = "CompleteTaskAction"
        val PARAM_TASK_ID = ActionParameters.Key<Int>("complete_task_id")
    }
}
