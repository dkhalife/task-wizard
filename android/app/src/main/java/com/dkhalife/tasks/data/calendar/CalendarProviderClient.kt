package com.dkhalife.tasks.data.calendar

import android.content.ContentResolver
import android.content.ContentUris
import android.content.ContentValues
import android.database.Cursor
import android.net.Uri
import android.provider.CalendarContract
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class CalendarProviderClient @Inject constructor() {

    fun getCalendarId(contentResolver: ContentResolver, accountName: String): Long? {
        val projection = arrayOf(CalendarContract.Calendars._ID)
        val selection = "${CalendarContract.Calendars.ACCOUNT_NAME} = ? AND " +
                "${CalendarContract.Calendars.ACCOUNT_TYPE} = ?"
        val selectionArgs = arrayOf(accountName, ACCOUNT_TYPE)

        var cursor: Cursor? = null
        return try {
            cursor = contentResolver.query(
                CalendarContract.Calendars.CONTENT_URI,
                projection,
                selection,
                selectionArgs,
                null
            )
            if (cursor != null && cursor.moveToFirst()) {
                cursor.getLong(0)
            } else {
                null
            }
        } finally {
            cursor?.close()
        }
    }

    fun createCalendar(
        contentResolver: ContentResolver,
        accountName: String,
        displayName: String,
        color: Int
    ): Long {
        val values = ContentValues().apply {
            put(CalendarContract.Calendars.ACCOUNT_NAME, accountName)
            put(CalendarContract.Calendars.ACCOUNT_TYPE, ACCOUNT_TYPE)
            put(CalendarContract.Calendars.NAME, accountName)
            put(CalendarContract.Calendars.CALENDAR_DISPLAY_NAME, displayName)
            put(CalendarContract.Calendars.CALENDAR_COLOR, color)
            put(CalendarContract.Calendars.CALENDAR_ACCESS_LEVEL, CalendarContract.Calendars.CAL_ACCESS_OWNER)
            put(CalendarContract.Calendars.VISIBLE, 1)
            put(CalendarContract.Calendars.SYNC_EVENTS, 1)
            put(CalendarContract.Calendars.OWNER_ACCOUNT, accountName)
        }

        val uri = contentResolver.insert(asSyncAdapter(CalendarContract.Calendars.CONTENT_URI, accountName), values)
        return uri?.lastPathSegment?.toLongOrNull()
            ?: throw IllegalStateException("Failed to create calendar")
    }

    fun deleteCalendar(contentResolver: ContentResolver, calendarId: Long, accountName: String) {
        val uri = ContentUris.withAppendedId(CalendarContract.Calendars.CONTENT_URI, calendarId)
        val syncAdapterUri = asSyncAdapter(uri, accountName)
        contentResolver.delete(syncAdapterUri, null, null)
    }

    fun insertEvent(
        contentResolver: ContentResolver,
        calendarId: Long,
        title: String,
        startMillis: Long,
        endMillis: Long,
        syncData: String,
        accountName: String
    ): Long {
        val values = ContentValues().apply {
            put(CalendarContract.Events.CALENDAR_ID, calendarId)
            put(CalendarContract.Events.TITLE, title)
            put(CalendarContract.Events.DTSTART, startMillis)
            put(CalendarContract.Events.DTEND, endMillis)
            put(CalendarContract.Events.EVENT_TIMEZONE, "UTC")
            put(CalendarContract.Events.HAS_ALARM, 0)
            put(CalendarContract.Events.SYNC_DATA1, syncData)
        }

        val uri = contentResolver.insert(asSyncAdapter(CalendarContract.Events.CONTENT_URI, accountName), values)
        return uri?.lastPathSegment?.toLongOrNull()
            ?: throw IllegalStateException("Failed to insert event")
    }

    fun updateEvent(
        contentResolver: ContentResolver,
        eventId: Long,
        title: String,
        startMillis: Long,
        endMillis: Long,
        accountName: String
    ) {
        val values = ContentValues().apply {
            put(CalendarContract.Events.TITLE, title)
            put(CalendarContract.Events.DTSTART, startMillis)
            put(CalendarContract.Events.DTEND, endMillis)
        }

        val uri = ContentUris.withAppendedId(CalendarContract.Events.CONTENT_URI, eventId)
        contentResolver.update(asSyncAdapter(uri, accountName), values, null, null)
    }

    fun deleteEvent(contentResolver: ContentResolver, eventId: Long, accountName: String) {
        val uri = ContentUris.withAppendedId(CalendarContract.Events.CONTENT_URI, eventId)
        contentResolver.delete(asSyncAdapter(uri, accountName), null, null)
    }

    fun getEventsBySyncData(contentResolver: ContentResolver, calendarId: Long): Map<String, Long> {
        val projection = arrayOf(
            CalendarContract.Events._ID,
            CalendarContract.Events.SYNC_DATA1
        )
        val selection = "${CalendarContract.Events.CALENDAR_ID} = ?"
        val selectionArgs = arrayOf(calendarId.toString())

        val result = mutableMapOf<String, Long>()
        var cursor: Cursor? = null
        try {
            cursor = contentResolver.query(
                CalendarContract.Events.CONTENT_URI,
                projection,
                selection,
                selectionArgs,
                null
            )
            if (cursor != null) {
                while (cursor.moveToNext()) {
                    val eventId = cursor.getLong(0)
                    val syncData = cursor.getString(1) ?: continue
                    result[syncData] = eventId
                }
            }
        } finally {
            cursor?.close()
        }
        return result
    }

    private fun asSyncAdapter(uri: Uri, accountName: String): Uri {
        return uri.buildUpon()
            .appendQueryParameter(CalendarContract.CALLER_IS_SYNCADAPTER, "true")
            .appendQueryParameter(CalendarContract.Calendars.ACCOUNT_NAME, accountName)
            .appendQueryParameter(CalendarContract.Calendars.ACCOUNT_TYPE, ACCOUNT_TYPE)
            .build()
    }

    companion object {
        const val ACCOUNT_TYPE = "com.dkhalife.tasks"
    }
}
