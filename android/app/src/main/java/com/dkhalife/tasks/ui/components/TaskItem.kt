package com.dkhalife.tasks.ui.components

import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Block
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material.icons.filled.EventBusy
import androidx.compose.material.icons.filled.RadioButtonUnchecked
import androidx.compose.material.icons.filled.SkipNext
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.SwipeToDismissBox
import androidx.compose.material3.SwipeToDismissBoxValue
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.rememberSwipeToDismissBoxState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.hapticfeedback.HapticFeedbackType
import androidx.compose.ui.platform.LocalHapticFeedback
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.semantics.CustomAccessibilityAction
import androidx.compose.ui.semantics.customActions
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.R
import com.dkhalife.tasks.data.SwipeAction
import com.dkhalife.tasks.data.SwipeSettings
import com.dkhalife.tasks.model.FrequencyType
import com.dkhalife.tasks.model.Task
import com.dkhalife.tasks.ui.utils.hasActiveNotification
import com.dkhalife.tasks.ui.utils.parseDueDate
import com.dkhalife.tasks.ui.utils.rememberTickingNow

@OptIn(ExperimentalMaterial3Api::class, ExperimentalFoundationApi::class)
@Composable
fun TaskItem(
    task: Task,
    onComplete: () -> Unit,
    onSkip: () -> Unit,
    onDelete: () -> Unit,
    onClick: () -> Unit,
    onCompleteAndEndRecurrence: () -> Unit = {},
    swipeSettings: SwipeSettings = SwipeSettings(),
    inlineCompleteEnabled: Boolean = true
){
    val ldt = remember(task.nextDueDate) { parseDueDate(task.nextDueDate) }
    val now by rememberTickingNow()
    val haptic = LocalHapticFeedback.current
    val isRecurring = task.frequency.type != FrequencyType.ONCE

    val completeLabel = stringResource(R.string.action_complete)
    val skipLabel = stringResource(R.string.action_skip)
    val deleteLabel = stringResource(R.string.action_delete)
    val editLabel = stringResource(R.string.action_edit)
    val completeEndRecurrenceLabel = stringResource(R.string.action_complete_end_recurrence)

    var showDeleteConfirmation by remember { mutableStateOf(false) }
    var showContextMenu by remember { mutableStateOf(false) }

    if (showDeleteConfirmation) {
        AlertDialog(
            onDismissRequest = { showDeleteConfirmation = false },
            title = { Text(stringResource(R.string.dialog_delete_title)) },
            text = { Text(stringResource(R.string.dialog_delete_message, task.title)) },
            confirmButton = {
                TextButton(
                    onClick = {
                        showDeleteConfirmation = false
                        onDelete()
                    },
                    colors = androidx.compose.material3.ButtonDefaults.textButtonColors(
                        contentColor = MaterialTheme.colorScheme.error
                    )
                ) {
                    Text(stringResource(R.string.btn_delete))
                }
            },
            dismissButton = {
                TextButton(onClick = { showDeleteConfirmation = false }) {
                    Text(stringResource(R.string.btn_cancel))
                }
            }
        )
    }

    val accessibilityActions = buildList {
        add(CustomAccessibilityAction(completeLabel) { onComplete(); true })
        if (isRecurring) {
            add(CustomAccessibilityAction(skipLabel) { onSkip(); true })
            add(CustomAccessibilityAction(completeEndRecurrenceLabel) { onCompleteAndEndRecurrence(); true })
        }
        add(CustomAccessibilityAction(deleteLabel) {
            if (swipeSettings.deleteConfirmationEnabled) {
                showDeleteConfirmation = true
            } else {
                onDelete()
            }
            true
        })
    }

    val taskCard = @Composable {
        Box {
            Card(
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.surface
                ),
                elevation = CardDefaults.cardElevation(defaultElevation = 1.dp),
                modifier = Modifier
                    .fillMaxWidth()
                    .combinedClickable(
                        onClick = onClick,
                        onLongClick = {
                            haptic.performHapticFeedback(HapticFeedbackType.LongPress)
                            showContextMenu = true
                        }
                    )
                    .semantics { customActions = accessibilityActions }
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(12.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    if (inlineCompleteEnabled) {
                        IconButton(onClick = onComplete) {
                            Icon(
                                Icons.Default.RadioButtonUnchecked,
                                contentDescription = completeLabel,
                                tint = MaterialTheme.colorScheme.primary
                            )
                        }
                    }

                    Column(modifier = Modifier.weight(1f)) {
                        Row(
                            horizontalArrangement = Arrangement.spacedBy(4.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            modifier = Modifier.padding(bottom = 4.dp)
                        ) {
                            DueDateChip(ldt, now)
                            RecurrenceChip(task, ldt)
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
                }
            }

            DropdownMenu(
                expanded = showContextMenu,
                onDismissRequest = { showContextMenu = false }
            ) {
                DropdownMenuItem(
                    text = { Text(completeLabel) },
                    leadingIcon = { Icon(Icons.Default.Check, contentDescription = null) },
                    onClick = {
                        showContextMenu = false
                        onComplete()
                    }
                )
                DropdownMenuItem(
                    text = { Text(editLabel) },
                    leadingIcon = { Icon(Icons.Default.Edit, contentDescription = null) },
                    onClick = {
                        showContextMenu = false
                        onClick()
                    }
                )
                if (isRecurring) {
                    DropdownMenuItem(
                        text = { Text(skipLabel) },
                        leadingIcon = { Icon(Icons.Default.SkipNext, contentDescription = null) },
                        onClick = {
                            showContextMenu = false
                            onSkip()
                        }
                    )
                    DropdownMenuItem(
                        text = { Text(completeEndRecurrenceLabel) },
                        leadingIcon = { Icon(Icons.Default.EventBusy, contentDescription = null) },
                        onClick = {
                            showContextMenu = false
                            onCompleteAndEndRecurrence()
                        }
                    )
                }
                DropdownMenuItem(
                    text = { Text(deleteLabel) },
                    leadingIcon = {
                        Icon(
                            Icons.Default.Delete,
                            contentDescription = null,
                            tint = MaterialTheme.colorScheme.error
                        )
                    },
                    onClick = {
                        showContextMenu = false
                        if (swipeSettings.deleteConfirmationEnabled) {
                            showDeleteConfirmation = true
                        } else {
                            onDelete()
                        }
                    }
                )
            }
        }
    }

    if (!swipeSettings.enabled) {
        taskCard()
        return
    }

    val isOnce = task.frequency.type == FrequencyType.ONCE
    val effectiveStartToEnd = if (isOnce && swipeSettings.startToEndAction == SwipeAction.SKIP) SwipeAction.NONE else swipeSettings.startToEndAction
    val effectiveEndToStart = if (isOnce && swipeSettings.endToStartAction == SwipeAction.SKIP) SwipeAction.NONE else swipeSettings.endToStartAction

    fun triggerAction(action: SwipeAction) {
        when (action) {
            SwipeAction.COMPLETE -> onComplete()
            SwipeAction.DELETE -> if (swipeSettings.deleteConfirmationEnabled) {
                showDeleteConfirmation = true
            } else {
                onDelete()
            }
            SwipeAction.SKIP -> onSkip()
            SwipeAction.NONE -> {}
        }
    }

    val dismissState = rememberSwipeToDismissBoxState()

    LaunchedEffect(dismissState.currentValue) {
        when (dismissState.currentValue) {
            SwipeToDismissBoxValue.StartToEnd -> {
                triggerAction(effectiveStartToEnd)
                dismissState.snapTo(SwipeToDismissBoxValue.Settled)
            }
            SwipeToDismissBoxValue.EndToStart -> {
                triggerAction(effectiveEndToStart)
                dismissState.snapTo(SwipeToDismissBoxValue.Settled)
            }
            SwipeToDismissBoxValue.Settled -> {}
        }
    }

    SwipeToDismissBox(
        state = dismissState,
        enableDismissFromStartToEnd = effectiveStartToEnd != SwipeAction.NONE,
        enableDismissFromEndToStart = effectiveEndToStart != SwipeAction.NONE,
        backgroundContent = {
            val isStartToEnd = dismissState.dismissDirection == SwipeToDismissBoxValue.StartToEnd
            val action = if (isStartToEnd) effectiveStartToEnd else effectiveEndToStart
            val alignment = if (isStartToEnd) Alignment.CenterStart else Alignment.CenterEnd
            val icon = when (action) {
                SwipeAction.COMPLETE -> Icons.Default.Check
                SwipeAction.DELETE -> Icons.Default.Delete
                SwipeAction.SKIP -> Icons.Default.SkipNext
                SwipeAction.NONE -> Icons.Default.Block
            }
            val containerColor = when (action) {
                SwipeAction.COMPLETE -> MaterialTheme.colorScheme.primaryContainer
                SwipeAction.DELETE -> MaterialTheme.colorScheme.errorContainer
                SwipeAction.SKIP -> MaterialTheme.colorScheme.tertiaryContainer
                SwipeAction.NONE -> MaterialTheme.colorScheme.surfaceVariant
            }
            val contentColor = when (action) {
                SwipeAction.COMPLETE -> MaterialTheme.colorScheme.onPrimaryContainer
                SwipeAction.DELETE -> MaterialTheme.colorScheme.onErrorContainer
                SwipeAction.SKIP -> MaterialTheme.colorScheme.onTertiaryContainer
                SwipeAction.NONE -> MaterialTheme.colorScheme.onSurfaceVariant
            }

            Card(
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = containerColor)
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(horizontal = 20.dp),
                    contentAlignment = alignment
                ) {
                    Icon(icon, contentDescription = null, tint = contentColor)
                }
            }
        }
    ) {
        taskCard()
    }
}
