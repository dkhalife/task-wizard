package app.taskwiz.data.db.entity

import androidx.room.Entity
import androidx.room.PrimaryKey
import app.taskwiz.data.db.LocalState
import app.taskwiz.model.Frequency
import app.taskwiz.model.NotificationTriggerOptions

@Entity(tableName = "tasks")
data class TaskEntity(
    @PrimaryKey val id: Int,
    val localId: String? = null,
    val title: String,
    val nextDueDate: String? = null,
    val endDate: String? = null,
    val isRolling: Boolean = false,
    val frequency: Frequency = Frequency(),
    val notification: NotificationTriggerOptions = NotificationTriggerOptions(),
    val createdAt: String? = null,
    val updatedAt: String? = null,
    val localState: String = LocalState.SYNCED,
)
