package com.dkhalife.tasks.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dkhalife.tasks.model.ActivityEntry
import com.dkhalife.tasks.repo.RevertConflictException
import com.dkhalife.tasks.repo.TaskRepository
import com.dkhalife.tasks.telemetry.TelemetryManager
import com.dkhalife.tasks.ws.WebSocketManager
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class ActivityViewModel @Inject constructor(
    private val taskRepository: TaskRepository,
    private val webSocketManager: WebSocketManager,
    private val telemetryManager: TelemetryManager,
) : ViewModel() {

    private val _items = MutableStateFlow<List<ActivityEntry>>(emptyList())
    val items: StateFlow<List<ActivityEntry>> = _items

    private val _isLoading = MutableStateFlow(true)
    val isLoading: StateFlow<Boolean> = _isLoading

    private val _isLoadingMore = MutableStateFlow(false)
    val isLoadingMore: StateFlow<Boolean> = _isLoadingMore

    private val _hasMore = MutableStateFlow(false)
    val hasMore: StateFlow<Boolean> = _hasMore

    private val _isReverting = MutableStateFlow(false)
    val isReverting: StateFlow<Boolean> = _isReverting

    private val _message = MutableStateFlow<String?>(null)
    val message: StateFlow<String?> = _message

    private var generation = 0

    init {
        refresh()
        observeWebSocket()
    }

    fun refresh() {
        val gen = ++generation
        viewModelScope.launch {
            _isLoading.value = true
            taskRepository.getActivity(beforeId = 0, limit = PAGE_SIZE)
                .onSuccess {
                    if (gen != generation) return@onSuccess
                    _items.value = it
                    _hasMore.value = it.size >= PAGE_SIZE
                }
                .onFailure {
                    if (gen != generation) return@onFailure
                    telemetryManager.logError(TAG, "Failed to load activity: ${it.message}", it)
                    _message.value = it.message
                }
            if (gen == generation) {
                _isLoading.value = false
            }
        }
    }

    fun loadMore() {
        if (_isLoadingMore.value || !_hasMore.value) return
        val cursor = _items.value.lastOrNull()?.id ?: return
        val gen = generation
        viewModelScope.launch {
            _isLoadingMore.value = true
            taskRepository.getActivity(beforeId = cursor, limit = PAGE_SIZE)
                .onSuccess { page ->
                    if (gen != generation) return@onSuccess
                    val existingIds = _items.value.mapTo(HashSet()) { it.id }
                    _items.value = _items.value + page.filter { it.id !in existingIds }
                    _hasMore.value = page.size >= PAGE_SIZE
                }
                .onFailure {
                    telemetryManager.logError(TAG, "Failed to load more activity: ${it.message}", it)
                    _message.value = it.message
                }
            _isLoadingMore.value = false
        }
    }

    fun revert(taskId: Int, historyId: Int) {
        if (_isReverting.value) return
        viewModelScope.launch {
            _isReverting.value = true
            taskRepository.revertActivity(taskId, historyId)
                .onSuccess { refresh() }
                .onFailure {
                    if (it is RevertConflictException) {
                        _message.value = it.message
                        refresh()
                    } else {
                        telemetryManager.logError(TAG, "Failed to revert action: ${it.message}", it)
                        _message.value = it.message
                    }
                }
            _isReverting.value = false
        }
    }

    fun clearMessage() {
        _message.value = null
    }

    private fun observeWebSocket() {
        viewModelScope.launch {
            webSocketManager.messages.collect { message ->
                if (message.action in REFRESH_EVENTS) {
                    refresh()
                }
            }
        }
    }

    companion object {
        private const val TAG = "ActivityViewModel"
        private const val PAGE_SIZE = 20
        private val REFRESH_EVENTS = setOf(
            "task_completed",
            "task_uncompleted",
            "task_skipped",
            "task_deleted",
        )
    }
}
