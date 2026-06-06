package app.taskwiz.di

import app.taskwiz.data.calendar.CalendarSyncEngine
import app.taskwiz.data.sync.SyncEngine
import app.taskwiz.data.widget.WidgetSyncEngine
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent

@Module
@InstallIn(SingletonComponent::class)
object SyncModule {

    @Provides
    fun provideSyncEngines(
        calendarSyncEngine: CalendarSyncEngine,
        widgetSyncEngine: WidgetSyncEngine
    ): List<@JvmSuppressWildcards SyncEngine> {
        return listOf(calendarSyncEngine, widgetSyncEngine)
    }
}
