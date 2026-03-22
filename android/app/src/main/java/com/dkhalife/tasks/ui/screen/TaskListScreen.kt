package com.dkhalife.tasks.ui.screen

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.ExpandLess
import androidx.compose.material.icons.filled.ExpandMore
import androidx.compose.material.icons.filled.RadioButtonUnchecked
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
import com.dkhalife.tasks.model.Task
import java.time.ZonedDateTime
import java.time.format.DateTimeFormatter

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
                            items(group.tasks, key = { it.id }) { task ->
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
        Surface(
            shape = MaterialTheme.shapes.extraSmall,
            color = group.color.copy(alpha = 0.2f),
            modifier = Modifier.size(12.dp)
        ) {}

        Spacer(modifier = Modifier.width(8.dp))

        Text(
            text = group.name,
            style = MaterialTheme.typography.titleSmall,
            color = group.color
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
                Text(
                    text = task.title,
                    style = MaterialTheme.typography.bodyLarge,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )

                task.nextDueDate?.let { dueDate ->
                    Text(
                        text = formatDueDate(dueDate),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }

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
        val zdt = ZonedDateTime.parse(dateStr)
        zdt.format(DateTimeFormatter.ofPattern("MMM d, yyyy"))
    } catch (_: Exception) {
        dateStr
    }
}
