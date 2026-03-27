package com.dkhalife.tasks.api

import com.dkhalife.tasks.auth.AuthTokenProvider
import com.dkhalife.tasks.telemetry.TelemetryManager
import kotlinx.coroutines.runBlocking
import okhttp3.Authenticator
import okhttp3.Interceptor
import okhttp3.Request
import okhttp3.Response
import okhttp3.Route

class AuthInterceptor(
    private val tokenProvider: AuthTokenProvider,
    private val telemetryManager: TelemetryManager
) : Interceptor, Authenticator {
    companion object {
        private const val TAG = "AuthInterceptor"
    }

    override fun intercept(chain: Interceptor.Chain): Response {
        val token = tokenProvider.getCachedAccessToken()

        val request = if (!token.isNullOrBlank()) {
            chain.request().newBuilder()
                .addHeader("Authorization", "Bearer $token")
                .build()
        } else {
            chain.request()
        }

        return chain.proceed(request)
    }

    override fun authenticate(route: Route?, response: Response): Request? {
        if (response.request.header("Authorization-Retry") != null) {
            telemetryManager.logWarning(TAG, "Auth retry exhausted")
            return null
        }

        val freshToken = runBlocking { tokenProvider.getAccessToken() } ?: run {
            telemetryManager.logWarning(TAG, "Token refresh failed")
            return null
        }

        return response.request.newBuilder()
            .removeHeader("Authorization")
            .addHeader("Authorization", "Bearer $freshToken")
            .addHeader("Authorization-Retry", "true")
            .build()
    }
}
