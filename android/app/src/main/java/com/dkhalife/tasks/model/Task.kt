package com.dkhalife.tasks.model

import com.google.gson.annotations.SerializedName

data class Task(
    val id: Int = 0,
    val title: String = "",
    @SerializedName("next_due_date") val nextDueDate: String? = null,
    @SerializedName("end_date") val endDate: String? = null,
    @SerializedName("is_rolling") val isRolling: Boolean = false,
    val frequency: Frequency = Frequency(),
    val notification: NotificationTriggerOptions = NotificationTriggerOptions(),
    val labels: List<Label> = emptyList(),
    @SerializedName("created_at") val createdAt: String? = null,
    @SerializedName("updated_at") val updatedAt: String? = null
)

data class CreateTaskReq(
    val title: String,
    @SerializedName("next_due_date") val nextDueDate: String? = null,
    @SerializedName("end_date") val endDate: String? = null,
    @SerializedName("is_rolling") val isRolling: Boolean = false,
    val frequency: Frequency = Frequency(),
    val notification: NotificationTriggerOptions = NotificationTriggerOptions(),
    val labels: List<Int> = emptyList()
)

data class UpdateTaskReq(
    val id: Int,
    val title: String,
    @SerializedName("next_due_date") val nextDueDate: String? = null,
    @SerializedName("end_date") val endDate: String? = null,
    @SerializedName("is_rolling") val isRolling: Boolean = false,
    val frequency: Frequency = Frequency(),
    val notification: NotificationTriggerOptions = NotificationTriggerOptions(),
    val labels: List<Int> = emptyList()
)

data class UpdateDueDateReq(
    @SerializedName("due_date") val dueDate: String
)

data class TaskHistory(
    val id: Int = 0,
    @SerializedName("task_id") val taskId: Int = 0,
    @SerializedName("completed_date") val completedDate: String? = null,
    @SerializedName("due_date") val dueDate: String? = null
)
