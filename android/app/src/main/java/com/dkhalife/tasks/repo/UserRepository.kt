package com.dkhalife.tasks.repo

import com.dkhalife.tasks.api.TaskWizardApi
import com.dkhalife.tasks.model.*
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class UserRepository @Inject constructor(
    private val api: TaskWizardApi
) {
    suspend fun getUserProfile(): Result<UserProfile> {
        return try {
            val response = api.getUserProfile()
            if (response.isSuccessful) {
                Result.success(response.body()!!.user)
            } else {
                Result.failure(Exception("Failed to fetch profile: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun updateNotificationSettings(req: NotificationUpdateRequest): Result<Unit> {
        return try {
            val response = api.updateNotificationSettings(req)
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to update notification settings: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
