package com.dkhalife.tasks.ui.widget.summary

import android.content.Context
import androidx.glance.appwidget.GlanceAppWidget
import androidx.glance.appwidget.GlanceAppWidgetReceiver
import androidx.work.WorkManager
import com.dkhalife.tasks.data.sync.TaskSyncScheduler
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class TaskSummaryWidgetReceiver : GlanceAppWidgetReceiver() {

    override val glanceAppWidget: GlanceAppWidget = TaskSummaryWidget()

    @Inject
    lateinit var taskSyncScheduler: TaskSyncScheduler

    override fun onEnabled(context: Context) {
        super.onEnabled(context)
        taskSyncScheduler.ensureScheduled(WorkManager.getInstance(context))
    }

    override fun onDisabled(context: Context) {
        super.onDisabled(context)
        taskSyncScheduler.cancelIfUnneeded(WorkManager.getInstance(context), context)
    }
}
