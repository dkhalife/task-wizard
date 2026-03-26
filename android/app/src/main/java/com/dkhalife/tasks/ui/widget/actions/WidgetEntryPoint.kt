package com.dkhalife.tasks.ui.widget.actions

import com.dkhalife.tasks.api.TaskWizardApi
import com.dkhalife.tasks.data.sync.TaskSyncScheduler
import dagger.hilt.EntryPoint
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent

@EntryPoint
@InstallIn(SingletonComponent::class)
interface WidgetEntryPoint {
    fun api(): TaskWizardApi
    fun taskSyncScheduler(): TaskSyncScheduler
}
