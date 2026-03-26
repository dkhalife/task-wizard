package com.dkhalife.tasks.ui.navigation

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Label
import androidx.compose.material.icons.filled.Checklist
import androidx.compose.material.icons.filled.Settings
import androidx.compose.ui.graphics.vector.ImageVector

sealed class Screen(val route: String, val title: String, val icon: ImageVector) {
    data object Tasks : Screen("tasks", "Tasks", Icons.Default.Checklist)
    data object Labels : Screen("labels", "Labels", Icons.AutoMirrored.Filled.Label)
    data object Settings : Screen("settings", "Settings", Icons.Default.Settings)
}

object Routes {
    const val TASK_FORM = "task_form?taskId={taskId}"
    const val TASK_FORM_CREATE = "task_form"

    fun taskFormEdit(taskId: Int) = "task_form?taskId=$taskId"
}
