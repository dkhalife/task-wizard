package com.dkhalife.tasks.data.calendar

import android.content.Context
import android.content.SharedPreferences
import com.dkhalife.tasks.data.AppPreferences
import com.dkhalife.tasks.data.sync.SyncEngine
import com.dkhalife.tasks.model.Task
import com.dkhalife.tasks.telemetry.TelemetryManager
import java.time.ZonedDateTime
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class CalendarSyncEngine @Inject constructor(
    private val calendarProviderClient: CalendarProviderClient,
    private val calendarRepository: CalendarRepository,
    private val sharedPreferences: SharedPreferences,
    private val telemetryManager: TelemetryManager
) : SyncEngine {

    override suspend fun sync(context: Context, tasks: List<Task>) {
        if (!sharedPreferences.getBoolean(AppPreferences.KEY_CALENDAR_SYNC, false)) return

        val accountName = calendarRepository.getAccountName()
        var calendarId = calendarProviderClient.getCalendarId(
            context.contentResolver, accountName
        )

        if (calendarId == null) {
            calendarProviderClient.createCalendar(
                context.contentResolver,
                accountName,
                CalendarRepository.CALENDAR_DISPLAY_NAME,
                CalendarRepository.CALENDAR_COLOR
            )
            calendarId = calendarProviderClient.getCalendarId(
                context.contentResolver, accountName
            ) ?: run {
                telemetryManager.logError(TAG, "Calendar not found after creation for account=$accountName")
                return
            }
        }

        val existingEvents = calendarProviderClient.getEventsBySyncData(context.contentResolver, calendarId)
        val taskSyncKeys = mutableSetOf<String>()

        for (task in tasks) {
            val startMillis = parseToMillis(task.nextDueDate) ?: continue
            val endMillis = startMillis + EVENT_DURATION_MS
            val syncKey = task.id.toString()
            taskSyncKeys.add(syncKey)

            val existingEventId = existingEvents[syncKey]
            if (existingEventId != null) {
                calendarProviderClient.updateEvent(
                    context.contentResolver, existingEventId, task.title, startMillis, endMillis, accountName
                )
            } else {
                calendarProviderClient.insertEvent(
                    context.contentResolver, calendarId, task.title, startMillis, endMillis, syncKey, accountName
                )
            }
        }

        for ((syncKey, eventId) in existingEvents) {
            if (syncKey !in taskSyncKeys) {
                calendarProviderClient.deleteEvent(context.contentResolver, eventId, accountName)
            }
        }
    }

    private fun parseToMillis(dateString: String?): Long? {
        if (dateString.isNullOrBlank()) return null
        return try {
            ZonedDateTime.parse(dateString).toInstant().toEpochMilli()
        } catch (e: Exception) {
            telemetryManager.logWarning(TAG, "Failed to parse date: $dateString: ${e.message}", e)
            null
        }
    }

    companion object {
        private const val TAG = "CalendarSyncEngine"
        const val EVENT_DURATION_MS = 15 * 60 * 1000L
    }
}
