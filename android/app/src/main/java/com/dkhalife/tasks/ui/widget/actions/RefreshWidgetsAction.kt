package com.dkhalife.tasks.ui.widget.actions

import android.content.Context
import androidx.glance.GlanceId
import androidx.glance.action.ActionParameters
import androidx.glance.appwidget.action.ActionCallback
import androidx.work.WorkManager
import dagger.hilt.android.EntryPointAccessors

class RefreshWidgetsAction : ActionCallback {

    override suspend fun onAction(
        context: Context,
        glanceId: GlanceId,
        parameters: ActionParameters
    ) {
        val entryPoint = EntryPointAccessors.fromApplication(
            context.applicationContext,
            WidgetEntryPoint::class.java
        )
        entryPoint.taskSyncScheduler().ensureScheduled(WorkManager.getInstance(context))
        entryPoint.taskSyncScheduler().triggerImmediate(WorkManager.getInstance(context))
    }
}
