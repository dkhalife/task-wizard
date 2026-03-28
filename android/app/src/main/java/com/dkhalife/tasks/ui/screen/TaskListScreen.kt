package com.dkhalife.tasks.ui.screen

import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.CheckCircleOutline
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExtendedFloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.LargeTopAppBar
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.input.nestedscroll.nestedScroll
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.R
import com.dkhalife.tasks.data.SwipeSettings
import com.dkhalife.tasks.data.TaskGroup
import com.dkhalife.tasks.ui.components.GroupHeader
import com.dkhalife.tasks.ui.components.TaskItem

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
    onCompleteAndEndRecurrenceTask: (Int) -> Unit,
    onTaskClick: (Int) -> Unit,
    onCreateTask: () -> Unit,
    onToggleGroup: (String) -> Unit,
    swipeSettings: SwipeSettings = SwipeSettings(),
    inlineCompleteEnabled: Boolean = true
){
    val scrollBehavior = TopAppBarDefaults.exitUntilCollapsedScrollBehavior()
    val newTaskLabel = stringResource(R.string.btn_new_task)

    Scaffold(
        modifier = Modifier.nestedScroll(scrollBehavior.nestedScrollConnection),
        topBar = {
            LargeTopAppBar(
                title = { Text(stringResource(R.string.nav_tasks)) },
                scrollBehavior = scrollBehavior,
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.background,
                    scrolledContainerColor = MaterialTheme.colorScheme.surface
                )
            )
        },
        floatingActionButton = {
            ExtendedFloatingActionButton(
                onClick = onCreateTask,
                icon = { Icon(Icons.Default.Add, contentDescription = null) },
                text = { Text(newTaskLabel) },
                containerColor = MaterialTheme.colorScheme.primaryContainer,
                contentColor = MaterialTheme.colorScheme.onPrimaryContainer
            )
        }
    ) { padding ->
        PullToRefreshBox(
            isRefreshing = isRefreshing,
            onRefresh = onRefresh,
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            AnimatedContent(
                targetState = taskGroups.isEmpty() && !isRefreshing,
                transitionSpec = { fadeIn() togetherWith fadeOut() },
                label = "task-list-content"
            ) { isEmpty ->
                if (isEmpty) {
                    Box(
                        modifier = Modifier.fillMaxSize(),
                        contentAlignment = Alignment.Center
                    ) {
                        Column(horizontalAlignment = Alignment.CenterHorizontally) {
                            Icon(
                                Icons.Default.CheckCircleOutline,
                                contentDescription = null,
                                modifier = Modifier.size(64.dp),
                                tint = MaterialTheme.colorScheme.primary.copy(alpha = 0.6f)
                            )
                            Spacer(modifier = Modifier.height(16.dp))
                            Text(
                                text = stringResource(R.string.task_list_empty_title),
                                style = MaterialTheme.typography.titleMedium,
                                color = MaterialTheme.colorScheme.onSurface
                            )
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(
                                text = stringResource(R.string.task_list_empty_hint, newTaskLabel),
                                style = MaterialTheme.typography.bodyMedium,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    }
                } else {
                    LazyColumn(
                        modifier = Modifier.fillMaxSize(),
                        contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp),
                        verticalArrangement = Arrangement.spacedBy(6.dp)
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
                                        onClick = { onTaskClick(task.id) },
                                        onCompleteAndEndRecurrence = { onCompleteAndEndRecurrenceTask(task.id) },
                                        swipeSettings = swipeSettings,
                                        inlineCompleteEnabled = inlineCompleteEnabled
                                    )
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
