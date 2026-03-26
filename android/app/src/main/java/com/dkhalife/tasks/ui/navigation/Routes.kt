package com.dkhalife.tasks.ui.navigation

import androidx.annotation.StringRes
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Label
import androidx.compose.material.icons.filled.Checklist
import androidx.compose.material.icons.filled.Settings
import androidx.compose.ui.graphics.vector.ImageVector
import com.dkhalife.tasks.R

sealed class Screen(val route: String, @StringRes val titleRes: Int, val icon: ImageVector) {
    data object Tasks : Screen("tasks", R.string.nav_tasks, Icons.Default.Checklist)
    data object Labels : Screen("labels", R.string.nav_labels, Icons.AutoMirrored.Filled.Label)
    data object Settings : Screen("settings", R.string.nav_settings, Icons.Default.Settings)
}

object Routes {
    const val TASK_FORM = "task_form?taskId={taskId}"
    const val TASK_FORM_CREATE = "task_form"

    fun taskFormEdit(taskId: Int) = "task_form?taskId=$taskId"
}
