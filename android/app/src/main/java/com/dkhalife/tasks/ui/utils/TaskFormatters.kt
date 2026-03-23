package com.dkhalife.tasks.ui.utils

import androidx.compose.runtime.Composable
import androidx.compose.runtime.State
import androidx.compose.runtime.produceState
import com.dkhalife.tasks.model.FrequencyType
import com.dkhalife.tasks.model.IntervalUnit
import com.dkhalife.tasks.model.RepeatOn
import com.dkhalife.tasks.model.Task
import java.time.LocalDateTime
import java.time.ZoneId
import java.time.ZonedDateTime
import java.time.format.DateTimeFormatter
import java.time.temporal.ChronoUnit
import java.util.Locale
import kotlinx.coroutines.delay

fun parseDueDate(dateStr: String?): LocalDateTime? =
    dateStr?.let {
        try {
            ZonedDateTime.parse(it).withZoneSameInstant(ZoneId.systemDefault()).toLocalDateTime()
        } catch (_: Exception) { null }
    }

@Composable
fun rememberTickingNow(): State<LocalDateTime> = produceState(LocalDateTime.now()) {
    while (true) {
        delay(60_000L)
        value = LocalDateTime.now()
    }
}

fun formatDueDate(ldt: LocalDateTime, now: LocalDateTime): String {
    val today = now.toLocalDate()
    val timeStr = ldt.format(DateTimeFormatter.ofPattern("hh:mm a", Locale.ENGLISH))
    return when {
        ldt.isBefore(now) -> "${formatDistance(ldt, now)} ago"
        ldt.toLocalDate() == today -> "Today at $timeStr"
        ldt.toLocalDate() == today.plusDays(1) -> "Tomorrow at $timeStr"
        else -> "in ${formatDistance(now, ldt)}"
    }
}

fun formatDistance(from: LocalDateTime, to: LocalDateTime): String {
    val seconds = ChronoUnit.SECONDS.between(from, to)
    val minutes = seconds / 60
    return when {
        seconds < 45 -> "less than a minute"
        seconds < 90 -> "1 minute"
        minutes < 45 -> "$minutes minutes"
        minutes < 90 -> "about 1 hour"
        minutes < 1440 -> "about ${minutes / 60} hours"
        minutes < 2520 -> "1 day"
        minutes < 43200 -> "${minutes / 1440} days"
        minutes < 64800 -> "about 1 month"
        minutes < 86400 -> "about 2 months"
        minutes < 525600 -> "${minutes / 43200} months"
        else -> "about ${minutes / 525600} years"
    }
}

fun intervalUnitSuffix(unit: String?): String = when (unit) {
    IntervalUnit.HOURS -> "h"
    IntervalUnit.DAYS -> "d"
    IntervalUnit.WEEKS -> "w"
    IntervalUnit.MONTHS -> "m"
    IntervalUnit.YEARS -> "y"
    else -> ""
}

fun getDayOfMonthSuffix(day: Int): String =
    if (day in 11..13) "th"
    else when (day % 10) {
        1 -> "st"
        2 -> "nd"
        3 -> "rd"
        else -> "th"
    }

fun getRecurrenceText(task: Task, nextDueLdt: LocalDateTime?): String {
    val frequency = task.frequency
    return when (frequency.type) {
        FrequencyType.ONCE -> ""
        FrequencyType.DAILY -> "1d"
        FrequencyType.WEEKLY -> "1w"
        FrequencyType.MONTHLY -> "1m"
        FrequencyType.YEARLY -> "1y"
        FrequencyType.CUSTOM -> when (frequency.on) {
            RepeatOn.INTERVAL -> {
                val every = frequency.every ?: 1
                val suffix = intervalUnitSuffix(frequency.unit)
                if (suffix.isNotEmpty()) "$every$suffix" else "Custom"
            }
            RepeatOn.DAYS_OF_THE_WEEK -> {
                val dayNames = arrayOf("Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat")
                frequency.days?.joinToString(", ") { dayNames.getOrElse(it) { "$it" } } ?: "Weekly"
            }
            RepeatOn.DAY_OF_THE_MONTHS -> {
                val day = nextDueLdt?.dayOfMonth ?: 0
                val suffix = getDayOfMonthSuffix(day)
                val monthNames = arrayOf("Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec")
                val months = frequency.months?.joinToString(", ") {
                    monthNames.getOrElse(it) { "$it" }
                } ?: ""
                "${day}${suffix} of $months"
            }
            else -> ""
        }
        else -> ""
    }
}

fun hasActiveNotification(task: Task): Boolean =
    task.notification.enabled &&
        (task.notification.dueDate || task.notification.preDue || task.notification.overdue)
