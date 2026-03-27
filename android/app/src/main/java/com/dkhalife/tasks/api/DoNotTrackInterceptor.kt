package com.dkhalife.tasks.api

import android.content.SharedPreferences
import com.dkhalife.tasks.data.AppPreferences
import okhttp3.Interceptor
import okhttp3.Response

class DoNotTrackInterceptor(
    private val sharedPreferences: SharedPreferences
) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val telemetryEnabled = sharedPreferences.getBoolean(AppPreferences.KEY_TELEMETRY_ENABLED, false)

        val request = if (!telemetryEnabled) {
            chain.request().newBuilder()
                .header("DNT", "1")
                .build()
        } else {
            chain.request()
        }

        return chain.proceed(request)
    }
}
