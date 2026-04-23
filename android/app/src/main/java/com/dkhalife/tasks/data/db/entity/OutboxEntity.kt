package com.dkhalife.tasks.data.db.entity

import androidx.room.Entity
import androidx.room.PrimaryKey

object OutboxEntityType {
    const val TASK = "TASK"
    const val LABEL = "LABEL"
}

object OutboxOpType {
    const val CREATE = "CREATE"
    const val UPDATE = "UPDATE"
    const val DELETE = "DELETE"
    const val COMPLETE = "COMPLETE"
    const val UNCOMPLETE = "UNCOMPLETE"
    const val SKIP = "SKIP"
    const val DUE_DATE = "DUE_DATE"
}

@Entity(tableName = "outbox")
data class OutboxEntity(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    val entityType: String,
    val opType: String,
    val targetLocalId: String? = null,
    val targetServerId: Int? = null,
    val payloadJson: String? = null,
    val createdAt: Long = System.currentTimeMillis(),
    val attempts: Int = 0,
    val lastError: String? = null,
)
