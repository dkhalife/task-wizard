package com.dkhalife.tasks

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.runtime.*
import androidx.hilt.navigation.compose.hiltViewModel
import com.dkhalife.tasks.data.GroupingRepository
import com.dkhalife.tasks.data.TaskGrouping
import com.dkhalife.tasks.data.ThemeMode
import com.dkhalife.tasks.data.ThemeRepository
import com.dkhalife.tasks.ui.navigation.AppNavigation
import com.dkhalife.tasks.ui.screen.SignInScreen
import com.dkhalife.tasks.ui.theme.TaskWizardTheme
import com.dkhalife.tasks.viewmodel.AuthViewModel
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class MainActivity : ComponentActivity() {

    @Inject
    lateinit var themeRepository: ThemeRepository

    @Inject
    lateinit var groupingRepository: GroupingRepository

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()

        setContent {
            var themeMode by remember { mutableStateOf(themeRepository.getThemeMode()) }
            var taskGrouping by remember { mutableStateOf(groupingRepository.getTaskGrouping()) }

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
                        }
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
