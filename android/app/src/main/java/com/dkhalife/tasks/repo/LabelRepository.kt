package com.dkhalife.tasks.repo

import com.dkhalife.tasks.api.TaskWizardApi
import com.dkhalife.tasks.model.*
import com.dkhalife.tasks.telemetry.TelemetryManager
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class LabelRepository @Inject constructor(
    private val api: TaskWizardApi,
    private val telemetryManager: TelemetryManager
) {
    private val _labels = MutableStateFlow<List<Label>>(emptyList())
    val labels: StateFlow<List<Label>> = _labels

    suspend fun refreshLabels(): Result<List<Label>> {
        return try {
            val response = api.getLabels()
            if (response.isSuccessful) {
                val labels = response.body()?.labels ?: emptyList()
                _labels.value = labels
                Result.success(labels)
            } else {
                telemetryManager.logError(TAG, "Failed to fetch labels: ${response.code()}")
                Result.failure(Exception("Failed to fetch labels: ${response.code()}"))
            }
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to fetch labels: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun createLabel(req: CreateLabelReq): Result<Int> {
        return try {
            val response = api.createLabel(req)
            if (response.isSuccessful) {
                refreshLabels()
                Result.success(response.body()!!.label)
            } else {
                telemetryManager.logError(TAG, "Failed to create label: ${response.code()}")
                Result.failure(Exception("Failed to create label: ${response.code()}"))
            }
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to create label: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun updateLabel(req: UpdateLabelReq): Result<Unit> {
        return try {
            val response = api.updateLabel(req)
            if (response.isSuccessful) {
                refreshLabels()
                Result.success(Unit)
            } else {
                telemetryManager.logError(TAG, "Failed to update label: ${response.code()}")
                Result.failure(Exception("Failed to update label: ${response.code()}"))
            }
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to update label: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun deleteLabel(id: Int): Result<Unit> {
        return try {
            val response = api.deleteLabel(id)
            if (response.isSuccessful) {
                refreshLabels()
                Result.success(Unit)
            } else {
                telemetryManager.logError(TAG, "Failed to delete label: ${response.code()}")
                Result.failure(Exception("Failed to delete label: ${response.code()}"))
            }
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to delete label: ${e.message}", e)
            Result.failure(e)
        }
    }

    fun updateLabelsFromWebSocket(labels: List<Label>) {
        _labels.value = labels
    }

    companion object {
        private const val TAG = "LabelRepository"
    }
}
