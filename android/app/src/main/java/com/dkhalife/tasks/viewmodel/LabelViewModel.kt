package com.dkhalife.tasks.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dkhalife.tasks.model.*
import com.dkhalife.tasks.repo.LabelRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class LabelViewModel @Inject constructor(
    private val labelRepository: LabelRepository
) : ViewModel() {

    val labels: StateFlow<List<Label>> = labelRepository.labels

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing: StateFlow<Boolean> = _isRefreshing

    private val _error = MutableStateFlow<String?>(null)
    val error: StateFlow<String?> = _error

    init {
        refreshLabels()
    }

    fun refreshLabels() {
        viewModelScope.launch {
            _isRefreshing.value = true
            labelRepository.refreshLabels()
                .onFailure { _error.value = it.message }
            _isRefreshing.value = false
        }
    }

    fun createLabel(name: String, color: String) {
        viewModelScope.launch {
            labelRepository.createLabel(CreateLabelReq(name, color))
                .onFailure { _error.value = it.message }
        }
    }

    fun updateLabel(id: Int, name: String, color: String) {
        viewModelScope.launch {
            labelRepository.updateLabel(UpdateLabelReq(id, name, color))
                .onFailure { _error.value = it.message }
        }
    }

    fun deleteLabel(id: Int) {
        viewModelScope.launch {
            labelRepository.deleteLabel(id)
                .onFailure { _error.value = it.message }
        }
    }

    fun clearError() {
        _error.value = null
    }
}
