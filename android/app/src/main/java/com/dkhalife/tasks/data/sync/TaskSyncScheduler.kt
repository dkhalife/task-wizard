package com.dkhalife.tasks.data.sync

import android.appwidget.AppWidgetManager
import android.content.ComponentName
import android.content.Context
import android.content.SharedPreferences
import androidx.work.Constraints
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.NetworkType
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import com.dkhalife.tasks.data.AppPreferences
import com.dkhalife.tasks.ui.widget.TaskGlanceWidgetReceiver
import java.util.concurrent.TimeUnit
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TaskSyncScheduler @Inject constructor(
    private val sharedPreferences: SharedPreferences
) {

    fun ensureScheduled(workManager: WorkManager) {
        val constraints = Constraints.Builder()
            .setRequiredNetworkType(NetworkType.CONNECTED)
            .build()

        val syncRequest = PeriodicWorkRequestBuilder<TaskSyncWorker>(
            SYNC_INTERVAL_MINUTES, TimeUnit.MINUTES
        )
            .setConstraints(constraints)
            .addTag(WORK_TAG)
            .build()

        workManager.enqueueUniquePeriodicWork(
            WORK_NAME,
            ExistingPeriodicWorkPolicy.UPDATE,
            syncRequest
        )
    }

    fun cancelIfUnneeded(workManager: WorkManager, context: Context) {
        val calendarSyncEnabled = sharedPreferences.getBoolean(AppPreferences.KEY_CALENDAR_SYNC, false)
        val hasWidgetInstances = hasActiveWidgets(context)

        if (!calendarSyncEnabled && !hasWidgetInstances) {
            workManager.cancelUniqueWork(WORK_NAME)
        }
    }

    private fun hasActiveWidgets(context: Context): Boolean {
        val widgetManager = AppWidgetManager.getInstance(context)
        val widgetComponent = ComponentName(context, TaskGlanceWidgetReceiver::class.java)
        return widgetManager.getAppWidgetIds(widgetComponent).isNotEmpty()
    }

    companion object {
        private const val SYNC_INTERVAL_MINUTES = 15L
        private const val WORK_NAME = "task_sync"
        private const val WORK_TAG = "task_sync"
    }
}
