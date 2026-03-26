package com.dkhalife.tasks.ui.widget

import android.content.Context
import androidx.compose.runtime.Composable
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.glance.GlanceId
import androidx.glance.GlanceModifier
import androidx.glance.GlanceTheme
import androidx.glance.LocalContext
import androidx.glance.action.ActionParameters
import androidx.glance.action.actionParametersOf
import androidx.glance.action.actionStartActivity
import androidx.glance.action.clickable
import androidx.glance.appwidget.GlanceAppWidget
import androidx.glance.appwidget.SizeMode
import androidx.glance.appwidget.provideContent
import androidx.glance.background
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
import com.dkhalife.tasks.data.widget.WidgetSyncEngine
import java.time.ZonedDateTime
import java.time.format.DateTimeFormatter
import java.time.format.FormatStyle

class TaskGlanceWidget : GlanceAppWidget() {

    override val sizeMode = SizeMode.Single

    override val stateDefinition: GlanceStateDefinition<*> = PreferencesGlanceStateDefinition

    override suspend fun provideGlance(context: Context, id: GlanceId) {
        provideContent {
            GlanceTheme {
                WidgetContent()
            }
        }
    }

    @Composable
    private fun WidgetContent() {
        val context = LocalContext.current
        val prefs = currentState<androidx.datastore.preferences.core.Preferences>()
        val taskId = prefs[WidgetSyncEngine.KEY_TASK_ID]
        val taskTitle = prefs[WidgetSyncEngine.KEY_TASK_TITLE]
        val taskDueDate = prefs[WidgetSyncEngine.KEY_TASK_DUE_DATE]

        val clickAction = if (taskId != null) {
            actionStartActivity<MainActivity>(actionParametersOf(PARAM_TASK_ID to taskId))
        } else {
            actionStartActivity<MainActivity>()
        }

        Row(
            modifier = GlanceModifier
                .fillMaxWidth()
                .background(GlanceTheme.colors.surface)
                .padding(12.dp)
                .clickable(clickAction),
            verticalAlignment = Alignment.CenterVertically
        ) {
            if (taskTitle != null) {
                Column(modifier = GlanceModifier.defaultWeight()) {
                    Text(
                        text = taskTitle,
                        style = TextStyle(
                            color = GlanceTheme.colors.onSurface,
                            fontSize = 14.sp,
                            fontWeight = FontWeight.Medium
                        ),
                        maxLines = 1
                    )
                    if (!taskDueDate.isNullOrBlank()) {
                        Text(
                            text = formatDueDate(taskDueDate),
                            style = TextStyle(
                                color = GlanceTheme.colors.onSurfaceVariant,
                                fontSize = 12.sp
                            ),
                            maxLines = 1
                        )
                    }
                }
            } else {
                Text(
                    text = context.getString(R.string.widget_no_upcoming_tasks),
                    style = TextStyle(
                        color = GlanceTheme.colors.onSurfaceVariant,
                        fontSize = 14.sp
                    )
                )
            }
        }
    }

    companion object {
        const val EXTRA_TASK_ID = "taskId"
        val PARAM_TASK_ID = ActionParameters.Key<Int>(EXTRA_TASK_ID)

        private val displayFormatter = DateTimeFormatter.ofLocalizedDateTime(FormatStyle.MEDIUM, FormatStyle.SHORT)

        fun formatDueDate(dateString: String): String {
            return try {
                val zdt = ZonedDateTime.parse(dateString)
                zdt.format(displayFormatter)
            } catch (_: Exception) {
                dateString
            }
        }
    }
}
