package com.dkhalife.tasks.ui.screen

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.ExpandLess
import androidx.compose.material.icons.filled.ExpandMore
import androidx.compose.material.icons.filled.NotificationsActive
import androidx.compose.material.icons.filled.RadioButtonUnchecked
import androidx.compose.material.icons.filled.Repeat
import androidx.compose.material.icons.filled.SkipNext
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material3.*
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.data.TaskGroup
import com.dkhalife.tasks.model.FrequencyType
import com.dkhalife.tasks.model.IntervalUnit
import com.dkhalife.tasks.model.RepeatOn
import com.dkhalife.tasks.model.Task
import java.time.LocalDate
import java.time.LocalDateTime
import java.time.ZoneId
import java.time.ZonedDateTime
import java.time.format.DateTimeFormatter
import java.time.temporal.ChronoUnit
import java.util.Locale

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TaskListScreen(
    taskGroups: List<TaskGroup>,
    expandedGroups: Set<String>,
    isRefreshing: Boolean,
    onRefresh: () -> Unit,
    onCompleteTask: (Int) -> Unit,
    onSkipTask: (Int) -> Unit,
    onDeleteTask: (Int) -> Unit,
    onTaskClick: (Int) -> Unit,
    onCreateTask: () -> Unit,
    onToggleGroup: (String) -> Unit
) {
    Scaffold(
        floatingActionButton = {
            FloatingActionButton(onClick = onCreateTask) {
                Icon(Icons.Default.Add, contentDescription = "Create task")
            }
        }
    ) { padding ->
        PullToRefreshBox(
            isRefreshing = isRefreshing,
            onRefresh = onRefresh,
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            if (taskGroups.isEmpty() && !isRefreshing) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "No tasks yet. Tap + to create one.",
                        style = MaterialTheme.typography.bodyLarge,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            } else {
                LazyColumn(
                    modifier = Modifier.fillMaxSize(),
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    taskGroups.forEach { group ->
                        val isExpanded = expandedGroups.contains(group.key)

                        item(key = "header_${group.key}") {
                            GroupHeader(
                                group = group,
                                isExpanded = isExpanded,
                                onToggle = { onToggleGroup(group.key) }
                            )
                        }

                        if (isExpanded) {
                            items(group.tasks, key = { "${group.key}_${it.id}" }) { task ->
                            TaskItem(
                                task = task,
                                onComplete = { onCompleteTask(task.id) },
                                onSkip = { onSkipTask(task.id) },
                                onDelete = { onDeleteTask(task.id) },
                                onClick = { onTaskClick(task.id) }
                            )
                        }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun GroupHeader(group: TaskGroup, isExpanded: Boolean, onToggle: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onToggle)
            .padding(vertical = 4.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        val headerColor = if (group.color == Color.Unspecified) {
            MaterialTheme.colorScheme.onSurface
        } else {
            group.color
        }

        val indicatorColor = if (group.color == Color.Unspecified) {
            MaterialTheme.colorScheme.onSurfaceVariant
        } else {
            group.color
        }

        Surface(
            shape = MaterialTheme.shapes.extraSmall,
            color = indicatorColor.copy(alpha = 0.2f),
            modifier = Modifier.size(12.dp)
        ) {}

        Spacer(modifier = Modifier.width(8.dp))

        Text(
            text = group.name,
            style = MaterialTheme.typography.titleSmall,
            color = headerColor
        )

        Spacer(modifier = Modifier.width(8.dp))

        Text(
            text = "(${group.tasks.size})",
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )

        Spacer(modifier = Modifier.weight(1f))

        Icon(
            imageVector = if (isExpanded) Icons.Default.ExpandLess else Icons.Default.ExpandMore,
            contentDescription = if (isExpanded) "Collapse" else "Expand",
            tint = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.size(20.dp)
        )
    }
}

@Composable
private fun TaskItem(
    task: Task,
    onComplete: () -> Unit,
    onSkip: () -> Unit,
    onDelete: () -> Unit,
    onClick: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            IconButton(onClick = onComplete) {
                Icon(
                    Icons.Default.RadioButtonUnchecked,
                    contentDescription = "Complete",
                    tint = MaterialTheme.colorScheme.primary
                )
            }

            Column(
                modifier = Modifier.weight(1f)
            ) {
                Row(
                    horizontalArrangement = Arrangement.spacedBy(4.dp),
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier.padding(bottom = 4.dp)
                ) {
                    DueDateChip(task.nextDueDate)
                    RecurrenceChip(task)
                    if (hasActiveNotification(task)) {
                        NotificationChip()
                    }
                }

                Text(
                    text = task.title,
                    style = MaterialTheme.typography.bodyLarge,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )

                if (task.labels.isNotEmpty()) {
                    Row(
                        horizontalArrangement = Arrangement.spacedBy(4.dp),
                        modifier = Modifier.padding(top = 4.dp)
                    ) {
                        task.labels.forEach { label ->
                            LabelChip(name = label.name, color = label.color)
                        }
                    }
                }
            }

            if (task.frequency.type != "once") {
                IconButton(onClick = onSkip) {
                    Icon(
                        Icons.Default.SkipNext,
                        contentDescription = "Skip",
                        tint = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }

            IconButton(onClick = onDelete) {
                Icon(
                    Icons.Default.Delete,
                    contentDescription = "Delete",
                    tint = MaterialTheme.colorScheme.error
                )
            }
        }
    }
}

@Composable
private fun DueDateChip(dateStr: String?) {
    val ldt = remember(dateStr) {
        dateStr?.let {
            try {
                ZonedDateTime.parse(it).withZoneSameInstant(ZoneId.systemDefault()).toLocalDateTime()
            } catch (_: Exception) { null }
        }
    }
    val now = LocalDateTime.now()
    val text = if (dateStr == null) "No Due Date" else formatDueDate(dateStr)
    val (bgColor, fgColor) = when {
        ldt == null -> Pair(MaterialTheme.colorScheme.surfaceVariant, MaterialTheme.colorScheme.onSurfaceVariant)
        ldt.isBefore(now) -> Pair(MaterialTheme.colorScheme.errorContainer, MaterialTheme.colorScheme.onErrorContainer)
        ChronoUnit.HOURS.between(now, ldt) < 4 -> Pair(MaterialTheme.colorScheme.tertiaryContainer, MaterialTheme.colorScheme.onTertiaryContainer)
        else -> Pair(MaterialTheme.colorScheme.surfaceVariant, MaterialTheme.colorScheme.onSurfaceVariant)
    }
    Surface(shape = MaterialTheme.shapes.extraSmall, color = bgColor) {
        Text(
            text = text,
            style = MaterialTheme.typography.labelSmall,
            color = fgColor,
            modifier = Modifier.padding(horizontal = 8.dp, vertical = 3.dp)
        )
    }
}

@Composable
private fun RecurrenceChip(task: Task) {
    val nextDueLdt = remember(task.nextDueDate) {
        task.nextDueDate?.let {
            try {
                ZonedDateTime.parse(it).withZoneSameInstant(ZoneId.systemDefault()).toLocalDateTime()
            } catch (_: Exception) { null }
        }
    }
    val text = getRecurrenceText(task, nextDueLdt)
    val isOnce = task.frequency.type == FrequencyType.ONCE
    Surface(
        shape = MaterialTheme.shapes.extraSmall,
        color = MaterialTheme.colorScheme.surfaceVariant
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 8.dp, vertical = 3.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(2.dp)
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
            }
            Text(
                text = text,
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

@Composable
private fun NotificationChip() {
    Surface(
        shape = MaterialTheme.shapes.extraSmall,
        color = MaterialTheme.colorScheme.surfaceVariant
    ) {
        Icon(
            Icons.Default.NotificationsActive,
            contentDescription = "Notifications active",
            tint = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier
                .padding(horizontal = 6.dp, vertical = 3.dp)
                .size(12.dp)
        )
    }
}

@Composable
private fun LabelChip(name: String, color: String) {
    val chipColor = try {
        Color(android.graphics.Color.parseColor(color))
    } catch (_: Exception) {
        MaterialTheme.colorScheme.primary
    }

    Surface(
        shape = MaterialTheme.shapes.small,
        color = chipColor.copy(alpha = 0.2f)
    ) {
        Text(
            text = name,
            style = MaterialTheme.typography.labelSmall,
            color = chipColor,
            modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp)
        )
    }
}

private fun formatDueDate(dateStr: String): String {
    return try {
        val ldt = ZonedDateTime.parse(dateStr)
            .withZoneSameInstant(ZoneId.systemDefault())
            .toLocalDateTime()
        val now = LocalDateTime.now()
        val today = LocalDate.now()
        val timeStr = ldt.format(DateTimeFormatter.ofPattern("hh:mm a", Locale.ENGLISH))
        when {
            ldt.isBefore(now) -> "${formatDistance(ldt, now)} ago"
            ldt.toLocalDate() == today -> "Today at $timeStr"
            ldt.toLocalDate() == today.plusDays(1) -> "Tomorrow at $timeStr"
            else -> "in ${formatDistance(now, ldt)}"
        }
    } catch (_: Exception) {
        dateStr
    }
}

private fun formatDistance(from: LocalDateTime, to: LocalDateTime): String {
    val seconds = ChronoUnit.SECONDS.between(from, to)
    val minutes = seconds / 60
    return when {
        seconds < 45 -> "less than a minute"
        seconds < 90 -> "1 minute"
        minutes < 45 -> "$minutes minutes"
        minutes < 90 -> "about 1 hour"
        minutes < 1440 -> "about ${minutes / 60} hours"
        minutes < 2520 -> "1 day"
        minutes < 43200 -> "${minutes / 1440} days"
        minutes < 64800 -> "about 1 month"
        minutes < 86400 -> "about 2 months"
        minutes < 525600 -> "${minutes / 43200} months"
        else -> "about ${minutes / 525600} years"
    }
}

private fun getRecurrenceText(task: Task, nextDueLdt: LocalDateTime?): String {
    val frequency = task.frequency
    return when (frequency.type) {
        FrequencyType.ONCE -> "Once"
        FrequencyType.DAILY -> "Daily"
        FrequencyType.WEEKLY -> "Weekly"
        FrequencyType.MONTHLY -> "Monthly"
        FrequencyType.YEARLY -> "Yearly"
        FrequencyType.CUSTOM -> when (frequency.on) {
            RepeatOn.INTERVAL -> {
                val every = frequency.every ?: 1
                if (every == 1) {
                    when (frequency.unit) {
                        IntervalUnit.HOURS -> "Hourly"
                        IntervalUnit.DAYS -> "Daily"
                        IntervalUnit.WEEKS -> "Weekly"
                        IntervalUnit.MONTHS -> "Monthly"
                        IntervalUnit.YEARS -> "Yearly"
                        else -> "Every $every ${frequency.unit}"
                    }
                } else {
                    "Every $every ${frequency.unit}"
                }
            }
            RepeatOn.DAYS_OF_THE_WEEK -> {
                val dayNames = arrayOf("Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat")
                frequency.days?.joinToString(", ") { dayNames.getOrElse(it) { "$it" } } ?: "Weekly"
            }
            RepeatOn.DAY_OF_THE_MONTHS -> {
                val day = nextDueLdt?.dayOfMonth ?: 0
                val suffix = getDayOfMonthSuffix(day)
                val monthNames = arrayOf("Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec")
                val months = frequency.months?.joinToString(", ") {
                    monthNames.getOrElse(it) { "$it" }
                } ?: ""
                "${day}${suffix} of $months"
            }
            else -> ""
        }
        else -> ""
    }
}

private fun getDayOfMonthSuffix(day: Int): String {
    return if (day in 11..13) "th"
    else when (day % 10) {
        1 -> "st"
        2 -> "nd"
        3 -> "rd"
        else -> "th"
    }
}

private fun hasActiveNotification(task: Task): Boolean {
    return task.notification.enabled &&
        (task.notification.dueDate || task.notification.preDue || task.notification.overdue)
}
