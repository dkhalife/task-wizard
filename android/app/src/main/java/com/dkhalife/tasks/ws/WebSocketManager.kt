package com.dkhalife.tasks.ws

import com.dkhalife.tasks.api.ApiEndpointProvider
import com.dkhalife.tasks.auth.AuthTokenProvider
import com.google.gson.Gson
import com.google.gson.JsonParser
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import okhttp3.*
import java.util.UUID
import java.util.concurrent.TimeUnit
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class WebSocketManager @Inject constructor(
    private val endpointProvider: ApiEndpointProvider,
    private val tokenProvider: AuthTokenProvider,
    private val gson: Gson
) {
    private var webSocket: WebSocket? = null
    private var reconnectAttempt = 0
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    private val _messages = MutableSharedFlow<WSResponse>(extraBufferCapacity = 64)
    val messages: SharedFlow<WSResponse> = _messages

    private val client = OkHttpClient.Builder()
        .pingInterval(54, TimeUnit.SECONDS)
        .readTimeout(60, TimeUnit.SECONDS)
        .build()

    fun connect() {
        scope.launch {
            val token = tokenProvider.getAccessToken() ?: return@launch

            val request = Request.Builder()
                .url(endpointProvider.getWebSocketUrl())
                .addHeader("Sec-WebSocket-Protocol", "tasks.websocket, $token")
                .build()

            webSocket = client.newWebSocket(request, createListener())
        }
    }

    fun disconnect() {
        webSocket?.close(1000, "Client disconnect")
        webSocket = null
    }

    fun sendAction(action: String, data: Any? = null): String {
        val requestId = UUID.randomUUID().toString()
        val message = WSMessage(
            requestId = requestId,
            action = action,
            data = data
        )
        webSocket?.send(gson.toJson(message))
        return requestId
    }

    private fun createListener() = object : WebSocketListener() {
        override fun onOpen(webSocket: WebSocket, response: Response) {
            reconnectAttempt = 0
        }

        override fun onMessage(webSocket: WebSocket, text: String) {
            try {
                val json = JsonParser.parseString(text).asJsonObject
                val response = WSResponse(
                    requestId = json.get("requestId")?.asString ?: "",
                    action = json.get("action")?.asString ?: "",
                    status = json.get("status")?.asInt ?: 0,
                    data = json.get("data")
                )
                _messages.tryEmit(response)
            } catch (_: Exception) { }
        }

        override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
            scheduleReconnect()
        }

        override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
            if (code != 1000) {
                scheduleReconnect()
            }
        }
    }

    private fun scheduleReconnect() {
        val delay = minOf(30_000L, 1000L * (1L shl reconnectAttempt))
        reconnectAttempt++

        scope.launch {
            delay(delay)
            connect()
        }
    }
}
