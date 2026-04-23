package com.dkhalife.tasks.ui.navigation

import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.slideOutHorizontally
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Icon
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.dkhalife.tasks.data.SwipeSettings
import com.dkhalife.tasks.data.TaskGrouping
import com.dkhalife.tasks.data.ThemeMode
import com.dkhalife.tasks.data.calendar.CalendarRepository
import com.dkhalife.tasks.model.CreateTaskReq
import com.dkhalife.tasks.model.Task
import com.dkhalife.tasks.model.UpdateTaskReq
import com.dkhalife.tasks.ui.screen.LabelsScreen
import com.dkhalife.tasks.ui.screen.SettingsScreen
import com.dkhalife.tasks.ui.screen.SwipeActionsSettingsScreen
import com.dkhalife.tasks.ui.screen.TaskFormScreen
import com.dkhalife.tasks.ui.screen.TaskHistoryScreen
import com.dkhalife.tasks.ui.screen.TaskListScreen
import com.dkhalife.tasks.viewmodel.AuthViewModel
import com.dkhalife.tasks.viewmodel.LabelViewModel
import com.dkhalife.tasks.viewmodel.TaskFormViewModel
import com.dkhalife.tasks.viewmodel.TaskHistoryViewModel
import com.dkhalife.tasks.viewmodel.TaskListViewModel
import com.dkhalife.tasks.viewmodel.UserViewModel

@Composable
fun AppNavigation(
    themeMode: ThemeMode,
    onThemeModeChanged: (ThemeMode) -> Unit,
    taskGrouping: TaskGrouping,
    onTaskGroupingChanged: (TaskGrouping) -> Unit,
    calendarSyncEnabled: Boolean,
    onCalendarSyncChanged: (Boolean) -> Unit,
    calendarRepository: CalendarRepository,
    swipeSettings: SwipeSettings,
    onSwipeEnabledChanged: (Boolean) -> Unit,
    onSwipeStartToEndActionChanged: (com.dkhalife.tasks.data.SwipeAction) -> Unit,
    onSwipeEndToStartActionChanged: (com.dkhalife.tasks.data.SwipeAction) -> Unit,
    onSwipeDeleteConfirmationChanged: (Boolean) -> Unit,
    inlineCompleteEnabled: Boolean,
    onInlineCompleteEnabledChanged: (Boolean) -> Unit,
    telemetryEnabled: Boolean,
    onTelemetryEnabledChanged: (Boolean) -> Unit,
    debugLoggingEnabled: Boolean,
    onDebugLoggingEnabledChanged: (Boolean) -> Unit,
    initialTaskId: Int = -1,
    createTask: Boolean = false
){
    val navController = rememberNavController()
    val bottomScreens = listOf(Screen.Tasks, Screen.Labels, Screen.Settings)
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route

    val showBottomBar = bottomScreens.any { it.route == currentRoute }

    LaunchedEffect(initialTaskId) {
        if (initialTaskId > 0) {
            navController.navigate(Routes.taskFormEdit(initialTaskId))
        }
    }

    LaunchedEffect(createTask) {
        if (createTask) {
            navController.navigate(Routes.TASK_FORM_CREATE)
        }
    }

    Scaffold(
        contentWindowInsets = WindowInsets(0, 0, 0, 0),
        bottomBar = {
            if (showBottomBar) {
                NavigationBar {
                    bottomScreens.forEach { screen ->
                        NavigationBarItem(
                            icon = { Icon(screen.icon, contentDescription = stringResource(screen.titleRes)) },
                            label = { Text(stringResource(screen.titleRes)) },
                            selected = currentRoute == screen.route,
                            onClick = {
                                navController.navigate(screen.route) {
                                    popUpTo(navController.graph.findStartDestination().id) {
                                        saveState = true
                                    }
                                    launchSingleTop = true
                                    restoreState = true
                                }
                            }
                        )
                    }
                }
            }
        }
    ) { innerPadding ->
        NavHost(
            navController = navController,
            startDestination = Screen.Tasks.route,
            modifier = Modifier.padding(innerPadding),
            enterTransition = { fadeIn() },
            exitTransition = { fadeOut() },
            popEnterTransition = { fadeIn() },
            popExitTransition = { fadeOut() }
        ) {
            composable(Screen.Tasks.route) {
                val viewModel: TaskListViewModel = hiltViewModel()
                val userViewModel: UserViewModel = hiltViewModel()
                val isRefreshing by viewModel.isRefreshing.collectAsState()
                val taskGroups by viewModel.taskGroups.collectAsState()
                val expandedGroups by viewModel.expandedGroups.collectAsState()
                val deletionRequestedAt by userViewModel.deletionRequestedAt.collectAsState()
                val isOnline by viewModel.isOnline.collectAsState()
                val pendingSyncCount by viewModel.pendingSyncCount.collectAsState()

                LaunchedEffect(taskGrouping) {
                    viewModel.setTaskGrouping(taskGrouping)
                }

                TaskListScreen(
                    taskGroups = taskGroups,
                    expandedGroups = expandedGroups,
                    isRefreshing = isRefreshing,
                    onRefresh = { viewModel.refreshTasks() },
                    onCompleteTask = { viewModel.completeTask(it) },
                    onSkipTask = { viewModel.skipTask(it) },
                    onDeleteTask = { viewModel.deleteTask(it) },
                    onCompleteAndEndRecurrenceTask = { viewModel.completeTask(it, endRecurrence = true) },
                    onTaskClick = { navController.navigate(Routes.taskFormEdit(it)) },
                    onViewHistory = { navController.navigate(Routes.taskHistory(it)) },
                    onCreateTask = { navController.navigate(Routes.TASK_FORM_CREATE) },
                    onToggleGroup = { viewModel.toggleGroupExpanded(it) },
                    swipeSettings = swipeSettings,
                    inlineCompleteEnabled = inlineCompleteEnabled,
                    isPendingDeletion = deletionRequestedAt != null,
                    isOnline = isOnline,
                    pendingSyncCount = pendingSyncCount,
                )
            }

            composable(Screen.Labels.route) {
                val viewModel: LabelViewModel = hiltViewModel()
                val labels by viewModel.labels.collectAsState()
                val isRefreshing by viewModel.isRefreshing.collectAsState()

                LabelsScreen(
                    labels = labels,
                    isRefreshing = isRefreshing,
                    onRefresh = { viewModel.refreshLabels() },
                    onCreateLabel = { name, color -> viewModel.createLabel(name, color) },
                    onUpdateLabel = { id, name, color -> viewModel.updateLabel(id, name, color) },
                    onDeleteLabel = { viewModel.deleteLabel(it) }
                )
            }

            composable(Screen.Settings.route) {
                val authViewModel: AuthViewModel = hiltViewModel()
                val userViewModel: UserViewModel = hiltViewModel()
                SettingsScreen(
                    authViewModel = authViewModel,
                    userViewModel = userViewModel,
                    themeMode = themeMode,
                    onThemeModeChanged = onThemeModeChanged,
                    taskGrouping = taskGrouping,
                    onTaskGroupingChanged = onTaskGroupingChanged,
                    calendarSyncEnabled = calendarSyncEnabled,
                    onCalendarSyncChanged = onCalendarSyncChanged,
                    calendarRepository = calendarRepository,
                    swipeSettings = swipeSettings,
                    onSwipeEnabledChanged = onSwipeEnabledChanged,
                    onSwipeDeleteConfirmationChanged = onSwipeDeleteConfirmationChanged,
                    onNavigateToSwipeSettings = { navController.navigate(Routes.SWIPE_SETTINGS) },
                    inlineCompleteEnabled = inlineCompleteEnabled,
                    onInlineCompleteEnabledChanged = onInlineCompleteEnabledChanged,
                    telemetryEnabled = telemetryEnabled,
                    onTelemetryEnabledChanged = onTelemetryEnabledChanged,
                    debugLoggingEnabled = debugLoggingEnabled,
                    onDebugLoggingEnabledChanged = onDebugLoggingEnabledChanged
                )
            }

            composable(
                route = Routes.SWIPE_SETTINGS,
                enterTransition = { slideInHorizontally(initialOffsetX = { it }) + fadeIn() },
                exitTransition = { fadeOut() },
                popEnterTransition = { fadeIn() },
                popExitTransition = { slideOutHorizontally(targetOffsetX = { it }) + fadeOut() }
            ) {
                SwipeActionsSettingsScreen(
                    swipeSettings = swipeSettings,
                    onStartToEndActionChanged = onSwipeStartToEndActionChanged,
                    onEndToStartActionChanged = onSwipeEndToStartActionChanged,
                    onBack = { navController.popBackStack() }
                )
            }

            composable(
                route = Routes.TASK_FORM,
                arguments = listOf(navArgument("taskId") {
                    type = NavType.IntType
                    defaultValue = -1
                }),
                enterTransition = { slideInHorizontally(initialOffsetX = { it }) + fadeIn() },
                exitTransition = { fadeOut() },
                popEnterTransition = { fadeIn() },
                popExitTransition = { slideOutHorizontally(targetOffsetX = { it }) + fadeOut() }
            ) { backStackEntry ->
                val taskId = backStackEntry.arguments?.getInt("taskId") ?: -1
                val viewModel: TaskFormViewModel = hiltViewModel()
                val availableLabels by viewModel.availableLabels.collectAsState()
                val isSaving by viewModel.isSaving.collectAsState()
                val saveResult by viewModel.saveResult.collectAsState()

                var existingTask by remember { mutableStateOf<Task?>(null) }

                LaunchedEffect(taskId) {
                    if (taskId > 0) {
                        viewModel.loadTask(taskId) { existingTask = it }
                    }
                }

                LaunchedEffect(saveResult) {
                    if (saveResult?.isSuccess == true) {
                        viewModel.clearSaveResult()
                        navController.popBackStack()
                    }
                }

                TaskFormScreen(
                    existingTask = existingTask,
                    availableLabels = availableLabels,
                    isSaving = isSaving,
                    onSave = { title, nextDueDate, endDate, frequency, notification, labelIds, isRolling ->
                        if (taskId > 0) {
                            viewModel.updateTask(UpdateTaskReq(
                                id = taskId,
                                title = title,
                                nextDueDate = nextDueDate,
                                endDate = endDate,
                                frequency = frequency,
                                notification = notification,
                                labels = labelIds,
                                isRolling = isRolling
                            ))
                        } else {
                            viewModel.createTask(CreateTaskReq(
                                title = title,
                                nextDueDate = nextDueDate,
                                endDate = endDate,
                                frequency = frequency,
                                notification = notification,
                                labels = labelIds,
                                isRolling = isRolling
                            ))
                        }
                    },
                    onBack = { navController.popBackStack() }
                )
            }

            composable(
                route = Routes.TASK_HISTORY,
                arguments = listOf(navArgument("taskId") {
                    type = NavType.IntType
                    defaultValue = -1
                }),
                enterTransition = { slideInHorizontally(initialOffsetX = { it }) + fadeIn() },
                exitTransition = { fadeOut() },
                popEnterTransition = { fadeIn() },
                popExitTransition = { slideOutHorizontally(targetOffsetX = { it }) + fadeOut() }
            ) { backStackEntry ->
                val taskId = backStackEntry.arguments?.getInt("taskId") ?: -1
                val viewModel: TaskHistoryViewModel = hiltViewModel()
                val history by viewModel.history.collectAsState()
                val isLoading by viewModel.isLoading.collectAsState()

                LaunchedEffect(taskId) {
                    if (taskId > 0) {
                        viewModel.loadHistory(taskId)
                    } else {
                        navController.popBackStack()
                    }
                }

                TaskHistoryScreen(
                    history = history,
                    isLoading = isLoading,
                    onBack = { navController.popBackStack() }
                )
            }
        }
    }
}
