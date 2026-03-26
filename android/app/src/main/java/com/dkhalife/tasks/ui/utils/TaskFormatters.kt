package com.dkhalife.tasks.ui.utils

import android.content.Context
import androidx.compose.runtime.Composable
import androidx.compose.runtime.State
import androidx.compose.runtime.produceState
import android.text.format.DateFormat
import com.dkhalife.tasks.R
import com.dkhalife.tasks.model.FrequencyType
import com.dkhalife.tasks.model.IntervalUnit
import com.dkhalife.tasks.model.RepeatOn
import com.dkhalife.tasks.model.Task
import java.time.LocalDateTime
import java.time.ZoneId
import java.time.ZonedDateTime
import java.time.temporal.ChronoUnit
import java.util.Date
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

fun formatDueDate(context: Context, ldt: LocalDateTime, now: LocalDateTime): String {
    val today = now.toLocalDate()
    val timeStr = DateFormat.getTimeFormat(context)
        .format(Date.from(ldt.atZone(ZoneId.systemDefault()).toInstant()))
    return when {
        ldt.isBefore(now) -> context.getString(R.string.due_date_past, formatDistance(context, ldt, now))
        ldt.toLocalDate() == today -> context.getString(R.string.due_date_today_at, timeStr)
        ldt.toLocalDate() == today.plusDays(1) -> context.getString(R.string.due_date_tomorrow_at, timeStr)
        else -> context.getString(R.string.due_date_future, formatDistance(context, now, ldt))
    }
}

fun formatDistance(context: Context, from: LocalDateTime, to: LocalDateTime): String {
    val seconds = ChronoUnit.SECONDS.between(from, to)
    val minutes = seconds / 60
    return when {
        seconds < 45 -> context.getString(R.string.duration_less_than_minute)
        seconds < 90 -> context.getString(R.string.duration_1_minute)
        minutes < 45 -> context.getString(R.string.duration_x_minutes, minutes.toInt())
        minutes < 90 -> context.getString(R.string.duration_about_1_hour)
        minutes < 1440 -> context.getString(R.string.duration_about_x_hours, (minutes / 60).toInt())
        minutes < 2520 -> context.getString(R.string.duration_1_day)
        minutes < 43200 -> context.getString(R.string.duration_x_days, (minutes / 1440).toInt())
        minutes < 64800 -> context.getString(R.string.duration_about_1_month)
        minutes < 86400 -> context.getString(R.string.duration_about_2_months)
        minutes < 525600 -> context.getString(R.string.duration_x_months, (minutes / 43200).toInt())
        else -> context.getString(R.string.duration_about_x_years, (minutes / 525600).toInt())
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

fun getDayOfMonthSuffix(context: Context, day: Int): String =
    if (day in 11..13) context.getString(R.string.ordinal_suffix_th)
    else when (day % 10) {
        1 -> context.getString(R.string.ordinal_suffix_st)
        2 -> context.getString(R.string.ordinal_suffix_nd)
        3 -> context.getString(R.string.ordinal_suffix_rd)
        else -> context.getString(R.string.ordinal_suffix_th)
    }

fun getRecurrenceText(context: Context, task: Task, nextDueLdt: LocalDateTime?): String {
    val frequency = task.frequency
    val dayNames = context.resources.getStringArray(R.array.day_names_short)
    val monthNames = context.resources.getStringArray(R.array.month_names_short)
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
                if (suffix.isNotEmpty()) "$every$suffix" else context.getString(R.string.recurrence_custom_label)
            }
            RepeatOn.DAYS_OF_THE_WEEK -> {
                frequency.days?.joinToString(", ") { dayNames.getOrElse(it) { "$it" } }
                    ?: context.getString(R.string.recurrence_weekly_label)
            }
            RepeatOn.DAY_OF_THE_MONTHS -> {
                val day = nextDueLdt?.dayOfMonth ?: 0
                val suffix = getDayOfMonthSuffix(context, day)
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

fun taskGroupNameResId(groupKey: String): Int? = when (groupKey) {
    "overdue" -> R.string.group_overdue
    "today" -> R.string.group_today
    "tomorrow" -> R.string.group_tomorrow
    "this_week" -> R.string.group_this_week
    "next_week" -> R.string.group_next_week
    "later" -> R.string.group_later
    "any_time" -> R.string.group_any_time
    "none" -> R.string.group_none
    else -> null
}
