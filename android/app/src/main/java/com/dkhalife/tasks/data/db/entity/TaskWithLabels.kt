package com.dkhalife.tasks.data.db.entity

import androidx.room.Embedded
import androidx.room.Junction
import androidx.room.Relation

data class TaskWithLabels(
    @Embedded val task: TaskEntity,
    @Relation(
        parentColumn = "id",
        entityColumn = "id",
        associateBy = Junction(
            value = TaskLabelCrossRef::class,
            parentColumn = "taskId",
            entityColumn = "labelId",
        ),
    )
    val labels: List<LabelEntity>,
)
