package com.dkhalife.tasks.data.db.dao

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Transaction
import com.dkhalife.tasks.data.db.entity.TaskEntity
import com.dkhalife.tasks.data.db.entity.TaskLabelCrossRef
import com.dkhalife.tasks.data.db.entity.TaskWithLabels
import kotlinx.coroutines.flow.Flow

@Dao
interface TaskDao {
    @Transaction
    @Query("SELECT * FROM tasks WHERE localState != 'PENDING_DELETE' ORDER BY id")
    fun observeTasks(): Flow<List<TaskWithLabels>>

    @Transaction
    @Query("SELECT * FROM tasks WHERE id = :id LIMIT 1")
    suspend fun getTaskById(id: Int): TaskWithLabels?

    @Transaction
    @Query("SELECT * FROM tasks WHERE localId = :localId LIMIT 1")
    suspend fun getTaskByLocalId(localId: String): TaskWithLabels?

    @Query("SELECT * FROM tasks WHERE localState != 'SYNCED'")
    suspend fun getDirtyTasks(): List<TaskEntity>

    @Query("SELECT MIN(id) FROM tasks")
    suspend fun getMinId(): Int?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun upsert(entity: TaskEntity)

    @Query("UPDATE tasks SET localState = :state WHERE id = :id")
    suspend fun setState(id: Int, state: String)

    @Query("UPDATE tasks SET id = :newId, localId = NULL, localState = 'SYNCED' WHERE id = :oldId")
    suspend fun remapId(oldId: Int, newId: Int)

    @Query("DELETE FROM tasks WHERE id = :id")
    suspend fun deleteById(id: Int)

    @Query("DELETE FROM task_labels WHERE taskId = :taskId")
    suspend fun clearLabelsFor(taskId: Int)

    @Insert(onConflict = OnConflictStrategy.IGNORE)
    suspend fun insertLabelRef(ref: TaskLabelCrossRef)

    @Transaction
    suspend fun replaceLabels(taskId: Int, labelIds: List<Int>) {
        clearLabelsFor(taskId)
        labelIds.forEach { insertLabelRef(TaskLabelCrossRef(taskId, it)) }
    }

    @Query("UPDATE tasks SET nextDueDate = :dueDate, localState = CASE WHEN localState = 'SYNCED' THEN 'PENDING_UPDATE' ELSE localState END WHERE id = :id")
    suspend fun updateDueDate(id: Int, dueDate: String)

    @Query("SELECT id FROM tasks")
    suspend fun allIds(): List<Int>

    @Query("SELECT id FROM tasks WHERE localState != 'SYNCED'")
    suspend fun dirtyIds(): List<Int>

    @Query("DELETE FROM tasks")
    suspend fun deleteAll()
}
