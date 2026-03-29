package com.dkhalife.tasks.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dkhalife.tasks.repo.UserRepository
import com.dkhalife.tasks.telemetry.TelemetryManager
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class UserViewModel @Inject constructor(
    private val userRepository: UserRepository,
    private val telemetryManager: TelemetryManager
) : ViewModel() {

    private val _deletionRequestedAt = MutableStateFlow<String?>(null)
    val deletionRequestedAt: StateFlow<String?> = _deletionRequestedAt

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading

    private val _errorMessage = MutableStateFlow<String?>(null)
    val errorMessage: StateFlow<String?> = _errorMessage

    init {
        loadProfile()
    }

    private fun loadProfile() {
        viewModelScope.launch {
            userRepository.getUserProfile()
                .onSuccess { profile ->
                    _deletionRequestedAt.value = profile.deletionRequestedAt
                }
                .onFailure { e ->
                    telemetryManager.logError(TAG, "Failed to load profile: ${e.message}", e)
                }
        }
    }

    fun requestDeletion() {
        viewModelScope.launch {
            _isLoading.value = true
            _errorMessage.value = null
            userRepository.requestDeletion()
                .onSuccess {
                    _deletionRequestedAt.value = java.time.Instant.now().toString()
                }
                .onFailure { e ->
                    telemetryManager.logError(TAG, "Failed to request deletion: ${e.message}", e)
                    _errorMessage.value = e.message
                }
            _isLoading.value = false
        }
    }

    fun cancelDeletion() {
        viewModelScope.launch {
            _isLoading.value = true
            _errorMessage.value = null
            userRepository.cancelDeletion()
                .onSuccess {
                    _deletionRequestedAt.value = null
                }
                .onFailure { e ->
                    telemetryManager.logError(TAG, "Failed to cancel deletion: ${e.message}", e)
                    _errorMessage.value = e.message
                }
            _isLoading.value = false
        }
    }

    fun clearError() {
        _errorMessage.value = null
    }

    companion object {
        private const val TAG = "UserViewModel"
    }
}
