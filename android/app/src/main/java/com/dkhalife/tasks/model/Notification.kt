package com.dkhalife.tasks.model

import com.google.gson.annotations.SerializedName

data class NotificationTriggerOptions(
    val enabled: Boolean = false,
    @SerializedName("due_date") val dueDate: Boolean = false,
    @SerializedName("pre_due") val preDue: Boolean = false,
    val overdue: Boolean = false
)
