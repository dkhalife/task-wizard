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
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.CheckCircleOutline
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExtendedFloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.LargeTopAppBar
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextField
import androidx.compose.material3.TextFieldDefaults
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.nestedscroll.nestedScroll
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextRange
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.TextFieldValue
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.R
import com.dkhalife.tasks.data.SwipeSettings
import com.dkhalife.tasks.data.TaskGroup
import com.dkhalife.tasks.ui.components.GroupHeader
import com.dkhalife.tasks.ui.components.SyncStatusBanner
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
    onViewHistory: (Int) -> Unit,
    onCreateTask: () -> Unit,
    onToggleGroup: (String) -> Unit,
    searchQuery: String = "",
    isSearchActive: Boolean = false,
    onSearchQueryChange: (String) -> Unit = {},
    onSearchActiveChange: (Boolean) -> Unit = {},
    swipeSettings: SwipeSettings = SwipeSettings(),
    inlineCompleteEnabled: Boolean = true,
    isPendingDeletion: Boolean = false,
    isOnline: Boolean = true,
    pendingSyncCount: Int = 0,
){
    val scrollBehavior = TopAppBarDefaults.exitUntilCollapsedScrollBehavior()
    val newTaskLabel = stringResource(R.string.btn_new_task)
    val focusRequester = remember { FocusRequester() }

    // TextFieldValue is kept locally to preserve cursor position — only .text is pushed to the
    // ViewModel, so the ViewModel never drives the cursor back, preventing cursor-jump bugs.
    // Using only isSearchActive as the remember key prevents the state being recreated on every
    // keystroke (which would happen if searchQuery were a key).
    var textFieldValue by remember(isSearchActive) {
        mutableStateOf(TextFieldValue(searchQuery, TextRange(searchQuery.length)))
    }

    // Re-sync from ViewModel only when the text differs (e.g. navigation restoration), not during
    // normal typing where textFieldValue.text already matches searchQuery.
    LaunchedEffect(searchQuery) {
        if (textFieldValue.text != searchQuery) {
            textFieldValue = TextFieldValue(searchQuery, TextRange(searchQuery.length))
        }
    }

    LaunchedEffect(isSearchActive) {
        if (isSearchActive) {
            focusRequester.requestFocus()
        }
    }

    Scaffold(
        modifier = if (!isSearchActive) Modifier.nestedScroll(scrollBehavior.nestedScrollConnection) else Modifier,
        topBar = {
            AnimatedContent(
                targetState = isSearchActive,
                transitionSpec = { fadeIn() togetherWith fadeOut() },
                label = "top-bar-search-transition"
            ) { searching ->
                if (searching) {
                    TopAppBar(
                        title = {
                            TextField(
                                value = textFieldValue,
                                onValueChange = { newValue ->
                                    textFieldValue = newValue
                                    onSearchQueryChange(newValue.text)
                                },
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .focusRequester(focusRequester),
                                placeholder = {
                                    Text(
                                        text = stringResource(R.string.search_hint),
                                        style = MaterialTheme.typography.bodyLarge,
                                        color = MaterialTheme.colorScheme.onSurfaceVariant
                                    )
                                },
                                textStyle = MaterialTheme.typography.bodyLarge.copy(
                                    color = MaterialTheme.colorScheme.onSurface
                                ),
                                singleLine = true,
                                keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search),
                                colors = TextFieldDefaults.colors(
                                    focusedContainerColor = Color.Transparent,
                                    unfocusedContainerColor = Color.Transparent,
                                    focusedIndicatorColor = Color.Transparent,
                                    unfocusedIndicatorColor = Color.Transparent,
                                    disabledIndicatorColor = Color.Transparent,
                                )
                            )
                        },
                        actions = {
                            IconButton(onClick = {
                                textFieldValue = TextFieldValue("")
                                onSearchActiveChange(false)
                            }) {
                                Icon(
                                    Icons.Default.Close,
                                    contentDescription = stringResource(R.string.search_clear_description)
                                )
                            }
                        },
                        colors = TopAppBarDefaults.topAppBarColors(
                            containerColor = MaterialTheme.colorScheme.background,
                        )
                    )
                } else {
                    LargeTopAppBar(
                        title = { Text(stringResource(R.string.nav_tasks)) },
                        scrollBehavior = scrollBehavior,
                        actions = {
                            IconButton(onClick = { onSearchActiveChange(true) }) {
                                Icon(
                                    Icons.Default.Search,
                                    contentDescription = stringResource(R.string.search_icon_description)
                                )
                            }
                        },
                        colors = TopAppBarDefaults.topAppBarColors(
                            containerColor = MaterialTheme.colorScheme.background,
                            scrolledContainerColor = MaterialTheme.colorScheme.surface
                        )
                    )
                }
            }
        },
        floatingActionButton = {
            if (!isPendingDeletion) {
                ExtendedFloatingActionButton(
                    onClick = onCreateTask,
                    icon = { Icon(Icons.Default.Add, contentDescription = null) },
                    text = { Text(newTaskLabel) },
                    containerColor = MaterialTheme.colorScheme.primaryContainer,
                    contentColor = MaterialTheme.colorScheme.onPrimaryContainer
                )
            }
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            SyncStatusBanner(isOnline = isOnline, pendingSyncCount = pendingSyncCount)
            if (isPendingDeletion) {
                androidx.compose.material3.Surface(
                    color = MaterialTheme.colorScheme.errorContainer,
                    modifier = Modifier.fillMaxWidth()
                ) {
                    androidx.compose.foundation.layout.Row(
                        modifier = Modifier.padding(horizontal = 16.dp, vertical = 10.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(
                            Icons.Default.Warning,
                            contentDescription = null,
                            tint = MaterialTheme.colorScheme.onErrorContainer,
                            modifier = Modifier.size(18.dp)
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = stringResource(R.string.settings_section_account_deletion),
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onErrorContainer
                        )
                    }
                }
            }

            PullToRefreshBox(
                isRefreshing = isRefreshing,
                onRefresh = onRefresh,
                modifier = Modifier.fillMaxSize()
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
                            if (isSearchActive && searchQuery.isNotBlank()) {
                                Text(
                                    text = stringResource(R.string.search_no_results_title),
                                    style = MaterialTheme.typography.titleMedium,
                                    color = MaterialTheme.colorScheme.onSurface
                                )
                                Spacer(modifier = Modifier.height(4.dp))
                                Text(
                                    text = stringResource(R.string.search_no_results_hint),
                                    style = MaterialTheme.typography.bodyMedium,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant
                                )
                            } else {
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
                                        onComplete = if (isPendingDeletion) ({}) else { { onCompleteTask(task.id) } },
                                        onSkip = if (isPendingDeletion) ({}) else { { onSkipTask(task.id) } },
                                        onDelete = if (isPendingDeletion) ({}) else { { onDeleteTask(task.id) } },
                                        onClick = if (isPendingDeletion) ({}) else { { onTaskClick(task.id) } },
                                        onViewHistory = { onViewHistory(task.id) },
                                        onCompleteAndEndRecurrence = if (isPendingDeletion) ({}) else { { onCompleteAndEndRecurrenceTask(task.id) } },
                                        swipeSettings = if (isPendingDeletion) SwipeSettings(enabled = false) else swipeSettings,
                                        inlineCompleteEnabled = inlineCompleteEnabled && !isPendingDeletion
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
}
