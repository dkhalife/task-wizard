package com.dkhalife.tasks.model

import com.google.gson.annotations.SerializedName

data class AuthConfig(
    val enabled: Boolean = false,
    @SerializedName("tenant_id") val tenantId: String? = null,
    @SerializedName("client_id") val clientId: String? = null,
    val audience: String? = null
)
