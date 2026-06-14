package app.taskwiz

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.runtime.*
import app.taskwiz.data.GroupingRepository
import app.taskwiz.data.SwipeAction
import app.taskwiz.data.SwipeActionsRepository
import app.taskwiz.data.TaskGrouping
import app.taskwiz.data.TaskListSettingsRepository
import app.taskwiz.data.TelemetryRepository
import app.taskwiz.data.ThemeMode
import app.taskwiz.data.ThemeRepository
import app.taskwiz.data.calendar.CalendarRepository
import app.taskwiz.ui.navigation.AppNavigation
import app.taskwiz.ui.theme.TaskWizardTheme
import app.taskwiz.ui.widget.TaskListWidget
import app.taskwiz.ui.widget.quickadd.QuickAddWidget
import app.taskwiz.telemetry.TelemetryManager
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class MainActivity : ComponentActivity() {

    @Inject
    lateinit var themeRepository: ThemeRepository

    @Inject
    lateinit var groupingRepository: GroupingRepository

    @Inject
    lateinit var calendarRepository: CalendarRepository

    @Inject
    lateinit var swipeActionsRepository: SwipeActionsRepository

    @Inject
    lateinit var taskListSettingsRepository: TaskListSettingsRepository

    @Inject
    lateinit var telemetryRepository: TelemetryRepository

    @Inject
    lateinit var telemetryManager: TelemetryManager

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()

        val initialTaskId = intent?.getIntExtra(TaskListWidget.EXTRA_TASK_ID, -1) ?: -1
        val createTask = intent?.getBooleanExtra(QuickAddWidget.EXTRA_CREATE_TASK, false) ?: false

        setContent {
            var themeMode by remember { mutableStateOf(themeRepository.getThemeMode()) }
            var taskGrouping by remember { mutableStateOf(groupingRepository.getTaskGrouping()) }
            var calendarSyncEnabled by remember { mutableStateOf(calendarRepository.isCalendarSyncEnabled()) }
            var swipeSettings by remember { mutableStateOf(swipeActionsRepository.getSettings()) }
            var inlineCompleteEnabled by remember { mutableStateOf(taskListSettingsRepository.isInlineCompleteEnabled()) }
            var telemetryEnabled by remember { mutableStateOf(telemetryRepository.isTelemetryEnabled()) }
            var debugLoggingEnabled by remember { mutableStateOf(telemetryRepository.isDebugLoggingEnabled()) }

            TaskWizardTheme(themeMode = themeMode) {
                AppNavigation(
                    themeMode = themeMode,
                    onThemeModeChanged = { mode ->
                        themeRepository.setThemeMode(mode)
                        themeMode = mode
                    },
                    taskGrouping = taskGrouping,
                    onTaskGroupingChanged = { grouping ->
                        groupingRepository.setTaskGrouping(grouping)
                        taskGrouping = grouping
                    },
                    calendarSyncEnabled = calendarSyncEnabled,
                    onCalendarSyncChanged = { enabled ->
                        calendarSyncEnabled = enabled
                    },
                    calendarRepository = calendarRepository,
                    swipeSettings = swipeSettings,
                    onSwipeEnabledChanged = { enabled ->
                        swipeActionsRepository.setEnabled(enabled)
                        swipeSettings = swipeSettings.copy(enabled = enabled)
                    },
                    onSwipeStartToEndActionChanged = { action ->
                        swipeActionsRepository.setStartToEndAction(action)
                        swipeSettings = swipeSettings.copy(startToEndAction = action)
                    },
                    onSwipeEndToStartActionChanged = { action ->
                        swipeActionsRepository.setEndToStartAction(action)
                        swipeSettings = swipeSettings.copy(endToStartAction = action)
                    },
                    onSwipeDeleteConfirmationChanged = { enabled ->
                        swipeActionsRepository.setDeleteConfirmationEnabled(enabled)
                        swipeSettings = swipeSettings.copy(deleteConfirmationEnabled = enabled)
                    },
                    inlineCompleteEnabled = inlineCompleteEnabled,
                    onInlineCompleteEnabledChanged = { enabled ->
                        taskListSettingsRepository.setInlineCompleteEnabled(enabled)
                        inlineCompleteEnabled = enabled
                    },
                    telemetryEnabled = telemetryEnabled,
                    onTelemetryEnabledChanged = { enabled ->
                        telemetryRepository.setTelemetryEnabled(enabled)
                        telemetryEnabled = enabled
                        if (enabled) {
                            telemetryManager.initialize(this@MainActivity)
                        }
                    },
                    debugLoggingEnabled = debugLoggingEnabled,
                    onDebugLoggingEnabledChanged = { enabled ->
                        telemetryRepository.setDebugLoggingEnabled(enabled)
                        debugLoggingEnabled = enabled
                    },
                    initialTaskId = initialTaskId,
                    createTask = createTask
                )
            }
        }
    }
}
