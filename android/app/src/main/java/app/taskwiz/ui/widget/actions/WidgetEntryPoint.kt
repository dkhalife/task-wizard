package app.taskwiz.ui.widget.actions

import app.taskwiz.api.TaskWizardApi
import app.taskwiz.data.sync.TaskSyncScheduler
import app.taskwiz.data.widget.WidgetSyncEngine
import app.taskwiz.repo.TaskRepository
import app.taskwiz.telemetry.TelemetryManager
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

