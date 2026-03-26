package com.dkhalife.tasks.ui.widget.labelfilter

import android.content.Context
import androidx.compose.runtime.Composable
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.intPreferencesKey
import androidx.datastore.preferences.core.stringPreferencesKey
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
import com.dkhalife.tasks.data.widget.WidgetSyncEngine
import com.dkhalife.tasks.ui.widget.WidgetTheme
import com.dkhalife.tasks.ui.widget.components.WidgetEmptyState
import com.dkhalife.tasks.ui.widget.components.WidgetTaskRow
import com.google.gson.Gson

class LabelFilterWidget : GlanceAppWidget() {

    override val sizeMode = SizeMode.Single

    override val stateDefinition: GlanceStateDefinition<*> = PreferencesGlanceStateDefinition

    override suspend fun provideGlance(context: Context, id: GlanceId) {
        provideContent {
            GlanceTheme(colors = WidgetTheme.colors) {
                LabelFilterContent()
            }
        }
    }

    @Composable
    private fun LabelFilterContent() {
        val context = LocalContext.current
        val prefs = currentState<Preferences>()
        val labelId = prefs[KEY_LABEL_ID]
        val labelName = prefs[KEY_LABEL_NAME] ?: context.getString(R.string.widget_label_filter_name)

        val allTasks = WidgetSyncEngine.deserializeTasks(Gson(), prefs[WidgetSyncEngine.KEY_TASKS_JSON])
        val filteredTasks = if (labelId != null) {
            allTasks.filter { task -> task.labels.any { it.id == labelId } }
        } else {
            emptyList()
        }

        val openAppAction = actionStartActivity<MainActivity>()

        Scaffold(
            titleBar = {
                TitleBar(
                    startIcon = ImageProvider(R.mipmap.ic_launcher),
                    title = labelName,
                )
            },
            backgroundColor = GlanceTheme.colors.surface,
        ) {
            if (labelId == null) {
                WidgetEmptyState(
                    message = context.getString(R.string.widget_label_filter_not_configured),
                    modifier = GlanceModifier.clickable(openAppAction)
                )
            } else if (filteredTasks.isEmpty()) {
                WidgetEmptyState(
                    message = context.getString(R.string.widget_label_filter_empty),
                    modifier = GlanceModifier.clickable(openAppAction)
                )
            } else {
                LazyColumn {
                    items(filteredTasks, itemId = { it.id.toLong() }) { task ->
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

    companion object {
        val KEY_LABEL_ID = intPreferencesKey("label_filter_label_id")
        val KEY_LABEL_NAME = stringPreferencesKey("label_filter_label_name")
    }
}
