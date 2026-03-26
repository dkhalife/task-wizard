package com.dkhalife.tasks.ui.widget.duetoday

import android.content.Context
import androidx.compose.runtime.Composable
import androidx.datastore.preferences.core.Preferences
import androidx.glance.GlanceId
import androidx.glance.GlanceModifier
import androidx.glance.GlanceTheme
import androidx.glance.ImageProvider
import androidx.glance.LocalContext
import androidx.glance.action.actionStartActivity
import androidx.glance.action.clickable
import androidx.glance.appwidget.GlanceAppWidget
import androidx.glance.appwidget.SizeMode
import androidx.glance.appwidget.components.Scaffold
import androidx.glance.appwidget.components.TitleBar
import androidx.glance.appwidget.lazy.LazyColumn
import androidx.glance.appwidget.lazy.items
import androidx.glance.appwidget.provideContent
import androidx.glance.currentState
import androidx.glance.state.GlanceStateDefinition
import androidx.glance.state.PreferencesGlanceStateDefinition
import com.dkhalife.tasks.MainActivity
import com.dkhalife.tasks.R
import com.dkhalife.tasks.data.TaskGrouper
import com.dkhalife.tasks.data.widget.WidgetSyncEngine
import com.dkhalife.tasks.ui.widget.WidgetTheme
import com.dkhalife.tasks.ui.widget.components.WidgetEmptyState
import com.dkhalife.tasks.ui.widget.components.WidgetGroupHeader
import com.dkhalife.tasks.ui.widget.components.WidgetTaskRow
import com.google.gson.Gson

class DueTodayWidget : GlanceAppWidget() {

    override val sizeMode = SizeMode.Single

    override val stateDefinition: GlanceStateDefinition<*> = PreferencesGlanceStateDefinition

    override suspend fun provideGlance(context: Context, id: GlanceId) {
        provideContent {
            GlanceTheme(colors = WidgetTheme.colors) {
                DueTodayContent()
            }
        }
    }

    @Composable
    private fun DueTodayContent() {
        val context = LocalContext.current
        val prefs = currentState<Preferences>()
        val tasks = WidgetSyncEngine.deserializeTasks(Gson(), prefs[WidgetSyncEngine.KEY_TASKS_JSON])
        val groups = TaskGrouper.groupByDueDate(tasks)
            .filter { it.key == "overdue" || it.key == "today" }

        val openAppAction = actionStartActivity<MainActivity>()

        Scaffold(
            titleBar = {
                TitleBar(
                    startIcon = ImageProvider(R.mipmap.ic_launcher),
                    title = context.getString(R.string.widget_due_today_name),
                )
            },
            backgroundColor = GlanceTheme.colors.surface,
        ) {
            val allTasks = groups.flatMap { it.tasks }
            if (allTasks.isEmpty()) {
                WidgetEmptyState(
                    message = context.getString(R.string.widget_due_today_empty),
                    modifier = GlanceModifier.clickable(openAppAction)
                )
            } else {
                LazyColumn {
                    for (group in groups) {
                        item {
                            WidgetGroupHeader(
                                name = group.name,
                                count = group.tasks.size,
                                groupKey = group.key
                            )
                        }
                        items(group.tasks, itemId = { it.id.toLong() }) { task ->
                            WidgetTaskRow(
                                taskId = task.id,
                                title = task.title,
                                dueDate = task.nextDueDate,
                                labels = task.labels,
                            )
                        }
                    }
                }
            }
        }
    }
}
