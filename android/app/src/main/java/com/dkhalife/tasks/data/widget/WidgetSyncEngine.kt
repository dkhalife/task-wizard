package com.dkhalife.tasks.data.widget

import android.content.Context
import androidx.datastore.preferences.core.intPreferencesKey
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.glance.appwidget.state.updateAppWidgetState
import androidx.glance.appwidget.updateAll
import com.dkhalife.tasks.data.sync.SyncEngine
import com.dkhalife.tasks.model.Task
import com.dkhalife.tasks.ui.widget.TaskGlanceWidget
import java.time.ZonedDateTime
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class WidgetSyncEngine @Inject constructor() : SyncEngine {

    override suspend fun sync(context: Context, tasks: List<Task>) {
        val nextTask = findNextUpcomingTask(tasks)

        val widgetManager = android.appwidget.AppWidgetManager.getInstance(context)
        val widgetComponent = android.content.ComponentName(context, com.dkhalife.tasks.ui.widget.TaskGlanceWidgetReceiver::class.java)
        val widgetIds = widgetManager.getAppWidgetIds(widgetComponent)
        if (widgetIds.isEmpty()) return

        val glanceIds = androidx.glance.appwidget.GlanceAppWidgetManager(context)
            .getGlanceIds(TaskGlanceWidget::class.java)

        for (glanceId in glanceIds) {
            updateAppWidgetState(context, glanceId) { prefs ->
                if (nextTask != null) {
                    prefs[KEY_TASK_ID] = nextTask.id
                    prefs[KEY_TASK_TITLE] = nextTask.title
                    prefs[KEY_TASK_DUE_DATE] = nextTask.nextDueDate ?: ""
                } else {
                    prefs.remove(KEY_TASK_ID)
                    prefs.remove(KEY_TASK_TITLE)
                    prefs.remove(KEY_TASK_DUE_DATE)
                }
            }
        }

        TaskGlanceWidget().updateAll(context)
    }

    private fun findNextUpcomingTask(tasks: List<Task>): Task? {
        val now = System.currentTimeMillis()
        var bestTask: Task? = null
        var bestMillis: Long? = null

        for (task in tasks) {
            val millis = parseToMillis(task.nextDueDate) ?: continue
            if (millis >= now) {
                if (bestMillis == null || millis < bestMillis) {
                    bestTask = task
                    bestMillis = millis
                }
            }
        }

        return bestTask
    }

    private fun parseToMillis(dateString: String?): Long? {
        if (dateString.isNullOrBlank()) return null
        return try {
            ZonedDateTime.parse(dateString).toInstant().toEpochMilli()
        } catch (_: Exception) {
            null
        }
    }

    companion object {
        val KEY_TASK_ID = intPreferencesKey("widget_task_id")
        val KEY_TASK_TITLE = stringPreferencesKey("widget_task_title")
        val KEY_TASK_DUE_DATE = stringPreferencesKey("widget_task_due_date")
    }
}
