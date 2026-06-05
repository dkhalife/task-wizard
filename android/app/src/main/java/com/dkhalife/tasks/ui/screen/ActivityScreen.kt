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
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Undo
import androidx.compose.material.icons.filled.History
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.R
import com.dkhalife.tasks.model.ActivityEntry
import com.dkhalife.tasks.ui.utils.formatDistance
import com.dkhalife.tasks.ui.utils.parseIsoDateTime
import java.time.LocalDateTime
import java.time.format.DateTimeFormatter
import java.time.temporal.ChronoUnit
import java.util.Locale

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ActivityScreen(
    items: List<ActivityEntry>,
    isLoading: Boolean,
    isLoadingMore: Boolean,
    hasMore: Boolean,
    isReverting: Boolean,
    message: String?,
    onRevert: (taskId: Int, historyId: Int) -> Unit,
    onLoadMore: () -> Unit,
    onMessageShown: () -> Unit,
) {
    val snackbarHostState = remember { SnackbarHostState() }

    LaunchedEffect(message) {
        message?.let {
            snackbarHostState.showSnackbar(it)
            onMessageShown()
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.activity_title)) },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.background
                )
            )
        },
        snackbarHost = { SnackbarHost(snackbarHostState) }
    ) { padding ->
        when {
            isLoading && items.isEmpty() -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding),
                    contentAlignment = Alignment.Center
                ) {
                    CircularProgressIndicator()
                }
            }
            items.isEmpty() -> EmptyActivityContent(Modifier.padding(padding))
            else -> ActivityContent(
                items = items,
                isLoadingMore = isLoadingMore,
                hasMore = hasMore,
                isReverting = isReverting,
                onRevert = onRevert,
                onLoadMore = onLoadMore,
                modifier = Modifier.padding(padding)
            )
        }
    }
}

@Composable
private fun EmptyActivityContent(modifier: Modifier = Modifier) {
    Box(
        modifier = modifier.fillMaxSize(),
        contentAlignment = Alignment.Center
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Icon(
                Icons.Default.History,
                contentDescription = null,
                modifier = Modifier.size(64.dp),
                tint = MaterialTheme.colorScheme.primary.copy(alpha = 0.6f)
            )
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = stringResource(R.string.activity_empty_title),
                style = MaterialTheme.typography.titleMedium,
                color = MaterialTheme.colorScheme.onSurface
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = stringResource(R.string.activity_empty_hint),
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

@Composable
private fun ActivityContent(
    items: List<ActivityEntry>,
    isLoadingMore: Boolean,
    hasMore: Boolean,
    isReverting: Boolean,
    onRevert: (taskId: Int, historyId: Int) -> Unit,
    onLoadMore: () -> Unit,
    modifier: Modifier = Modifier
) {
    val now = LocalDateTime.now()

    LazyColumn(
        modifier = modifier.fillMaxSize(),
        contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        items(items, key = { it.id }) { entry ->
            ActivityCard(entry = entry, now = now, isReverting = isReverting, onRevert = onRevert)
        }

        if (hasMore) {
            item(key = "load_more") {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(vertical = 8.dp),
                    contentAlignment = Alignment.Center
                ) {
                    if (isLoadingMore) {
                        CircularProgressIndicator(modifier = Modifier.size(24.dp))
                    } else {
                        OutlinedButton(onClick = onLoadMore) {
                            Text(stringResource(R.string.activity_load_more))
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun ActivityCard(
    entry: ActivityEntry,
    now: LocalDateTime,
    isReverting: Boolean,
    onRevert: (taskId: Int, historyId: Int) -> Unit
) {
    val context = LocalContext.current
    val completedZdt = parseIsoDateTime(entry.completedDate)
    val isCompleted = completedZdt != null

    Card(
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        ),
        elevation = CardDefaults.cardElevation(defaultElevation = 1.dp),
        modifier = Modifier.fillMaxWidth()
    ) {
        Column(modifier = Modifier.padding(12.dp)) {
            Text(
                text = entry.taskTitle,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold
            )

            Spacer(modifier = Modifier.height(4.dp))

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
                        val delayHours = ChronoUnit.HOURS.between(dueZdt, completedZdt)
                        val chipColor = when {
                            delayHours <= 0 -> MaterialTheme.colorScheme.primary
                            delayHours <= 24 -> MaterialTheme.colorScheme.tertiary
                            else -> MaterialTheme.colorScheme.error
                        }
                        val statusText = if (delayHours <= 0) {
                            stringResource(R.string.history_on_time)
                        } else {
                            stringResource(R.string.history_late_suffix, formatDurationHours(context, delayHours))
                        }
                        Text(
                            text = statusText,
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

            parseIsoDateTime(entry.dueDate)?.let { dueZdt ->
                Text(
                    text = stringResource(
                        R.string.history_due_ago,
                        formatDistance(context, dueZdt.toLocalDateTime(), now)
                    ),
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            if (entry.isLatest) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.End
                ) {
                    TextButton(
                        onClick = { onRevert(entry.taskId, entry.id) },
                        enabled = !isReverting
                    ) {
                        Icon(
                            Icons.AutoMirrored.Filled.Undo,
                            contentDescription = null,
                            modifier = Modifier.size(18.dp)
                        )
                        Spacer(modifier = Modifier.size(4.dp))
                        Text(stringResource(R.string.activity_revert))
                    }
                }
            }
        }
    }
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
