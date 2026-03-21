package com.dkhalife.tasks

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.runtime.*
import androidx.hilt.navigation.compose.hiltViewModel
import com.dkhalife.tasks.ui.navigation.AppNavigation
import com.dkhalife.tasks.ui.screen.SignInScreen
import com.dkhalife.tasks.ui.theme.TaskWizardTheme
import com.dkhalife.tasks.viewmodel.AuthViewModel
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class MainActivity : ComponentActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()

        setContent {
            TaskWizardTheme {
                val authViewModel: AuthViewModel = hiltViewModel()
                val isSignedIn by authViewModel.isSignedIn.collectAsState()
                val isLoading by authViewModel.isLoading.collectAsState()
                val errorMessage by authViewModel.errorMessage.collectAsState()
                val serverEndpoint by authViewModel.serverEndpoint.collectAsState()

                if (isSignedIn) {
                    AppNavigation()
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
