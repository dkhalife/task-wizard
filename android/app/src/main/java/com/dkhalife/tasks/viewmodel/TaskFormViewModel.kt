package com.dkhalife.tasks.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dkhalife.tasks.model.*
import com.dkhalife.tasks.repo.LabelRepository
import com.dkhalife.tasks.repo.TaskRepository
import com.dkhalife.tasks.telemetry.TelemetryManager
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class TaskFormViewModel @Inject constructor(
    private val taskRepository: TaskRepository,
    private val labelRepository: LabelRepository,
    private val telemetryManager: TelemetryManager
) : ViewModel() {

    val availableLabels: StateFlow<List<Label>> = labelRepository.labels

    private val _isSaving = MutableStateFlow(false)
    val isSaving: StateFlow<Boolean> = _isSaving

    private val _saveResult = MutableStateFlow<Result<Unit>?>(null)
    val saveResult: StateFlow<Result<Unit>?> = _saveResult

    private val _error = MutableStateFlow<String?>(null)
    val error: StateFlow<String?> = _error

    fun loadTask(taskId: Int, onLoaded: (Task) -> Unit) {
        viewModelScope.launch {
            taskRepository.getTask(taskId)
                .onSuccess { onLoaded(it) }
                .onFailure {
                    telemetryManager.logError(TAG, "Failed to load task $taskId: ${it.message}", it)
                    _error.value = it.message
                }
        }
    }

    fun createTask(req: CreateTaskReq) {
        viewModelScope.launch {
            _isSaving.value = true
            taskRepository.createTask(req)
                .onSuccess { _saveResult.value = Result.success(Unit) }
                .onFailure {
                    telemetryManager.logError(TAG, "Failed to create task: ${it.message}", it)
                    _error.value = it.message
                }
            _isSaving.value = false
        }
    }

    fun updateTask(req: UpdateTaskReq) {
        viewModelScope.launch {
            _isSaving.value = true
            taskRepository.updateTask(req)
                .onSuccess { _saveResult.value = Result.success(Unit) }
                .onFailure {
                    telemetryManager.logError(TAG, "Failed to update task: ${it.message}", it)
                    _error.value = it.message
                }
            _isSaving.value = false
        }
    }

    fun clearSaveResult() {
        _saveResult.value = null
    }

    fun clearError() {
        _error.value = null
    }

    companion object {
        private const val TAG = "TaskFormViewModel"
    }
}
