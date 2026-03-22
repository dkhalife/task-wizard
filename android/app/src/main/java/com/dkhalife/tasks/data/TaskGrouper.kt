package com.dkhalife.tasks.data

import androidx.compose.ui.graphics.Color
import com.dkhalife.tasks.model.Label
import com.dkhalife.tasks.model.Task
import java.time.DayOfWeek
import java.time.LocalDate
import java.time.LocalDateTime
import java.time.ZoneId
import java.time.ZonedDateTime
import java.time.temporal.TemporalAdjusters

data class TaskGroup(
    val key: String,
    val name: String,
    val tasks: List<Task>,
    val color: Color
)

object TaskGroupColors {
    val OVERDUE = Color(0xFFF03A47)
    val TODAY = Color(0xFFffc107)
    val TOMORROW = Color(0xFFff9800)
    val THIS_WEEK = Color(0xFF4ec1a2)
    val NEXT_WEEK = Color(0xFF00bcd4)
    val LATER = Color(0xFFd7ccc8)
    val ANY_TIME = Color(0xFF90a4ae)
    val NONE = Color.Unspecified
}

object TaskGrouper {

    fun groupByDueDate(tasks: List<Task>): List<TaskGroup> {
        val now = LocalDateTime.now()
        val endOfToday = LocalDate.now().atTime(23, 59, 59, 999_999_999)
        val endOfTomorrow = LocalDate.now().plusDays(1).atTime(23, 59, 59, 999_999_999)
        val endOfThisWeek = LocalDate.now()
            .with(TemporalAdjusters.nextOrSame(DayOfWeek.SUNDAY))
            .atTime(23, 59, 59, 999_999_999)
        val endOfNextWeek = LocalDate.now()
            .with(TemporalAdjusters.nextOrSame(DayOfWeek.SUNDAY))
            .plusWeeks(1)
            .atTime(23, 59, 59, 999_999_999)

        val overdue = mutableListOf<Task>()
        val today = mutableListOf<Task>()
        val tomorrow = mutableListOf<Task>()
        val thisWeek = mutableListOf<Task>()
        val nextWeek = mutableListOf<Task>()
        val later = mutableListOf<Task>()
        val anyTime = mutableListOf<Task>()

        for (task in tasks) {
            val dueDateStr = task.nextDueDate
            if (dueDateStr == null) {
                anyTime.add(task)
                continue
            }

            val dueDate = try {
                ZonedDateTime.parse(dueDateStr)
                    .withZoneSameInstant(ZoneId.systemDefault())
                    .toLocalDateTime()
            } catch (_: Exception) {
                anyTime.add(task)
                continue
            }

            when {
                now >= dueDate -> overdue.add(task)
                endOfToday >= dueDate -> today.add(task)
                endOfTomorrow >= dueDate -> tomorrow.add(task)
                endOfThisWeek >= dueDate -> thisWeek.add(task)
                endOfNextWeek >= dueDate -> nextWeek.add(task)
                else -> later.add(task)
            }
        }

        return listOf(
            TaskGroup("overdue", "Overdue", sortByDueDate(overdue), TaskGroupColors.OVERDUE),
            TaskGroup("today", "Today", sortByDueDate(today), TaskGroupColors.TODAY),
            TaskGroup("tomorrow", "Tomorrow", sortByDueDate(tomorrow), TaskGroupColors.TOMORROW),
            TaskGroup("this_week", "This week", sortByDueDate(thisWeek), TaskGroupColors.THIS_WEEK),
            TaskGroup("next_week", "Next week", sortByDueDate(nextWeek), TaskGroupColors.NEXT_WEEK),
            TaskGroup("later", "Later", sortByDueDate(later), TaskGroupColors.LATER),
            TaskGroup("any_time", "Any time", sortByDueDate(anyTime), TaskGroupColors.ANY_TIME),
        ).filter { it.tasks.isNotEmpty() }
    }

    fun groupByLabel(tasks: List<Task>, labels: List<Label>): List<TaskGroup> {
        val groups = mutableListOf<TaskGroup>()
        val seenLabelIds = mutableSetOf<Int>()

        val allLabels = buildList {
            addAll(labels)
            for (task in tasks) {
                for (label in task.labels) {
                    if (label.id !in seenLabelIds && labels.none { it.id == label.id }) {
                        add(label)
                    }
                    seenLabelIds.add(label.id)
                }
            }
        }

        for (label in allLabels) {
            val matching = tasks.filter { task ->
                task.labels.any { it.id == label.id }
            }
            if (matching.isNotEmpty()) {
                val color = try {
                    Color(android.graphics.Color.parseColor(label.color))
                } catch (_: Exception) {
                    TaskGroupColors.NONE
                }
                groups.add(TaskGroup("label_${label.id}", label.name, sortByDueDate(matching), color))
            }
        }

        val unlabeled = tasks.filter { it.labels.isEmpty() }
        if (unlabeled.isNotEmpty()) {
            groups.add(TaskGroup("none", "None", sortByDueDate(unlabeled), TaskGroupColors.NONE))
        }

        return groups
    }

    private fun sortByDueDate(tasks: List<Task>): List<Task> {
        return tasks.sortedWith(compareBy(nullsLast()) { task ->
            task.nextDueDate?.let {
                try {
                    ZonedDateTime.parse(it).toInstant()
                } catch (_: Exception) {
                    null
                }
            }
        })
    }
}
