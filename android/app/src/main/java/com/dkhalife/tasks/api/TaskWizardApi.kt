package com.dkhalife.tasks.api

import com.dkhalife.tasks.model.*
import retrofit2.Response
import retrofit2.http.*

interface TaskWizardApi {

    @GET("api/v1/auth/config")
    suspend fun getAuthConfig(): Response<AuthConfig>

    @GET("api/v1/tasks/")
    suspend fun getTasks(): Response<TasksResponse>

    @GET("api/v1/tasks/completed")
    suspend fun getCompletedTasks(
        @Query("limit") limit: Int = 10,
        @Query("page") page: Int = 1
    ): Response<TasksResponse>

    @GET("api/v1/tasks/{id}")
    suspend fun getTask(@Path("id") id: Int): Response<TaskResponse>

    @GET("api/v1/tasks/{id}/history")
    suspend fun getTaskHistory(@Path("id") id: Int): Response<TaskHistoryResponse>

    @POST("api/v1/tasks/")
    suspend fun createTask(@Body req: CreateTaskReq): Response<TaskCreatedResponse>

    @PUT("api/v1/tasks/")
    suspend fun updateTask(@Body req: UpdateTaskReq): Response<Void>

    @DELETE("api/v1/tasks/{id}")
    suspend fun deleteTask(@Path("id") id: Int): Response<Void>

    @POST("api/v1/tasks/{id}/do")
    suspend fun completeTask(
        @Path("id") id: Int,
        @Query("endRecurrence") endRecurrence: Boolean = false
    ): Response<TaskResponse>

    @POST("api/v1/tasks/{id}/undo")
    suspend fun uncompleteTask(@Path("id") id: Int): Response<TaskResponse>

    @POST("api/v1/tasks/{id}/skip")
    suspend fun skipTask(@Path("id") id: Int): Response<TaskResponse>

    @PUT("api/v1/tasks/{id}/dueDate")
    suspend fun updateDueDate(
        @Path("id") id: Int,
        @Body req: UpdateDueDateReq
    ): Response<Void>

    @GET("api/v1/labels")
    suspend fun getLabels(): Response<LabelsResponse>

    @POST("api/v1/labels")
    suspend fun createLabel(@Body req: CreateLabelReq): Response<LabelCreatedResponse>

    @PUT("api/v1/labels")
    suspend fun updateLabel(@Body req: UpdateLabelReq): Response<Void>

    @DELETE("api/v1/labels/{id}")
    suspend fun deleteLabel(@Path("id") id: Int): Response<Void>

    @GET("api/v1/users/profile")
    suspend fun getUserProfile(): Response<UserResponse>

    @PUT("api/v1/users/notifications")
    suspend fun updateNotificationSettings(@Body req: NotificationUpdateRequest): Response<Void>

    @POST("api/v1/users/deletion")
    suspend fun requestAccountDeletion(): Response<Void>

    @DELETE("api/v1/users/deletion")
    suspend fun cancelAccountDeletion(): Response<Void>
}
