package com.dkhalife.tasks.data.db

import androidx.room.Database
import androidx.room.RoomDatabase
import androidx.room.TypeConverters
import com.dkhalife.tasks.data.db.dao.LabelDao
import com.dkhalife.tasks.data.db.dao.OutboxDao
import com.dkhalife.tasks.data.db.dao.TaskDao
import com.dkhalife.tasks.data.db.entity.LabelEntity
import com.dkhalife.tasks.data.db.entity.OutboxEntity
import com.dkhalife.tasks.data.db.entity.TaskEntity
import com.dkhalife.tasks.data.db.entity.TaskLabelCrossRef

@Database(
    entities = [
        TaskEntity::class,
        LabelEntity::class,
        TaskLabelCrossRef::class,
        OutboxEntity::class,
    ],
    version = 1,
    exportSchema = false,
)
@TypeConverters(Converters::class)
abstract class TaskWizardDatabase : RoomDatabase() {
    abstract fun taskDao(): TaskDao
    abstract fun labelDao(): LabelDao
    abstract fun outboxDao(): OutboxDao

    companion object {
        const val DB_NAME = "task_wizard.db"
    }
}
