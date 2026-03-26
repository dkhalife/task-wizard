package com.dkhalife.tasks.data.calendar

import android.content.Context
import android.content.SharedPreferences
import com.dkhalife.tasks.data.AppPreferences
import com.dkhalife.tasks.data.sync.SyncEngine
import com.dkhalife.tasks.model.Task
import java.time.ZonedDateTime
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class CalendarSyncEngine @Inject constructor(
    private val calendarProviderClient: CalendarProviderClient,
    private val sharedPreferences: SharedPreferences
) : SyncEngine {

    override suspend fun sync(context: Context, tasks: List<Task>) {
        if (!sharedPreferences.getBoolean(AppPreferences.KEY_CALENDAR_SYNC, false)) return

        var calendarId = calendarProviderClient.getCalendarId(
            context.contentResolver, CalendarRepository.ACCOUNT_NAME
        )

        if (calendarId == null) {
            calendarProviderClient.createCalendar(
                context.contentResolver,
                CalendarRepository.ACCOUNT_NAME,
                CalendarRepository.CALENDAR_DISPLAY_NAME,
                CalendarRepository.CALENDAR_COLOR
            )
            calendarId = calendarProviderClient.getCalendarId(
                context.contentResolver, CalendarRepository.ACCOUNT_NAME
            ) ?: return
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
                    context.contentResolver, existingEventId, task.title, startMillis, endMillis
                )
            } else {
                calendarProviderClient.insertEvent(
                    context.contentResolver, calendarId, task.title, startMillis, endMillis, syncKey
                )
            }
        }

        for ((syncKey, eventId) in existingEvents) {
            if (syncKey !in taskSyncKeys) {
                calendarProviderClient.deleteEvent(context.contentResolver, eventId)
            }
        }
    }

    private fun parseToMillis(dateString: String?): Long? {
        if (dateString.isNullOrBlank()) return null
        return try {
            ZonedDateTime.parse(dateString).toInstant().toEpochMilli()
        } catch (_: Exception) {
            null
        }
    }

    companion object {
        const val EVENT_DURATION_MS = 15 * 60 * 1000L
    }
}
