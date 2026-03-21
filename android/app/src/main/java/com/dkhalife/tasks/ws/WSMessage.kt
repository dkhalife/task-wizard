package com.dkhalife.tasks.ws

import com.google.gson.JsonElement

data class WSMessage(
    val requestId: String,
    val action: String,
    val data: Any? = null
)

data class WSResponse(
    val requestId: String,
    val action: String,
    val status: Int,
    val data: JsonElement? = null
)
