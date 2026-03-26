package com.dkhalife.tasks.di

import com.dkhalife.tasks.data.calendar.CalendarSyncEngine
import com.dkhalife.tasks.data.sync.SyncEngine
import com.dkhalife.tasks.data.widget.WidgetSyncEngine
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
