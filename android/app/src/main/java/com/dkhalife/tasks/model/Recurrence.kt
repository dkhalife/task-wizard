package com.dkhalife.tasks.model

data class Frequency(
    val type: String = "once",
    val on: String? = null,
    val every: Int? = null,
    val unit: String? = null,
    val days: List<Int>? = null,
    val months: List<Int>? = null
)

object FrequencyType {
    const val ONCE = "once"
    const val DAILY = "daily"
    const val WEEKLY = "weekly"
    const val MONTHLY = "monthly"
    const val YEARLY = "yearly"
    const val CUSTOM = "custom"
}

object RepeatOn {
    const val INTERVAL = "interval"
    const val DAYS_OF_THE_WEEK = "days_of_the_week"
    const val DAY_OF_THE_MONTHS = "day_of_the_months"
}

object IntervalUnit {
    const val HOURS = "hours"
    const val DAYS = "days"
    const val WEEKS = "weeks"
    const val MONTHS = "months"
    const val YEARS = "years"
}
