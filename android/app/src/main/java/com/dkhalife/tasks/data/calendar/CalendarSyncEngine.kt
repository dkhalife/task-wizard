package com.dkhalife.tasks.data.calendar

import android.content.ContentResolver
import android.content.Context
import android.content.SharedPreferences
import com.dkhalife.tasks.data.AppPreferences
import com.dkhalife.tasks.data.sync.SyncEngine
import com.dkhalife.tasks.model.NotificationTriggerOptions
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

            val reminderMinutes = computeReminderMinutes(task.notification)
            val hasAlarm = reminderMinutes.isNotEmpty()

            val existingEventId = existingEvents[syncKey]
            if (existingEventId != null) {
                calendarProviderClient.updateEvent(
                    context.contentResolver, existingEventId, task.title, startMillis, endMillis, hasAlarm, accountName
                )
                syncReminders(context.contentResolver, existingEventId, reminderMinutes, accountName)
            } else {
                val eventId = calendarProviderClient.insertEvent(
                    context.contentResolver, calendarId, task.title, startMillis, endMillis, syncKey, hasAlarm, accountName
                )
                syncReminders(context.contentResolver, eventId, reminderMinutes, accountName)
            }
        }

        for ((syncKey, eventId) in existingEvents) {
            if (syncKey !in taskSyncKeys) {
                calendarProviderClient.deleteEvent(context.contentResolver, eventId, accountName)
            }
        }
    }

    private fun computeReminderMinutes(notification: NotificationTriggerOptions): List<Int> {
        if (!notification.enabled) return emptyList()

        val minutes = mutableListOf<Int>()
        if (notification.dueDate || notification.overdue) {
            minutes.add(REMINDER_AT_EVENT)
        }
        if (notification.preDue) {
            minutes.add(REMINDER_PRE_DUE)
        }
        return minutes
    }

    private fun syncReminders(
        contentResolver: ContentResolver,
        eventId: Long,
        reminderMinutes: List<Int>,
        accountName: String
    ) {
        calendarProviderClient.deleteReminders(contentResolver, eventId, accountName)
        for (minutes in reminderMinutes) {
            calendarProviderClient.insertReminder(contentResolver, eventId, minutes, accountName)
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
        private const val REMINDER_AT_EVENT = 0
        private const val REMINDER_PRE_DUE = 180
    }
}
