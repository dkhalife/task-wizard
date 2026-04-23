package com.dkhalife.tasks.data.db.entity

import androidx.room.Entity
import androidx.room.PrimaryKey
import com.dkhalife.tasks.data.db.LocalState
import com.dkhalife.tasks.model.Frequency
import com.dkhalife.tasks.model.NotificationTriggerOptions

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
