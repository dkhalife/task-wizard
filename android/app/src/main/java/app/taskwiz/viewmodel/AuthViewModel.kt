package app.taskwiz.viewmodel

import android.app.Activity
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import app.taskwiz.api.ApiEndpointProvider
import app.taskwiz.api.TaskWizardApi
import app.taskwiz.auth.AuthManager
import app.taskwiz.data.sync.SyncCoordinator
import app.taskwiz.telemetry.TelemetryManager
import com.microsoft.identity.client.AuthenticationCallback
import com.microsoft.identity.client.IAuthenticationResult
import com.microsoft.identity.client.ISingleAccountPublicClientApplication
import com.microsoft.identity.client.exception.MsalException
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class AuthViewModel @Inject constructor(
    private val authManager: AuthManager,
    private val endpointProvider: ApiEndpointProvider,
    private val api: TaskWizardApi,
    private val syncCoordinator: SyncCoordinator,
    private val telemetryManager: TelemetryManager
) : ViewModel() {

    private val _isSignedIn = MutableStateFlow(authManager.isSignedIn())
    val isSignedIn: StateFlow<Boolean> = _isSignedIn

    private val _isLoading = MutableStateFlow(!authManager.isLoaded())
    val isLoading: StateFlow<Boolean> = _isLoading

    private val _errorMessage = MutableStateFlow<String?>(null)
    val errorMessage: StateFlow<String?> = _errorMessage

    private val _serverEndpoint = MutableStateFlow(endpointProvider.getServerEndpoint())
    val serverEndpoint: StateFlow<String> = _serverEndpoint

    val sessionExpired: StateFlow<Boolean> = authManager.sessionExpired

    private val userChangeListener = AuthManager.UserChangeListener { isSignedIn ->
        _isSignedIn.value = isSignedIn
        _isLoading.value = false
    }

    init {
        authManager.addUserChangeListener(userChangeListener)
    }

    fun signIn(activity: Activity) {
        _isLoading.value = true
        _errorMessage.value = null

        if (authManager.getSingleAccountApp() != null) {
            performSignIn(activity)
            return
        }

        viewModelScope.launch(Dispatchers.IO) {
            try {
                val response = api.getAuthConfig()
                if (!response.isSuccessful) {
                    _isLoading.value = false
                    _errorMessage.value = "Failed to fetch auth config: HTTP ${response.code()}"
                    return@launch
                }

                val authConfig = response.body()
                if (authConfig == null || !authConfig.enabled) {
                    _isLoading.value = false
                    _errorMessage.value = "This server does not support authentication"
                    return@launch
                }

                authManager.initializeMsal(activity.applicationContext, authConfig, object : AuthManager.MsalInitCallback {
                    override fun onReady(app: ISingleAccountPublicClientApplication) {
                        performSignIn(activity)
                    }

                    override fun onError(exception: Exception) {
                        _isLoading.value = false
                        _errorMessage.value = "Failed to initialize authentication: ${exception.message}"
                    }
                })
            } catch (e: Exception) {
                telemetryManager.logError(TAG, "Failed to fetch auth config: ${e.message}", e)
                _isLoading.value = false
                _errorMessage.value = "Could not connect to server: ${e.message}"
            }
        }
    }

    private fun performSignIn(activity: Activity) {
        authManager.signIn(activity, object : AuthenticationCallback {
            override fun onSuccess(authenticationResult: IAuthenticationResult) {
                _isLoading.value = false
                syncCoordinator.syncOnce()
            }

            override fun onError(exception: MsalException) {
                telemetryManager.logError(TAG, "Sign-in failed: ${exception.message}", exception)
                _isLoading.value = false
                _errorMessage.value = exception.message
            }

            override fun onCancel() {
                _isLoading.value = false
            }
        })
    }

    fun signOut() {
        authManager.signOut(object : ISingleAccountPublicClientApplication.SignOutCallback {
            override fun onSignOut() {
                _errorMessage.value = null
            }

            override fun onError(exception: MsalException) {
                telemetryManager.logError(TAG, "Sign-out failed: ${exception.message}", exception)
                _errorMessage.value = exception.message
            }
        })
    }

    fun updateServerEndpoint(endpoint: String) {
        endpointProvider.setServerEndpoint(endpoint)
        _serverEndpoint.value = endpoint
        authManager.invalidateMsalApp()
    }

    fun clearError() {
        _errorMessage.value = null
    }

    override fun onCleared() {
        super.onCleared()
        authManager.removeUserChangeListener(userChangeListener)
    }

    companion object {
        private const val TAG = "AuthViewModel"
    }
}
