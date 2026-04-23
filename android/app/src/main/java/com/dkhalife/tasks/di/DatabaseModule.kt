package com.dkhalife.tasks.di

import android.content.Context
import androidx.room.Room
import com.dkhalife.tasks.data.db.TaskWizardDatabase
import com.dkhalife.tasks.data.db.dao.LabelDao
import com.dkhalife.tasks.data.db.dao.OutboxDao
import com.dkhalife.tasks.data.db.dao.TaskDao
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object DatabaseModule {

    @Provides
    @Singleton
    fun provideDatabase(@ApplicationContext context: Context): TaskWizardDatabase =
        Room.databaseBuilder(context, TaskWizardDatabase::class.java, TaskWizardDatabase.DB_NAME)
            .fallbackToDestructiveMigration(dropAllTables = true)
            .build()

    @Provides
    fun provideTaskDao(db: TaskWizardDatabase): TaskDao = db.taskDao()

    @Provides
    fun provideLabelDao(db: TaskWizardDatabase): LabelDao = db.labelDao()

    @Provides
    fun provideOutboxDao(db: TaskWizardDatabase): OutboxDao = db.outboxDao()
}
