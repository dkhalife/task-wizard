package com.dkhalife.tasks.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.NotificationsActive
import androidx.compose.material.icons.filled.Repeat
import androidx.compose.material.icons.filled.Schedule
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.R
import com.dkhalife.tasks.model.FrequencyType
import com.dkhalife.tasks.model.Task
import com.dkhalife.tasks.ui.utils.formatDueDate
import com.dkhalife.tasks.ui.utils.getRecurrenceText
import java.time.LocalDateTime
import java.time.temporal.ChronoUnit

private val ChipShape = RoundedCornerShape(8.dp)

@Composable
fun DueDateChip(ldt: LocalDateTime?, now: LocalDateTime) {
    val context = LocalContext.current
    val text = if (ldt == null) stringResource(R.string.chip_no_due_date) else formatDueDate(context, ldt, now)
    val (bgColor, fgColor) = when {
        ldt == null -> Pair(MaterialTheme.colorScheme.surfaceVariant, MaterialTheme.colorScheme.onSurfaceVariant)
        ldt.isBefore(now) -> Pair(MaterialTheme.colorScheme.errorContainer, MaterialTheme.colorScheme.onErrorContainer)
        ChronoUnit.HOURS.between(now, ldt) < 4 -> Pair(MaterialTheme.colorScheme.tertiaryContainer, MaterialTheme.colorScheme.onTertiaryContainer)
        else -> Pair(MaterialTheme.colorScheme.surfaceVariant, MaterialTheme.colorScheme.onSurfaceVariant)
    }
    Surface(shape = ChipShape, color = bgColor) {
        Row(
            modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(4.dp)
        ) {
            Icon(
                Icons.Default.Schedule,
                contentDescription = null,
                tint = fgColor,
                modifier = Modifier.size(12.dp)
            )
            Text(
                text = text,
                style = MaterialTheme.typography.labelSmall,
                color = fgColor
            )
        }
    }
}

@Composable
fun RecurrenceChip(task: Task, nextDueLdt: LocalDateTime?) {
    val context = LocalContext.current
    val text = getRecurrenceText(context, task, nextDueLdt)
    val isOnce = task.frequency.type == FrequencyType.ONCE
    Surface(
        shape = ChipShape,
        color = MaterialTheme.colorScheme.surfaceVariant
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(3.dp)
        ) {
            if (isOnce) {
                Text(
                    text = "1\u00D7",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            } else {
                Icon(
                    Icons.Default.Repeat,
                    contentDescription = null,
                    tint = MaterialTheme.colorScheme.onSurfaceVariant,
                    modifier = Modifier.size(12.dp)
                )
                if (text.isNotEmpty()) {
                    Text(
                        text = text,
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
        }
    }
}

@Composable
fun NotificationChip() {
    Surface(
        shape = ChipShape,
        color = MaterialTheme.colorScheme.surfaceVariant
    ) {
        Icon(
            Icons.Default.NotificationsActive,
            contentDescription = stringResource(R.string.notifications_active_description),
            tint = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier
                .padding(horizontal = 6.dp, vertical = 4.dp)
                .size(12.dp)
        )
    }
}

@Composable
fun LabelChip(name: String, color: String) {
    val chipColor = try {
        Color(android.graphics.Color.parseColor(color))
    } catch (_: Exception) {
        MaterialTheme.colorScheme.primary
    }
    Surface(
        shape = ChipShape,
        color = chipColor.copy(alpha = 0.15f)
    ) {
        Text(
            text = name,
            style = MaterialTheme.typography.labelSmall,
            color = chipColor,
            modifier = Modifier.padding(horizontal = 8.dp, vertical = 3.dp)
        )
    }
}
