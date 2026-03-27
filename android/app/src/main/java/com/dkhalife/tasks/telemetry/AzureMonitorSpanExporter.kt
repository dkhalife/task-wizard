package com.dkhalife.tasks.telemetry

import android.util.Log
import io.opentelemetry.sdk.common.CompletableResultCode
import io.opentelemetry.sdk.trace.data.SpanData
import io.opentelemetry.sdk.trace.export.SpanExporter
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONArray
import org.json.JSONObject
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import java.util.TimeZone
import java.util.concurrent.TimeUnit

class AzureMonitorSpanExporter private constructor(
    private val ingestionEndpoint: String,
    private val instrumentationKey: String,
    private val httpClient: OkHttpClient
) : SpanExporter {

    companion object {
        private const val TAG = "AzureMonitorExporter"
        private val JSON_MEDIA_TYPE = "application/json".toMediaType()

        fun create(connectionString: String): AzureMonitorSpanExporter? {
            Log.i(TAG, "Creating Azure Monitor exporter...")
            val parts = parseConnectionString(connectionString)
            val ingestionEndpoint = parts["IngestionEndpoint"]
            val instrumentationKey = parts["InstrumentationKey"]

            if (ingestionEndpoint == null || instrumentationKey == null) {
                Log.e(TAG, "Invalid connection string: missing IngestionEndpoint or InstrumentationKey")
                return null
            }

            Log.i(TAG, "Ingestion endpoint: $ingestionEndpoint")
            Log.i(TAG, "Instrumentation key: ${instrumentationKey.take(8)}...")

            val httpClient = OkHttpClient.Builder()
                .connectTimeout(10, TimeUnit.SECONDS)
                .writeTimeout(30, TimeUnit.SECONDS)
                .readTimeout(30, TimeUnit.SECONDS)
                .build()

            val exporter = AzureMonitorSpanExporter(
                ingestionEndpoint.trimEnd('/'),
                instrumentationKey,
                httpClient
            )

            Log.i(TAG, "Azure Monitor exporter created successfully, endpoint: ${exporter.trackEndpoint}")
            return exporter
        }

        private fun parseConnectionString(connectionString: String): Map<String, String> {
            return connectionString.split(";")
                .filter { it.isNotBlank() && it.contains("=") }
                .associate {
                    val (key, value) = it.split("=", limit = 2)
                    key.trim() to value.trim()
                }
        }
    }

    private val trackEndpoint = "$ingestionEndpoint/v2.1/track"
    private val isoDateFormat = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss.SSS'Z'", Locale.US).apply {
        timeZone = TimeZone.getTimeZone("UTC")
    }

    override fun export(spans: Collection<SpanData>): CompletableResultCode {
        if (spans.isEmpty()) {
            return CompletableResultCode.ofSuccess()
        }

        Log.d(TAG, "Exporting ${spans.size} spans to Azure Monitor...")
        val result = CompletableResultCode()

        try {
            val telemetryItems = JSONArray()
            for (span in spans) {
                telemetryItems.put(spanToTelemetry(span))
            }

            val requestBody = telemetryItems.toString().toRequestBody(JSON_MEDIA_TYPE)
            val request = Request.Builder()
                .url(trackEndpoint)
                .post(requestBody)
                .build()

            httpClient.newCall(request).execute().use { response ->
                if (response.isSuccessful) {
                    result.succeed()
                } else {
                    val responseBody = response.body?.string()
                    Log.w(TAG, "Failed to export spans: ${response.code} ${response.message}")
                    Log.w(TAG, "Response body: $responseBody")
                    result.fail()
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error exporting spans", e)
            result.fail()
        }

        return result
    }

    private fun spanToTelemetry(span: SpanData): JSONObject {
        val startTime = Date(TimeUnit.NANOSECONDS.toMillis(span.startEpochNanos))
        val durationNanos = span.endEpochNanos - span.startEpochNanos
        val durationMs = TimeUnit.NANOSECONDS.toMillis(durationNanos)

        val tags = JSONObject().apply {
            put("ai.operation.id", span.traceId)
            span.parentSpanContext?.let { parent ->
                if (parent.isValid) {
                    put("ai.operation.parentId", parent.spanId)
                }
            }
            put("ai.cloud.role", span.resource.attributes.get(io.opentelemetry.api.common.AttributeKey.stringKey("service.name")) ?: "unknown")
        }

        val properties = JSONObject().apply {
            put("spanId", span.spanId)
            put("duration_ms", durationMs.toString())
            put("span_kind", span.kind.name)
            put("status", span.status.statusCode.name)
        }
        span.resource.attributes.get(
            io.opentelemetry.api.common.AttributeKey.stringKey("service.version.commit")
        )?.let { commitHash ->
            properties.put("commit_hash", commitHash)
        }
        span.attributes.forEach { key, value ->
            properties.put(key.key, value.toString())
        }

        span.events.forEach { event ->
            if (event.name == "exception") {
                event.attributes.forEach { key, value ->
                    properties.put("exception.${key.key}", value.toString())
                }
            }
        }

        return JSONObject().apply {
            put("name", "Microsoft.ApplicationInsights.$instrumentationKey.Event")
            put("time", isoDateFormat.format(startTime))
            put("iKey", instrumentationKey)
            put("tags", tags)
            put("data", JSONObject().apply {
                put("baseType", "EventData")
                put("baseData", JSONObject().apply {
                    put("ver", 2)
                    put("name", span.name)
                    put("properties", properties)
                })
            })
        }
    }

    override fun flush(): CompletableResultCode {
        return CompletableResultCode.ofSuccess()
    }

    override fun shutdown(): CompletableResultCode {
        httpClient.dispatcher.executorService.shutdown()
        return CompletableResultCode.ofSuccess()
    }
}
