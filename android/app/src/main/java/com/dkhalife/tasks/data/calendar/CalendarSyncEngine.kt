package com.dkhalife.tasks.data.calendar

import android.content.ContentResolver
import com.dkhalife.tasks.model.Task
import java.text.SimpleDateFormat
import java.util.Locale
import java.util.TimeZone
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class CalendarSyncEngine @Inject constructor(
    private val calendarProviderClient: CalendarProviderClient
) {

    fun sync(contentResolver: ContentResolver, calendarId: Long, tasks: List<Task>) {
        val existingEvents = calendarProviderClient.getEventsBySyncData(contentResolver, calendarId)
        val taskSyncKeys = mutableSetOf<String>()

        for (task in tasks) {
            val startMillis = parseToMillis(task.nextDueDate) ?: continue
            val endMillis = startMillis + EVENT_DURATION_MS
            val syncKey = task.id.toString()
            taskSyncKeys.add(syncKey)

            val existingEventId = existingEvents[syncKey]
            if (existingEventId != null) {
                calendarProviderClient.updateEvent(
                    contentResolver, existingEventId, task.title, startMillis, endMillis
                )
            } else {
                calendarProviderClient.insertEvent(
                    contentResolver, calendarId, task.title, startMillis, endMillis, syncKey
                )
            }
        }

        for ((syncKey, eventId) in existingEvents) {
            if (syncKey !in taskSyncKeys) {
                calendarProviderClient.deleteEvent(contentResolver, eventId)
            }
        }
    }

    private fun parseToMillis(dateString: String?): Long? {
        if (dateString.isNullOrBlank()) return null
        return try {
            val format = SimpleDateFormat(ISO_8601_FORMAT, Locale.US).apply {
                timeZone = TimeZone.getTimeZone("UTC")
            }
            format.parse(dateString)?.time
        } catch (_: Exception) {
            null
        }
    }

    companion object {
        const val EVENT_DURATION_MS = 15 * 60 * 1000L
        private const val ISO_8601_FORMAT = "yyyy-MM-dd'T'HH:mm:ss'Z'"
    }
}
