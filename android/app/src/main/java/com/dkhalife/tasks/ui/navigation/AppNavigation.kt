package com.dkhalife.tasks.ui.navigation

import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Checklist
import androidx.compose.material.icons.filled.Label
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.dkhalife.tasks.model.*
import com.dkhalife.tasks.ui.screen.*
import com.dkhalife.tasks.viewmodel.*

sealed class Screen(val route: String, val title: String, val icon: ImageVector) {
    data object Tasks : Screen("tasks", "Tasks", Icons.Default.Checklist)
    data object Labels : Screen("labels", "Labels", Icons.Default.Label)
    data object Settings : Screen("settings", "Settings", Icons.Default.Settings)
}

object Routes {
    const val TASK_FORM = "task_form?taskId={taskId}"
    const val TASK_FORM_CREATE = "task_form"

    fun taskFormEdit(taskId: Int) = "task_form?taskId=$taskId"
}

@Composable
fun AppNavigation() {
    val navController = rememberNavController()
    val bottomScreens = listOf(Screen.Tasks, Screen.Labels, Screen.Settings)
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route

    val showBottomBar = bottomScreens.any { it.route == currentRoute }

    Scaffold(
        bottomBar = {
            if (showBottomBar) {
                NavigationBar {
                    bottomScreens.forEach { screen ->
                        NavigationBarItem(
                            icon = { Icon(screen.icon, contentDescription = screen.title) },
                            label = { Text(screen.title) },
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
            modifier = Modifier.padding(innerPadding)
        ) {
            composable(Screen.Tasks.route) {
                val viewModel: TaskListViewModel = hiltViewModel()
                val tasks by viewModel.tasks.collectAsState()
                val isRefreshing by viewModel.isRefreshing.collectAsState()

                TaskListScreen(
                    tasks = tasks,
                    isRefreshing = isRefreshing,
                    onRefresh = { viewModel.refreshTasks() },
                    onCompleteTask = { viewModel.completeTask(it) },
                    onSkipTask = { viewModel.skipTask(it) },
                    onDeleteTask = { viewModel.deleteTask(it) },
                    onTaskClick = { navController.navigate(Routes.taskFormEdit(it)) },
                    onCreateTask = { navController.navigate(Routes.TASK_FORM_CREATE) }
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
                SettingsScreen(authViewModel = authViewModel)
            }

            composable(
                route = Routes.TASK_FORM,
                arguments = listOf(navArgument("taskId") {
                    type = NavType.IntType
                    defaultValue = -1
                })
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
                    onSave = { title, nextDueDate, frequency, labelIds, isRolling ->
                        if (taskId > 0) {
                            viewModel.updateTask(UpdateTaskReq(
                                id = taskId,
                                title = title,
                                nextDueDate = nextDueDate,
                                frequency = frequency,
                                labels = labelIds,
                                isRolling = isRolling
                            ))
                        } else {
                            viewModel.createTask(CreateTaskReq(
                                title = title,
                                nextDueDate = nextDueDate,
                                frequency = frequency,
                                labels = labelIds,
                                isRolling = isRolling
                            ))
                        }
                    },
                    onBack = { navController.popBackStack() }
                )
            }
        }
    }
}
