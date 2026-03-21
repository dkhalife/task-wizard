package com.dkhalife.tasks.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dkhalife.tasks.model.Task
import com.dkhalife.tasks.repo.TaskRepository
import com.dkhalife.tasks.ws.WebSocketManager
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class TaskListViewModel @Inject constructor(
    private val taskRepository: TaskRepository,
    private val webSocketManager: WebSocketManager
) : ViewModel() {

    val tasks: StateFlow<List<Task>> = taskRepository.tasks
    val completedTasks: StateFlow<List<Task>> = taskRepository.completedTasks

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing: StateFlow<Boolean> = _isRefreshing

    private val _error = MutableStateFlow<String?>(null)
    val error: StateFlow<String?> = _error

    init {
        refreshTasks()
        webSocketManager.connect()
        collectWebSocketMessages()
    }

    private fun collectWebSocketMessages() {
        viewModelScope.launch {
            webSocketManager.messages.collect { message ->
                when (message.action) {
                    "task_created", "task_updated", "task_skipped" -> refreshTasks()
                    "task_completed" -> {
                        refreshTasks()
                        refreshCompletedTasks()
                    }
                    "task_uncompleted" -> {
                        refreshTasks()
                        refreshCompletedTasks()
                    }
                    "task_deleted" -> {
                        refreshTasks()
                        refreshCompletedTasks()
                    }
                }
            }
        }
    }

    fun refreshTasks() {
        viewModelScope.launch {
            _isRefreshing.value = true
            taskRepository.refreshTasks().onFailure { _error.value = it.message }
            _isRefreshing.value = false
        }
    }

    fun refreshCompletedTasks(limit: Int = 10, page: Int = 1) {
        viewModelScope.launch {
            taskRepository.refreshCompletedTasks(limit, page)
                .onFailure { _error.value = it.message }
        }
    }

    fun completeTask(id: Int, endRecurrence: Boolean = false) {
        viewModelScope.launch {
            taskRepository.completeTask(id, endRecurrence)
                .onFailure { _error.value = it.message }
        }
    }

    fun uncompleteTask(id: Int) {
        viewModelScope.launch {
            taskRepository.uncompleteTask(id)
                .onFailure { _error.value = it.message }
        }
    }

    fun skipTask(id: Int) {
        viewModelScope.launch {
            taskRepository.skipTask(id)
                .onFailure { _error.value = it.message }
        }
    }

    fun deleteTask(id: Int) {
        viewModelScope.launch {
            taskRepository.deleteTask(id)
                .onFailure { _error.value = it.message }
        }
    }

    fun clearError() {
        _error.value = null
    }

    override fun onCleared() {
        super.onCleared()
        webSocketManager.disconnect()
    }
}
