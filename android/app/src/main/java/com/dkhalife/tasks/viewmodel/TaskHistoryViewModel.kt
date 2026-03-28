package com.dkhalife.tasks.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dkhalife.tasks.model.TaskHistory
import com.dkhalife.tasks.repo.TaskRepository
import com.dkhalife.tasks.telemetry.TelemetryManager
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class TaskHistoryViewModel @Inject constructor(
    private val taskRepository: TaskRepository,
    private val telemetryManager: TelemetryManager
) : ViewModel() {

    private val _history = MutableStateFlow<List<TaskHistory>>(emptyList())
    val history: StateFlow<List<TaskHistory>> = _history

    private val _isLoading = MutableStateFlow(true)
    val isLoading: StateFlow<Boolean> = _isLoading

    fun loadHistory(taskId: Int) {
        viewModelScope.launch {
            _history.value = emptyList()
            _isLoading.value = true
            taskRepository.getTaskHistory(taskId)
                .onSuccess { _history.value = it }
                .onFailure {
                    telemetryManager.logError(TAG, "Failed to load task history: ${it.message}", it)
                }
            _isLoading.value = false
        }
    }

    companion object {
        private const val TAG = "TaskHistoryViewModel"
    }
}
