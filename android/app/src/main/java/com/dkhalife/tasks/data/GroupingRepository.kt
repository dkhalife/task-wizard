package com.dkhalife.tasks.data

import android.content.SharedPreferences
import androidx.core.content.edit
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class GroupingRepository @Inject constructor(
    private val sharedPreferences: SharedPreferences
) {
    fun getTaskGrouping(): TaskGrouping {
        val name = sharedPreferences.getString(AppPreferences.KEY_TASK_GROUPING, null)
        return TaskGrouping.entries.find { it.name == name } ?: TaskGrouping.DUE_DATE
    }

    fun setTaskGrouping(grouping: TaskGrouping) {
        sharedPreferences.edit { putString(AppPreferences.KEY_TASK_GROUPING, grouping.name) }
    }
}
