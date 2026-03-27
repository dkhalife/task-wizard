package com.dkhalife.tasks

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.runtime.*
import androidx.hilt.navigation.compose.hiltViewModel
import com.dkhalife.tasks.data.GroupingRepository
import com.dkhalife.tasks.data.SwipeAction
import com.dkhalife.tasks.data.SwipeActionsRepository
import com.dkhalife.tasks.data.TaskGrouping
import com.dkhalife.tasks.data.TelemetryRepository
import com.dkhalife.tasks.data.ThemeMode
import com.dkhalife.tasks.data.ThemeRepository
import com.dkhalife.tasks.data.calendar.CalendarRepository
import com.dkhalife.tasks.ui.navigation.AppNavigation
import com.dkhalife.tasks.ui.screen.SignInScreen
import com.dkhalife.tasks.ui.theme.TaskWizardTheme
import com.dkhalife.tasks.ui.widget.TaskListWidget
import com.dkhalife.tasks.ui.widget.quickadd.QuickAddWidget
import com.dkhalife.tasks.viewmodel.AuthViewModel
import com.dkhalife.tasks.telemetry.TelemetryManager
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
            var telemetryEnabled by remember { mutableStateOf(telemetryRepository.isTelemetryEnabled()) }
            var debugLoggingEnabled by remember { mutableStateOf(telemetryRepository.isDebugLoggingEnabled()) }

            TaskWizardTheme(themeMode = themeMode) {
                val authViewModel: AuthViewModel = hiltViewModel()
                val isSignedIn by authViewModel.isSignedIn.collectAsState()
                val isLoading by authViewModel.isLoading.collectAsState()
                val errorMessage by authViewModel.errorMessage.collectAsState()
                val serverEndpoint by authViewModel.serverEndpoint.collectAsState()

                if (isSignedIn) {
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
                } else {
                    SignInScreen(
                        serverEndpoint = serverEndpoint,
                        isLoading = isLoading,
                        errorMessage = errorMessage,
                        onSignIn = { activity -> authViewModel.signIn(activity) },
                        onEndpointChanged = { authViewModel.updateServerEndpoint(it) }
                    )
                }
            }
        }
    }
}
