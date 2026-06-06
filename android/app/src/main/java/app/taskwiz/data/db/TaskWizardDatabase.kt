package app.taskwiz.data.db

import androidx.room.Database
import androidx.room.RoomDatabase
import androidx.room.TypeConverters
import app.taskwiz.data.db.dao.LabelDao
import app.taskwiz.data.db.dao.OutboxDao
import app.taskwiz.data.db.dao.TaskDao
import app.taskwiz.data.db.entity.LabelEntity
import app.taskwiz.data.db.entity.OutboxEntity
import app.taskwiz.data.db.entity.TaskEntity
import app.taskwiz.data.db.entity.TaskLabelCrossRef

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
