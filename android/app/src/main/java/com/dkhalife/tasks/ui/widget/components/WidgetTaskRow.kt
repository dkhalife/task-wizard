package com.dkhalife.tasks.ui.widget.components

import androidx.compose.runtime.Composable
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.glance.GlanceModifier
import androidx.glance.GlanceTheme
import androidx.glance.action.ActionParameters
import androidx.glance.action.actionParametersOf
import androidx.glance.action.actionStartActivity
import androidx.glance.action.clickable
import androidx.glance.appwidget.action.actionRunCallback
import androidx.glance.layout.Alignment
import androidx.glance.layout.Column
import androidx.glance.layout.Row
import androidx.glance.layout.fillMaxWidth
import androidx.glance.layout.padding
import androidx.glance.layout.size
import androidx.glance.text.FontWeight
import androidx.glance.text.Text
import androidx.glance.text.TextStyle
import com.dkhalife.tasks.MainActivity
import com.dkhalife.tasks.model.Label
import com.dkhalife.tasks.ui.utils.formatDueDate
import com.dkhalife.tasks.ui.utils.parseDueDate
import com.dkhalife.tasks.ui.widget.TaskListWidget
import com.dkhalife.tasks.ui.widget.actions.CompleteTaskAction
import java.time.LocalDateTime

@Composable
fun WidgetTaskRow(
    taskId: Int,
    title: String,
    dueDate: String?,
    labels: List<Label>,
    modifier: GlanceModifier = GlanceModifier,
) {
    val now = LocalDateTime.now()
    val dueLdt = parseDueDate(dueDate)
    val context = androidx.glance.LocalContext.current
    val dueText = if (dueLdt != null) formatDueDate(context, dueLdt, now) else null

    val openIntent = actionStartActivity<MainActivity>(
        actionParametersOf(TaskListWidget.PARAM_TASK_ID to taskId)
    )

    Row(
        modifier = modifier
            .fillMaxWidth()
            .padding(horizontal = 12.dp, vertical = 6.dp)
            .clickable(openIntent),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text = "✓",
            style = TextStyle(
                color = GlanceTheme.colors.primary,
                fontSize = 18.sp,
                fontWeight = FontWeight.Bold
            ),
            modifier = GlanceModifier
                .padding(end = 8.dp)
                .size(28.dp)
                .clickable(
                    actionRunCallback<CompleteTaskAction>(
                        actionParametersOf(CompleteTaskAction.PARAM_TASK_ID to taskId)
                    )
                )
        )

        Column(modifier = GlanceModifier.defaultWeight()) {
            Text(
                text = title,
                style = TextStyle(
                    color = GlanceTheme.colors.onSurface,
                    fontSize = 14.sp,
                    fontWeight = FontWeight.Medium
                ),
                maxLines = 1
            )
            Row(
                horizontalAlignment = Alignment.Start,
                modifier = GlanceModifier.fillMaxWidth()
            ) {
                if (dueText != null) {
                    val isOverdue = dueLdt != null && dueLdt.isBefore(now)
                    Text(
                        text = dueText,
                        style = TextStyle(
                            color = if (isOverdue) GlanceTheme.colors.error else GlanceTheme.colors.onSurfaceVariant,
                            fontSize = 11.sp
                        ),
                        maxLines = 1
                    )
                }
                if (labels.isNotEmpty()) {
                    Text(
                        text = if (dueText != null) " · " else "",
                        style = TextStyle(
                            color = GlanceTheme.colors.onSurfaceVariant,
                            fontSize = 11.sp
                        )
                    )
                    Text(
                        text = labels.joinToString(", ") { it.name },
                        style = TextStyle(
                            color = GlanceTheme.colors.onSurfaceVariant,
                            fontSize = 11.sp
                        ),
                        maxLines = 1
                    )
                }
            }
        }
    }
}
