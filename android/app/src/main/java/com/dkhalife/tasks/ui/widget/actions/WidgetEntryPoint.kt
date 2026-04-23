package com.dkhalife.tasks.ui.widget.actions

import com.dkhalife.tasks.api.TaskWizardApi
import com.dkhalife.tasks.data.sync.TaskSyncScheduler
import com.dkhalife.tasks.data.widget.WidgetSyncEngine
import com.dkhalife.tasks.repo.TaskRepository
import com.dkhalife.tasks.telemetry.TelemetryManager
import com.google.gson.Gson
import dagger.hilt.EntryPoint
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent

@EntryPoint
@InstallIn(SingletonComponent::class)
interface WidgetEntryPoint {
    fun api(): TaskWizardApi
    fun taskSyncScheduler(): TaskSyncScheduler
    fun widgetSyncEngine(): WidgetSyncEngine
    fun gson(): Gson
    fun telemetryManager(): TelemetryManager
    fun taskRepository(): TaskRepository
}

