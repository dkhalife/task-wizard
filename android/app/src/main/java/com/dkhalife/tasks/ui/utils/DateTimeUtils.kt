package com.dkhalife.tasks.ui.utils

import java.time.ZoneId
import java.time.ZonedDateTime
import java.time.format.DateTimeFormatter
import java.util.Locale

fun parseIsoDateTime(iso: String?): ZonedDateTime? = iso?.let {
    runCatching { ZonedDateTime.parse(it).withZoneSameInstant(ZoneId.systemDefault()) }.getOrNull()
}

fun ZonedDateTime.toIsoString(): String =
    format(DateTimeFormatter.ISO_OFFSET_DATE_TIME)

fun ZonedDateTime.toDisplayString(): String =
    format(DateTimeFormatter.ofPattern("MM/dd/yyyy, hh:mm a", Locale.ENGLISH))
