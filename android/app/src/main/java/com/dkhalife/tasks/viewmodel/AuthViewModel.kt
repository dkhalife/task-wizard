package com.dkhalife.tasks.viewmodel

import android.app.Activity
import androidx.lifecycle.ViewModel
import com.dkhalife.tasks.api.ApiEndpointProvider
import com.dkhalife.tasks.auth.AuthManager
import com.microsoft.identity.client.AuthenticationCallback
import com.microsoft.identity.client.IAuthenticationResult
import com.microsoft.identity.client.ISingleAccountPublicClientApplication
import com.microsoft.identity.client.exception.MsalException
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import javax.inject.Inject

@HiltViewModel
class AuthViewModel @Inject constructor(
    private val authManager: AuthManager,
    private val endpointProvider: ApiEndpointProvider
) : ViewModel() {

    private val _isSignedIn = MutableStateFlow(authManager.isSignedIn())
    val isSignedIn: StateFlow<Boolean> = _isSignedIn

    private val _isLoading = MutableStateFlow(!authManager.isLoaded())
    val isLoading: StateFlow<Boolean> = _isLoading

    private val _errorMessage = MutableStateFlow<String?>(null)
    val errorMessage: StateFlow<String?> = _errorMessage

    private val _serverEndpoint = MutableStateFlow(endpointProvider.getServerEndpoint())
    val serverEndpoint: StateFlow<String> = _serverEndpoint

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

        authManager.signIn(activity, object : AuthenticationCallback {
            override fun onSuccess(authenticationResult: IAuthenticationResult) {
                _isLoading.value = false
            }

            override fun onError(exception: MsalException) {
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
                _errorMessage.value = exception.message
            }
        })
    }

    fun updateServerEndpoint(endpoint: String) {
        endpointProvider.setServerEndpoint(endpoint)
        _serverEndpoint.value = endpoint
    }

    fun clearError() {
        _errorMessage.value = null
    }

    override fun onCleared() {
        super.onCleared()
        authManager.removeUserChangeListener(userChangeListener)
    }
}
