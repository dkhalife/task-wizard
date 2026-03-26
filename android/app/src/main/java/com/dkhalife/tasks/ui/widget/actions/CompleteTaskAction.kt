package com.dkhalife.tasks.ui.widget.actions

import android.content.Context
import androidx.glance.GlanceId
import androidx.glance.action.ActionParameters
import androidx.glance.appwidget.action.ActionCallback
import androidx.work.WorkManager
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

        try {
            val response = entryPoint.api().completeTask(taskId)
            if (response.isSuccessful) {
                entryPoint.taskSyncScheduler().ensureScheduled(WorkManager.getInstance(context))
            }
        } catch (_: Exception) {
            // Silently fail — widget will update on next sync
        }
    }

    companion object {
        val PARAM_TASK_ID = ActionParameters.Key<Int>("complete_task_id")
    }
}
