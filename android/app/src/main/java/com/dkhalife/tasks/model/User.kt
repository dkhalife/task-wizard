package com.dkhalife.tasks.model

import com.google.gson.annotations.SerializedName

data class UserProfile(
    val notifications: NotificationSettings = NotificationSettings(),
    @SerializedName("deletion_requested_at")
    val deletionRequestedAt: String? = null
)

data class NotificationSettings(
    val provider: NotificationProvider = NotificationProvider(),
    val triggers: NotificationTriggerOptions = NotificationTriggerOptions()
)

data class NotificationProvider(
    val provider: String = "none",
    val url: String = "",
    val method: String = "",
    val token: String = ""
)

data class NotificationUpdateRequest(
    val provider: NotificationProvider,
    val triggers: NotificationTriggerOptions
)
