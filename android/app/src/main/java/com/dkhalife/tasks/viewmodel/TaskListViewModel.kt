package com.dkhalife.tasks.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dkhalife.tasks.data.GroupingRepository
import com.dkhalife.tasks.data.TaskGroup
import com.dkhalife.tasks.data.TaskGrouper
import com.dkhalife.tasks.data.TaskGrouping
import com.dkhalife.tasks.model.Label
import com.dkhalife.tasks.model.Task
import com.dkhalife.tasks.repo.LabelRepository
import com.dkhalife.tasks.repo.TaskRepository
import com.dkhalife.tasks.utils.SoundEffect
import com.dkhalife.tasks.utils.SoundManager
import com.dkhalife.tasks.ws.WebSocketManager
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.flowOn
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class TaskListViewModel @Inject constructor(
    private val taskRepository: TaskRepository,
    private val labelRepository: LabelRepository,
    private val groupingRepository: GroupingRepository,
    private val webSocketManager: WebSocketManager,
    private val soundManager: SoundManager
) : ViewModel() {

    val tasks: StateFlow<List<Task>> = taskRepository.tasks
    val completedTasks: StateFlow<List<Task>> = taskRepository.completedTasks

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing: StateFlow<Boolean> = _isRefreshing

    private val _error = MutableStateFlow<String?>(null)
    val error: StateFlow<String?> = _error

    private val _taskGrouping = MutableStateFlow(groupingRepository.getTaskGrouping())
    val taskGrouping: StateFlow<TaskGrouping> = _taskGrouping

    private val _taskGroups = MutableStateFlow<List<TaskGroup>>(emptyList())
    val taskGroups: StateFlow<List<TaskGroup>> = _taskGroups

    private val _expandedGroups = MutableStateFlow(groupingRepository.getExpandedGroups())
    val expandedGroups: StateFlow<Set<String>> = _expandedGroups

    init {
        refreshTasks()
        viewModelScope.launch {
            labelRepository.refreshLabels().onFailure { _error.value = it.message }
        }
        webSocketManager.connect()
        collectWebSocketMessages()
        observeGrouping()
    }

    fun setTaskGrouping(grouping: TaskGrouping) {
        if (_taskGrouping.value == grouping) return
        _taskGrouping.value = grouping
        _expandedGroups.value = emptySet()
        groupingRepository.setExpandedGroups(emptySet())
    }

    fun toggleGroupExpanded(groupKey: String) {
        val current = _expandedGroups.value.toMutableSet()
        if (current.contains(groupKey)) {
            current.remove(groupKey)
        } else {
            current.add(groupKey)
        }
        _expandedGroups.value = current
        groupingRepository.setExpandedGroups(current)
    }

    private fun observeGrouping() {
        viewModelScope.launch {
            combine(tasks, _taskGrouping, labelRepository.labels) { tasks, grouping, labels ->
                computeGroups(tasks, grouping, labels)
            }.flowOn(Dispatchers.Default).collect { groups ->
                _taskGroups.value = groups
            }
        }
    }

    private fun computeGroups(
        tasks: List<Task>,
        grouping: TaskGrouping,
        labels: List<Label>
    ): List<TaskGroup> {
        return when (grouping) {
            TaskGrouping.DUE_DATE -> TaskGrouper.groupByDueDate(tasks)
            TaskGrouping.LABEL -> TaskGrouper.groupByLabel(tasks, labels)
        }
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
                .onSuccess {
                    soundManager.playSound(SoundEffect.TASK_COMPLETE)
                }
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
        soundManager.release()
    }
}
