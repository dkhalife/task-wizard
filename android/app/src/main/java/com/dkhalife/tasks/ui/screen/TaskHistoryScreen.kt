package com.dkhalife.tasks.ui.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Checklist
import androidx.compose.material.icons.filled.EventBusy
import androidx.compose.material.icons.filled.Timelapse
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.R
import com.dkhalife.tasks.model.TaskHistory
import com.dkhalife.tasks.ui.utils.formatDistance
import com.dkhalife.tasks.ui.utils.parseIsoDateTime
import java.time.LocalDateTime
import java.time.ZoneId
import java.time.format.DateTimeFormatter
import java.time.temporal.ChronoUnit
import java.util.Locale

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TaskHistoryScreen(
    history: List<TaskHistory>,
    isLoading: Boolean,
    onBack: () -> Unit
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.history_title)) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(
                            Icons.AutoMirrored.Filled.ArrowBack,
                            contentDescription = stringResource(R.string.btn_back)
                        )
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.background
                )
            )
        }
    ) { padding ->
        when {
            isLoading -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding),
                    contentAlignment = Alignment.Center
                ) {
                    CircularProgressIndicator()
                }
            }
            history.isEmpty() -> EmptyHistoryContent(Modifier.padding(padding))
            else -> HistoryContent(history, Modifier.padding(padding))
        }
    }
}

@Composable
private fun EmptyHistoryContent(modifier: Modifier = Modifier) {
    Box(
        modifier = modifier.fillMaxSize(),
        contentAlignment = Alignment.Center
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Icon(
                Icons.Default.EventBusy,
                contentDescription = null,
                modifier = Modifier.size(64.dp),
                tint = MaterialTheme.colorScheme.primary.copy(alpha = 0.6f)
            )
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = stringResource(R.string.history_empty_title),
                style = MaterialTheme.typography.titleMedium,
                color = MaterialTheme.colorScheme.onSurface
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = stringResource(R.string.history_empty_hint),
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

@Composable
private fun HistoryContent(
    history: List<TaskHistory>,
    modifier: Modifier = Modifier
) {
    val context = LocalContext.current
    val now = LocalDateTime.now()
    val stats = computeStats(history)

    LazyColumn(
        modifier = modifier.fillMaxSize(),
        contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        item(key = "summary_header") {
            Text(
                text = stringResource(R.string.history_summary_title),
                style = MaterialTheme.typography.titleMedium,
                modifier = Modifier.padding(bottom = 4.dp)
            )
        }

        item(key = "summary_cards") {
            Row(
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                modifier = Modifier.fillMaxWidth()
            ) {
                StatCard(
                    icon = Icons.Default.Checklist,
                    label = stringResource(R.string.history_total_completed),
                    value = stringResource(R.string.history_times, stats.totalCompleted),
                    modifier = Modifier.weight(1f)
                )
                StatCard(
                    icon = Icons.Default.Timelapse,
                    label = stringResource(R.string.history_usually_within),
                    value = stats.averageDelay?.let { formatDurationHours(context, it) } ?: "--",
                    modifier = Modifier.weight(1f)
                )
                StatCard(
                    icon = Icons.Default.Timelapse,
                    label = stringResource(R.string.history_max_delay),
                    value = stats.maxDelay?.let { formatDurationHours(context, it) } ?: "--",
                    modifier = Modifier.weight(1f)
                )
            }
        }

        item(key = "history_header") {
            Text(
                text = stringResource(R.string.history_entries_title),
                style = MaterialTheme.typography.titleMedium,
                modifier = Modifier.padding(top = 8.dp, bottom = 4.dp)
            )
        }

        item(key = "history_card") {
            Card(
                shape = RoundedCornerShape(12.dp),
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.surface
                ),
                elevation = CardDefaults.cardElevation(defaultElevation = 1.dp),
                modifier = Modifier.fillMaxWidth()
            ) {
                Column(modifier = Modifier.padding(12.dp)) {
                    history.forEachIndexed { index, entry ->
                        HistoryEntryRow(entry, now)
                        if (index < history.lastIndex) {
                            val nextEntry = history[index + 1]
                            val dueZdt = parseIsoDateTime(nextEntry.dueDate)
                            val dueText = if (dueZdt != null) {
                                val dueLdt = dueZdt.toLocalDateTime()
                                stringResource(R.string.history_due_ago, formatDistance(context, dueLdt, now))
                            } else {
                                "--"
                            }
                            HorizontalDivider(modifier = Modifier.padding(vertical = 8.dp))
                            Text(
                                text = dueText,
                                style = MaterialTheme.typography.labelSmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant,
                                modifier = Modifier.align(Alignment.CenterHorizontally)
                            )
                            HorizontalDivider(modifier = Modifier.padding(vertical = 8.dp))
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun StatCard(
    icon: ImageVector,
    label: String,
    value: String,
    modifier: Modifier = Modifier
) {
    Card(
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant
        ),
        modifier = modifier
    ) {
        Column(
            modifier = Modifier.padding(12.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Icon(
                icon,
                contentDescription = null,
                tint = MaterialTheme.colorScheme.primary,
                modifier = Modifier.size(20.dp)
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = label,
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Spacer(modifier = Modifier.height(2.dp))
            Text(
                text = value,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.primary
            )
        }
    }
}

@Composable
private fun HistoryEntryRow(entry: TaskHistory, now: LocalDateTime) {
    val context = LocalContext.current
    val completedZdt = parseIsoDateTime(entry.completedDate)
    val isCompleted = completedZdt != null

    Column {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = stringResource(if (isCompleted) R.string.history_completed else R.string.history_skipped),
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.SemiBold
            )

            if (isCompleted) {
                val dueZdt = parseIsoDateTime(entry.dueDate)
                if (dueZdt != null) {
                    val delayHours = ChronoUnit.HOURS.between(dueZdt.toLocalDateTime(), completedZdt.toLocalDateTime())
                    val chipColor = when {
                        delayHours <= 0 -> MaterialTheme.colorScheme.primary
                        delayHours <= 24 -> MaterialTheme.colorScheme.tertiary
                        else -> MaterialTheme.colorScheme.error
                    }
                    Text(
                        text = if (delayHours <= 0) "On time" else formatDurationHours(context, delayHours) + " late",
                        style = MaterialTheme.typography.labelSmall,
                        color = chipColor,
                        fontWeight = FontWeight.Bold
                    )
                }
            }
        }

        if (isCompleted) {
            val formatter = DateTimeFormatter.ofPattern("MMMM d yyyy, h:mm a", Locale.ENGLISH)
            Text(
                text = stringResource(
                    R.string.history_completed_on,
                    completedZdt!!.toLocalDateTime().format(formatter)
                ),
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

private data class HistoryStats(
    val totalCompleted: Int,
    val averageDelay: Long?,
    val maxDelay: Long?
)

private fun computeStats(history: List<TaskHistory>): HistoryStats {
    val withDates = history.filter { entry ->
        entry.dueDate != null && entry.completedDate != null
    }.mapNotNull { entry ->
        val due = parseIsoDateTime(entry.dueDate)
        val completed = parseIsoDateTime(entry.completedDate)
        if (due != null && completed != null) {
            ChronoUnit.HOURS.between(
                due.withZoneSameInstant(ZoneId.systemDefault()),
                completed.withZoneSameInstant(ZoneId.systemDefault())
            )
        } else null
    }

    return HistoryStats(
        totalCompleted = history.size,
        averageDelay = if (withDates.isNotEmpty()) withDates.sum() / withDates.size else null,
        maxDelay = withDates.maxOrNull()
    )
}

private fun formatDurationHours(context: android.content.Context, hours: Long): String {
    val absHours = kotlin.math.abs(hours)
    val minutes = absHours * 60
    return formatDistance(
        context,
        LocalDateTime.now(),
        LocalDateTime.now().plusMinutes(minutes)
    )
}
