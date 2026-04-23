package com.dkhalife.tasks.data.db.dao

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Update
import com.dkhalife.tasks.data.db.entity.OutboxEntity
import kotlinx.coroutines.flow.Flow

@Dao
interface OutboxDao {
    @Insert
    suspend fun insert(entity: OutboxEntity): Long

    @Update
    suspend fun update(entity: OutboxEntity)

    @Query("SELECT * FROM outbox ORDER BY id ASC")
    suspend fun getAll(): List<OutboxEntity>

    @Query("SELECT * FROM outbox ORDER BY id ASC LIMIT 1")
    suspend fun peekNext(): OutboxEntity?

    @Query("SELECT * FROM outbox ORDER BY id ASC")
    fun observeAll(): Flow<List<OutboxEntity>>

    @Query("SELECT COUNT(*) FROM outbox")
    fun observeCount(): Flow<Int>

    @Query("DELETE FROM outbox WHERE id = :id")
    suspend fun deleteById(id: Long)

    @Query("DELETE FROM outbox WHERE entityType = :type AND targetLocalId = :localId")
    suspend fun deleteByLocalId(type: String, localId: String)

    @Query("DELETE FROM outbox WHERE entityType = :type AND targetServerId = :serverId")
    suspend fun deleteByServerId(type: String, serverId: Int)

    @Query("UPDATE outbox SET targetServerId = :newServerId, targetLocalId = NULL WHERE entityType = :type AND targetLocalId = :localId")
    suspend fun remapLocalIdToServer(type: String, localId: String, newServerId: Int)

    @Query("UPDATE outbox SET targetServerId = :newServerId WHERE entityType = :type AND targetServerId = :oldServerId")
    suspend fun remapServerId(type: String, oldServerId: Int, newServerId: Int)

    @Query("DELETE FROM outbox")
    suspend fun deleteAll()
}
