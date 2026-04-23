package com.dkhalife.tasks.data.db.dao

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import com.dkhalife.tasks.data.db.entity.LabelEntity
import kotlinx.coroutines.flow.Flow

@Dao
interface LabelDao {
    @Query("SELECT * FROM labels WHERE localState != 'PENDING_DELETE' ORDER BY name")
    fun observeLabels(): Flow<List<LabelEntity>>

    @Query("SELECT * FROM labels WHERE id = :id LIMIT 1")
    suspend fun getById(id: Int): LabelEntity?

    @Query("SELECT * FROM labels WHERE localId = :localId LIMIT 1")
    suspend fun getByLocalId(localId: String): LabelEntity?

    @Query("SELECT MIN(id) FROM labels")
    suspend fun getMinId(): Int?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun upsert(entity: LabelEntity)

    @Query("UPDATE labels SET id = :newId, localId = NULL, localState = 'SYNCED' WHERE id = :oldId")
    suspend fun remapId(oldId: Int, newId: Int)

    @Query("DELETE FROM labels WHERE id = :id")
    suspend fun deleteById(id: Int)

    @Query("SELECT id FROM labels")
    suspend fun allIds(): List<Int>

    @Query("SELECT id FROM labels WHERE localState != 'SYNCED'")
    suspend fun dirtyIds(): List<Int>

    @Query("DELETE FROM labels")
    suspend fun deleteAll()
}
