package com.dkhalife.tasks.repo

import com.dkhalife.tasks.api.TaskWizardApi
import com.dkhalife.tasks.model.*
import com.dkhalife.tasks.telemetry.TelemetryManager
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class UserRepository @Inject constructor(
    private val api: TaskWizardApi,
    private val telemetryManager: TelemetryManager
) {
    suspend fun getUserProfile(): Result<UserProfile> {
        return try {
            val response = api.getUserProfile()
            if (response.isSuccessful) {
                Result.success(response.body()!!.user)
            } else {
                telemetryManager.logError(TAG, "Failed to fetch profile: ${response.code()}")
                Result.failure(Exception("Failed to fetch profile: ${response.code()}"))
            }
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to fetch profile: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun updateNotificationSettings(req: NotificationUpdateRequest): Result<Unit> {
        return try {
            val response = api.updateNotificationSettings(req)
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                telemetryManager.logError(TAG, "Failed to update notification settings: ${response.code()}")
                Result.failure(Exception("Failed to update notification settings: ${response.code()}"))
            }
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to update notification settings: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun requestDeletion(): Result<Unit> {
        return try {
            val response = api.requestAccountDeletion()
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                telemetryManager.logError(TAG, "Failed to request account deletion: ${response.code()}")
                Result.failure(Exception("Failed to request account deletion: ${response.code()}"))
            }
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to request account deletion: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun cancelDeletion(): Result<Unit> {
        return try {
            val response = api.cancelAccountDeletion()
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                telemetryManager.logError(TAG, "Failed to cancel account deletion: ${response.code()}")
                Result.failure(Exception("Failed to cancel account deletion: ${response.code()}"))
            }
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to cancel account deletion: ${e.message}", e)
            Result.failure(e)
        }
    }

    companion object {
        private const val TAG = "UserRepository"
    }
}
