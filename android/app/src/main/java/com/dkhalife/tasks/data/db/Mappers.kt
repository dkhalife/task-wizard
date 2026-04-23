package com.dkhalife.tasks.data.db

import com.dkhalife.tasks.data.db.entity.LabelEntity
import com.dkhalife.tasks.data.db.entity.TaskEntity
import com.dkhalife.tasks.data.db.entity.TaskWithLabels
import com.dkhalife.tasks.model.Label
import com.dkhalife.tasks.model.Task

fun TaskWithLabels.toDomain(): Task = Task(
    id = task.id,
    title = task.title,
    nextDueDate = task.nextDueDate,
    endDate = task.endDate,
    isRolling = task.isRolling,
    frequency = task.frequency,
    notification = task.notification,
    labels = labels.map { it.toDomain() },
    createdAt = task.createdAt,
    updatedAt = task.updatedAt,
)

fun LabelEntity.toDomain(): Label = Label(
    id = id,
    name = name,
    color = color,
    createdAt = createdAt,
    updatedAt = updatedAt,
)

fun Task.toEntity(
    localId: String? = null,
    localState: String = LocalState.SYNCED,
): TaskEntity = TaskEntity(
    id = id,
    localId = localId,
    title = title,
    nextDueDate = nextDueDate,
    endDate = endDate,
    isRolling = isRolling,
    frequency = frequency,
    notification = notification,
    createdAt = createdAt,
    updatedAt = updatedAt,
    localState = localState,
)

fun Label.toEntity(
    localId: String? = null,
    localState: String = LocalState.SYNCED,
): LabelEntity = LabelEntity(
    id = id,
    localId = localId,
    name = name,
    color = color,
    createdAt = createdAt,
    updatedAt = updatedAt,
    localState = localState,
)
