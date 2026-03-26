package com.dkhalife.tasks.ui.screen

import android.Manifest
import android.content.pm.PackageManager
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import androidx.work.WorkManager
import com.dkhalife.tasks.data.TaskGrouping
import com.dkhalife.tasks.data.ThemeMode
import com.dkhalife.tasks.data.calendar.CalendarRepository
import com.dkhalife.tasks.viewmodel.AuthViewModel
import kotlinx.coroutines.launch

private val CALENDAR_PERMISSIONS = arrayOf(
    Manifest.permission.READ_CALENDAR,
    Manifest.permission.WRITE_CALENDAR
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsScreen(
    authViewModel: AuthViewModel,
    themeMode: ThemeMode,
    onThemeModeChanged: (ThemeMode) -> Unit,
    taskGrouping: TaskGrouping,
    onTaskGroupingChanged: (TaskGrouping) -> Unit,
    calendarSyncEnabled: Boolean,
    onCalendarSyncChanged: (Boolean) -> Unit,
    calendarRepository: CalendarRepository
) {
    val serverEndpoint by authViewModel.serverEndpoint.collectAsState()
    val context = LocalContext.current
    val contentResolver = context.contentResolver
    val workManager = remember { WorkManager.getInstance(context) }
    val scope = rememberCoroutineScope()

    var errorMessage by remember { mutableStateOf<String?>(null) }

    val calendarPermissionLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestMultiplePermissions()
    ) { permissions ->
        val allGranted = permissions.values.all { it }
        if (allGranted) {
            scope.launch {
                val result = calendarRepository.enableCalendarSync(contentResolver, workManager)
                if (result.isSuccess) {
                    onCalendarSyncChanged(true)
                } else {
                    errorMessage = "Failed to enable calendar sync: ${result.exceptionOrNull()?.message}"
                }
            }
        }
    }

    if (errorMessage != null) {
        AlertDialog(
            onDismissRequest = { errorMessage = null },
            title = { Text("Error") },
            text = { Text(errorMessage ?: "") },
            confirmButton = {
                TextButton(onClick = { errorMessage = null }) {
                    Text("OK")
                }
            }
        )
    }

    Scaffold(
        topBar = {
            TopAppBar(title = { Text("Settings") })
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            Card(modifier = Modifier.fillMaxWidth()) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Server", style = MaterialTheme.typography.titleMedium)
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        text = serverEndpoint,
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }

            Card(modifier = Modifier.fillMaxWidth()) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Theme", style = MaterialTheme.typography.titleMedium)
                    Spacer(modifier = Modifier.height(8.dp))
                    SingleChoiceSegmentedButtonRow(modifier = Modifier.fillMaxWidth()) {
                        ThemeMode.entries.forEachIndexed { index, mode ->
                            SegmentedButton(
                                selected = themeMode == mode,
                                onClick = { onThemeModeChanged(mode) },
                                shape = SegmentedButtonDefaults.itemShape(
                                    index = index,
                                    count = ThemeMode.entries.size
                                )
                            ) {
                                Text(mode.name.lowercase().replaceFirstChar { it.uppercase() })
                            }
                        }
                    }
                }
            }

            Card(modifier = Modifier.fillMaxWidth()) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Task grouping", style = MaterialTheme.typography.titleMedium)
                    Spacer(modifier = Modifier.height(8.dp))
                    SingleChoiceSegmentedButtonRow(modifier = Modifier.fillMaxWidth()) {
                        TaskGrouping.entries.forEachIndexed { index, grouping ->
                            SegmentedButton(
                                selected = taskGrouping == grouping,
                                onClick = { onTaskGroupingChanged(grouping) },
                                shape = SegmentedButtonDefaults.itemShape(
                                    index = index,
                                    count = TaskGrouping.entries.size
                                )
                            ) {
                                Text(
                                    when (grouping) {
                                        TaskGrouping.DUE_DATE -> "Due date"
                                        TaskGrouping.LABEL -> "Label"
                                    }
                                )
                            }
                        }
                    }
                }
            }

            Card(modifier = Modifier.fillMaxWidth()) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text("Calendar sync", style = MaterialTheme.typography.titleMedium)
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(
                            text = "Show tasks in your device calendar",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                    Switch(
                        checked = calendarSyncEnabled,
                        onCheckedChange = { enabled ->
                            if (enabled) {
                                val hasPermissions = CALENDAR_PERMISSIONS.all {
                                    ContextCompat.checkSelfPermission(context, it) == PackageManager.PERMISSION_GRANTED
                                }
                                if (hasPermissions) {
                                    scope.launch {
                                        val result = calendarRepository.enableCalendarSync(contentResolver, workManager)
                                        if (result.isSuccess) {
                                            onCalendarSyncChanged(true)
                                        } else {
                                            errorMessage = "Failed to enable calendar sync: ${result.exceptionOrNull()?.message}"
                                        }
                                    }
                                } else {
                                    calendarPermissionLauncher.launch(CALENDAR_PERMISSIONS)
                                }
                            } else {
                                scope.launch {
                                    val result = calendarRepository.disableCalendarSync(contentResolver, workManager)
                                    if (result.isSuccess) {
                                        onCalendarSyncChanged(false)
                                    } else {
                                        errorMessage = "Failed to disable calendar sync: ${result.exceptionOrNull()?.message}"
                                    }
                                }
                            }
                        }
                    )
                }
            }

            Spacer(modifier = Modifier.weight(1f))

            OutlinedButton(
                onClick = { authViewModel.signOut() },
                modifier = Modifier.fillMaxWidth(),
                colors = ButtonDefaults.outlinedButtonColors(
                    contentColor = MaterialTheme.colorScheme.error
                )
            ) {
                Text("Sign Out")
            }
        }
    }
}
