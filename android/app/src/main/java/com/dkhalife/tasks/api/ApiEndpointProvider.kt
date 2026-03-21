package com.dkhalife.tasks.api

import android.content.SharedPreferences
import androidx.core.content.edit
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ApiEndpointProvider @Inject constructor(
    private val sharedPreferences: SharedPreferences
) {
    fun getBaseUrl(): String {
        val host = sharedPreferences.getString(KEY_SERVER_ENDPOINT, DEFAULT_HOST) ?: DEFAULT_HOST
        return "https://$host"
    }

    fun getWebSocketUrl(): String {
        val host = sharedPreferences.getString(KEY_SERVER_ENDPOINT, DEFAULT_HOST) ?: DEFAULT_HOST
        return "wss://$host/ws"
    }

    fun setServerEndpoint(host: String) {
        sharedPreferences.edit { putString(KEY_SERVER_ENDPOINT, host) }
    }

    fun getServerEndpoint(): String {
        return sharedPreferences.getString(KEY_SERVER_ENDPOINT, DEFAULT_HOST) ?: DEFAULT_HOST
    }

    companion object {
        private const val KEY_SERVER_ENDPOINT = "server_endpoint"
        private const val DEFAULT_HOST = "10.0.2.2:2021"
    }
}
