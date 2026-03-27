package com.dkhalife.tasks.data.widget

import android.content.Context
import android.util.Log
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.glance.appwidget.GlanceAppWidgetManager
import androidx.glance.appwidget.state.updateAppWidgetState
import androidx.glance.appwidget.updateAll
import com.dkhalife.tasks.data.sync.SyncEngine
import com.dkhalife.tasks.model.Task
import com.dkhalife.tasks.telemetry.TelemetryManager
import com.dkhalife.tasks.ui.widget.TaskListWidget
import com.dkhalife.tasks.ui.widget.duetoday.DueTodayWidget
import com.dkhalife.tasks.ui.widget.labelfilter.LabelFilterWidget
import com.dkhalife.tasks.ui.widget.summary.TaskSummaryWidget
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class WidgetSyncEngine @Inject constructor(
    private val gson: Gson,
    private val telemetryManager: TelemetryManager
) : SyncEngine {

    override suspend fun sync(context: Context, tasks: List<Task>) {
        val json = gson.toJson(tasks)

        val widgetTypes = listOf(
            TaskListWidget::class.java,
            TaskSummaryWidget::class.java,
            DueTodayWidget::class.java,
            LabelFilterWidget::class.java,
        )

        for (widgetType in widgetTypes) {
            val glanceIds = try {
                GlanceAppWidgetManager(context).getGlanceIds(widgetType)
            } catch (e: Exception) {
                telemetryManager.logWarning(TAG, "Failed to get Glance IDs for ${widgetType.simpleName}: ${e.message}", e)
                continue
            }

            for (glanceId in glanceIds) {
                updateAppWidgetState(context, glanceId) { prefs ->
                    prefs[KEY_TASKS_JSON] = json
                }
            }
        }

        TaskListWidget().updateAll(context)
        TaskSummaryWidget().updateAll(context)
        DueTodayWidget().updateAll(context)
        LabelFilterWidget().updateAll(context)
    }

    companion object {
        private const val TAG = "WidgetSyncEngine"
        val KEY_TASKS_JSON = stringPreferencesKey("widget_tasks_json")

        fun deserializeTasks(gson: Gson, json: String?, telemetryManager: TelemetryManager? = null): List<Task> {
            if (json.isNullOrBlank()) return emptyList()
            return try {
                val type = object : TypeToken<List<Task>>() {}.type
                gson.fromJson(json, type)
            } catch (e: Exception) {
                telemetryManager?.logWarning(TAG, "Failed to deserialize tasks: ${e.message}", e)
                    ?: Log.w(TAG, "Failed to deserialize tasks: ${e.message}", e)
                emptyList()
            }
        }
    }
}
