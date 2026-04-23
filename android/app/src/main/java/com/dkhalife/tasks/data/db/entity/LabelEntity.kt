package com.dkhalife.tasks.data.db.entity

import androidx.room.Entity
import androidx.room.PrimaryKey
import com.dkhalife.tasks.data.db.LocalState

@Entity(tableName = "labels")
data class LabelEntity(
    @PrimaryKey val id: Int,
    val localId: String? = null,
    val name: String,
    val color: String = "#000000",
    val createdAt: String? = null,
    val updatedAt: String? = null,
    val localState: String = LocalState.SYNCED,
)
