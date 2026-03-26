package com.dkhalife.tasks.ui.widget.summary

import android.content.Context
import androidx.compose.runtime.Composable
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
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
import androidx.glance.appwidget.provideContent
import androidx.glance.currentState
import androidx.glance.layout.Alignment
import androidx.glance.layout.Column
import androidx.glance.layout.Row
import androidx.glance.layout.fillMaxWidth
import androidx.glance.layout.padding
import androidx.glance.state.GlanceStateDefinition
import androidx.glance.state.PreferencesGlanceStateDefinition
import androidx.glance.text.FontWeight
import androidx.glance.text.Text
import androidx.glance.text.TextStyle
import com.dkhalife.tasks.MainActivity
import com.dkhalife.tasks.R
import com.dkhalife.tasks.data.TaskGrouper
import com.dkhalife.tasks.data.widget.WidgetSyncEngine
import com.dkhalife.tasks.ui.widget.WidgetTheme
import com.google.gson.Gson

class TaskSummaryWidget : GlanceAppWidget() {

    override val sizeMode = SizeMode.Single

    override val stateDefinition: GlanceStateDefinition<*> = PreferencesGlanceStateDefinition

    override suspend fun provideGlance(context: Context, id: GlanceId) {
        provideContent {
            GlanceTheme(colors = WidgetTheme.colors) {
                SummaryContent()
            }
        }
    }

    @Composable
    private fun SummaryContent() {
        val context = LocalContext.current
        val prefs = currentState<Preferences>()
        val tasks = WidgetSyncEngine.deserializeTasks(Gson(), prefs[WidgetSyncEngine.KEY_TASKS_JSON])
        val groups = TaskGrouper.groupByDueDate(context, tasks)

        val openAppAction = actionStartActivity<MainActivity>()

        Scaffold(
            titleBar = {
                TitleBar(
                    startIcon = ImageProvider(R.mipmap.ic_launcher),
                    title = context.getString(R.string.widget_summary_name),
                )
            },
            backgroundColor = GlanceTheme.colors.surface,
        ) {
            Row(
                modifier = GlanceModifier
                    .fillMaxWidth()
                    .padding(horizontal = 8.dp, vertical = 4.dp)
                    .clickable(openAppAction),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                for (group in groups) {
                    val color = WidgetTheme.groupColor(group.key)
                    Column(
                        modifier = GlanceModifier.padding(horizontal = 6.dp),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Text(
                            text = "${group.tasks.size}",
                            style = TextStyle(
                                color = color,
                                fontSize = 20.sp,
                                fontWeight = FontWeight.Bold
                            )
                        )
                        Text(
                            text = group.name,
                            style = TextStyle(
                                color = GlanceTheme.colors.onSurfaceVariant,
                                fontSize = 10.sp
                            ),
                            maxLines = 1
                        )
                    }
                }
            }
        }
    }
}
