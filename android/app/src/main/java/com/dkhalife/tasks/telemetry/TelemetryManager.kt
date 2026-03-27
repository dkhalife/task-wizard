package com.dkhalife.tasks.telemetry

import android.content.Context
import android.content.SharedPreferences
import android.util.Log
import com.dkhalife.tasks.BuildConfig
import com.dkhalife.tasks.data.AppPreferences
import io.opentelemetry.api.OpenTelemetry
import io.opentelemetry.api.common.AttributeKey
import io.opentelemetry.api.common.Attributes
import io.opentelemetry.api.trace.Tracer
import io.opentelemetry.exporter.logging.LoggingSpanExporter
import io.opentelemetry.sdk.OpenTelemetrySdk
import io.opentelemetry.sdk.resources.Resource
import io.opentelemetry.sdk.trace.SdkTracerProvider
import io.opentelemetry.sdk.trace.export.BatchSpanProcessor
import io.opentelemetry.sdk.trace.export.SimpleSpanProcessor
import java.util.UUID
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TelemetryManager @Inject constructor(
    private val sharedPreferences: SharedPreferences
) {

    companion object {
        private const val TAG = "TelemetryManager"
        private const val INSTRUMENTATION_NAME = "com.dkhalife.tasks"
    }

    private var openTelemetry: OpenTelemetry? = null
    private var tracer: Tracer? = null
    private var isInitialized = false

    fun initialize(context: Context) {
        if (isInitialized) {
            return
        }

        if (!isTelemetryEnabled()) {
            Log.i(TAG, "Telemetry disabled by user setting")
            isInitialized = true
            return
        }

        try {
            val deviceId = getOrCreateDeviceId()

            val resource = Resource.create(
                Attributes.of(
                    AttributeKey.stringKey("service.name"), INSTRUMENTATION_NAME,
                    AttributeKey.stringKey("service.version"), BuildConfig.VERSION_NAME,
                    AttributeKey.longKey("service.version.code"), BuildConfig.VERSION_CODE.toLong(),
                    AttributeKey.stringKey("service.version.commit"), BuildConfig.GIT_SHA,
                    AttributeKey.stringKey("device.id"), deviceId
                )
            )

            val connectionStr = BuildConfig.APPINSIGHTS_CONNECTION_STRING
            val (otel, exporterName) = createOpenTelemetry(connectionStr, resource)

            openTelemetry = otel
            tracer = openTelemetry?.getTracer(INSTRUMENTATION_NAME)
            isInitialized = true

            Log.i(TAG, "Telemetry initialized with $exporterName exporter")
        } catch (e: Exception) {
            Log.e(TAG, "Failed to initialize telemetry", e)
            isInitialized = true
        }
    }

    private fun isTelemetryEnabled(): Boolean {
        return sharedPreferences.getBoolean(AppPreferences.KEY_TELEMETRY_ENABLED, false)
    }

    private fun getOrCreateDeviceId(): String {
        val existingId = sharedPreferences.getString(AppPreferences.KEY_DEVICE_IDENTIFIER, null)
        if (!existingId.isNullOrBlank()) {
            return existingId
        }

        val newId = UUID.randomUUID().toString()
        sharedPreferences.edit().putString(AppPreferences.KEY_DEVICE_IDENTIFIER, newId).apply()
        return newId
    }

    private fun createOpenTelemetry(connectionString: String, resource: Resource): Pair<OpenTelemetry, String> {
        val tracerProviderBuilder = SdkTracerProvider.builder()
            .setResource(resource)

        val exporterName = if (connectionString.isNotBlank()) {
            val azureExporter = AzureMonitorSpanExporter.create(connectionString)
            if (azureExporter != null) {
                tracerProviderBuilder.addSpanProcessor(BatchSpanProcessor.builder(azureExporter).build())
                "Azure Monitor"
            } else {
                tracerProviderBuilder.addSpanProcessor(SimpleSpanProcessor.create(LoggingSpanExporter.create()))
                "Logging (Azure Monitor config failed)"
            }
        } else {
            tracerProviderBuilder.addSpanProcessor(SimpleSpanProcessor.create(LoggingSpanExporter.create()))
            "Logging"
        }

        val otel = OpenTelemetrySdk.builder()
            .setTracerProvider(tracerProviderBuilder.build())
            .build()
        return Pair(otel, exporterName)
    }

    fun trackEvent(eventName: String, attributes: Map<String, String> = emptyMap()) {
        if (!isTelemetryEnabled()) {
            return
        }

        val currentTracer = tracer ?: return

        val span = currentTracer.spanBuilder(eventName).startSpan()
        try {
            span.setAttribute("build_number", BuildConfig.VERSION_NAME)
            attributes.forEach { (key, value) ->
                span.setAttribute(key, value)
            }
        } finally {
            span.end()
        }
    }

    fun trackException(throwable: Throwable, attributes: Map<String, String> = emptyMap()) {
        if (!isTelemetryEnabled()) {
            return
        }

        val currentTracer = tracer ?: return

        val span = currentTracer.spanBuilder("exception").startSpan()
        try {
            span.setAttribute("build_number", BuildConfig.VERSION_NAME)
            attributes.forEach { (key, value) ->
                span.setAttribute(key, value)
            }
            span.recordException(throwable)
        } finally {
            span.end()
        }
    }

    fun logInfo(tag: String, message: String, attributes: Map<String, String> = emptyMap()) {
        Log.i(tag, message)

        if (!isTelemetryEnabled()) {
            return
        }

        val currentTracer = tracer ?: return
        val span = currentTracer.spanBuilder("info").startSpan()
        try {
            span.setAttribute("build_number", BuildConfig.VERSION_NAME)
            span.setAttribute("log_level", "info")
            span.setAttribute("app_component", tag)
            span.setAttribute("message", message)
            attributes.forEach { (key, value) ->
                span.setAttribute(key, value)
            }
        } finally {
            span.end()
        }
    }

    fun logWarning(tag: String, message: String, throwable: Throwable? = null) {
        if (throwable != null) {
            Log.w(tag, message, throwable)
        } else {
            Log.w(tag, message)
        }

        if (!isTelemetryEnabled()) {
            return
        }

        val currentTracer = tracer ?: return
        val span = currentTracer.spanBuilder("warning").startSpan()
        try {
            span.setAttribute("build_number", BuildConfig.VERSION_NAME)
            span.setAttribute("log_level", "warn")
            span.setAttribute("app_component", tag)
            span.setAttribute("message", message)
            if (throwable != null) {
                span.recordException(throwable)
            }
        } finally {
            span.end()
        }
    }

    fun logError(tag: String, message: String, throwable: Throwable? = null) {
        if (throwable != null) {
            Log.e(tag, message, throwable)
        } else {
            Log.e(tag, message)
        }

        if (!isTelemetryEnabled()) {
            return
        }

        val currentTracer = tracer ?: return
        val span = currentTracer.spanBuilder("error").startSpan()
        try {
            span.setAttribute("build_number", BuildConfig.VERSION_NAME)
            span.setAttribute("log_level", "error")
            span.setAttribute("app_component", tag)
            span.setAttribute("message", message)
            if (throwable != null) {
                span.recordException(throwable)
            }
        } finally {
            span.end()
        }
    }

    fun logDebug(tag: String, message: String, attributes: Map<String, String> = emptyMap()) {
        Log.d(tag, message)

        if (!isTelemetryEnabled() ||
            !sharedPreferences.getBoolean(AppPreferences.KEY_DEBUG_LOGGING_ENABLED, false)
        ) {
            return
        }

        val currentTracer = tracer ?: return
        val span = currentTracer.spanBuilder("debug").startSpan()
        try {
            span.setAttribute("build_number", BuildConfig.VERSION_NAME)
            span.setAttribute("log_level", "debug")
            span.setAttribute("app_component", tag)
            span.setAttribute("message", message)
            attributes.forEach { (key, value) ->
                span.setAttribute(key, value)
            }
        } finally {
            span.end()
        }
    }

    fun logTrace(tag: String, message: String, attributes: Map<String, String> = emptyMap()) {
        Log.v(tag, message)

        if (!isTelemetryEnabled() ||
            !sharedPreferences.getBoolean(AppPreferences.KEY_DEBUG_LOGGING_ENABLED, false)
        ) {
            return
        }

        val currentTracer = tracer ?: return
        val span = currentTracer.spanBuilder("trace").startSpan()
        try {
            span.setAttribute("build_number", BuildConfig.VERSION_NAME)
            span.setAttribute("log_level", "trace")
            span.setAttribute("app_component", tag)
            span.setAttribute("message", message)
            attributes.forEach { (key, value) ->
                span.setAttribute(key, value)
            }
        } finally {
            span.end()
        }
    }
}
