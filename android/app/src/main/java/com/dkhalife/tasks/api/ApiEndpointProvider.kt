package com.dkhalife.tasks.api

import android.content.SharedPreferences
import androidx.core.content.edit
import com.dkhalife.tasks.BuildConfig
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ApiEndpointProvider @Inject constructor(
    private val sharedPreferences: SharedPreferences
) {
    fun getBaseUrl(): String {
        val host = normalizeHost(
            sharedPreferences.getString(KEY_SERVER_ENDPOINT, DEFAULT_HOST) ?: DEFAULT_HOST
        )
        val scheme = if (host.contains(':') && !host.endsWith(":443")) "http" else "https"
        return "$scheme://$host"
    }

    fun getWebSocketUrl(): String {
        val host = normalizeHost(
            sharedPreferences.getString(KEY_SERVER_ENDPOINT, DEFAULT_HOST) ?: DEFAULT_HOST
        )
        val scheme = if (host.contains(':') && !host.endsWith(":443")) "ws" else "wss"
        return "$scheme://$host/ws"
    }

    fun setServerEndpoint(host: String) {
        sharedPreferences.edit { putString(KEY_SERVER_ENDPOINT, normalizeHost(host)) }
    }

    fun getServerEndpoint(): String {
        return normalizeHost(
            sharedPreferences.getString(KEY_SERVER_ENDPOINT, DEFAULT_HOST) ?: DEFAULT_HOST
        )
    }

    companion object {
        private const val KEY_SERVER_ENDPOINT = "server_endpoint"
        private const val PRODUCTION_HOST = "tasks.dkhalife.com"
        private const val DEBUG_HOST = "10.0.2.2:2021"

        private val DEFAULT_HOST: String
            get() = if (BuildConfig.DEBUG) DEBUG_HOST else PRODUCTION_HOST

        private fun normalizeHost(input: String): String {
            var host = input.trim()
            host = host.removePrefix("https://")
            host = host.removePrefix("http://")
            host = host.trimEnd('/')
            return host.ifEmpty { DEFAULT_HOST }
        }
    }
}
