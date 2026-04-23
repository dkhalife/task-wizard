package com.dkhalife.tasks.repo

import com.dkhalife.tasks.api.TaskWizardApi
import com.dkhalife.tasks.data.LocalIdGenerator
import com.dkhalife.tasks.data.db.LocalState
import com.dkhalife.tasks.data.db.dao.LabelDao
import com.dkhalife.tasks.data.db.dao.OutboxDao
import com.dkhalife.tasks.data.db.entity.LabelEntity
import com.dkhalife.tasks.data.db.entity.OutboxEntity
import com.dkhalife.tasks.data.db.entity.OutboxEntityType
import com.dkhalife.tasks.data.db.entity.OutboxOpType
import com.dkhalife.tasks.data.db.toDomain
import com.dkhalife.tasks.data.network.NetworkMonitor
import com.dkhalife.tasks.data.sync.SyncCoordinator
import com.dkhalife.tasks.model.CreateLabelReq
import com.dkhalife.tasks.model.Label
import com.dkhalife.tasks.model.UpdateLabelReq
import com.dkhalife.tasks.telemetry.TelemetryManager
import com.google.gson.Gson
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import java.util.UUID
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class LabelRepository @Inject constructor(
    private val api: TaskWizardApi,
    private val labelDao: LabelDao,
    private val outboxDao: OutboxDao,
    private val localIdGenerator: LocalIdGenerator,
    private val networkMonitor: NetworkMonitor,
    private val syncCoordinator: SyncCoordinator,
    private val gson: Gson,
    private val telemetryManager: TelemetryManager,
) {
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    val labels: StateFlow<List<Label>> = labelDao.observeLabels()
        .map { rows -> rows.map { it.toDomain() } }
        .stateIn(scope, SharingStarted.Eagerly, emptyList())

    suspend fun refreshLabels(): Result<List<Label>> {
        if (!networkMonitor.isOnline.value) return Result.success(labels.value)
        return try {
            syncCoordinator.syncOnceBlocking()
            Result.success(labels.value)
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to refresh labels: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun createLabel(req: CreateLabelReq): Result<Int> {
        return try {
            val localId = UUID.randomUUID().toString()
            val placeholderId = localIdGenerator.nextId()
            labelDao.upsert(
                LabelEntity(
                    id = placeholderId,
                    localId = localId,
                    name = req.name,
                    color = req.color,
                    localState = LocalState.PENDING_CREATE,
                )
            )
            outboxDao.insert(
                OutboxEntity(
                    entityType = OutboxEntityType.LABEL,
                    opType = OutboxOpType.CREATE,
                    targetLocalId = localId,
                    targetServerId = placeholderId,
                )
            )
            syncCoordinator.syncOnce()
            Result.success(placeholderId)
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to create label locally: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun updateLabel(req: UpdateLabelReq): Result<Unit> {
        return try {
            val existing = labelDao.getById(req.id)
            val nextState = when (existing?.localState) {
                LocalState.PENDING_CREATE -> LocalState.PENDING_CREATE
                else -> LocalState.PENDING_UPDATE
            }
            labelDao.upsert(
                (existing ?: LabelEntity(id = req.id, name = req.name)).copy(
                    name = req.name,
                    color = req.color,
                    localState = nextState,
                )
            )

            if (nextState != LocalState.PENDING_CREATE) {
                outboxDao.insert(
                    OutboxEntity(
                        entityType = OutboxEntityType.LABEL,
                        opType = OutboxOpType.UPDATE,
                        targetServerId = req.id,
                    )
                )
            }
            syncCoordinator.syncOnce()
            Result.success(Unit)
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to update label locally: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun deleteLabel(id: Int): Result<Unit> {
        return try {
            val existing = labelDao.getById(id)
            if (existing?.localState == LocalState.PENDING_CREATE) {
                existing.localId?.let { outboxDao.deleteByLocalId(OutboxEntityType.LABEL, it) }
                labelDao.deleteById(id)
            } else {
                labelDao.upsert(
                    (existing ?: LabelEntity(id = id, name = "")).copy(localState = LocalState.PENDING_DELETE)
                )
                outboxDao.insert(
                    OutboxEntity(
                        entityType = OutboxEntityType.LABEL,
                        opType = OutboxOpType.DELETE,
                        targetServerId = id,
                    )
                )
                syncCoordinator.syncOnce()
            }
            Result.success(Unit)
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to delete label locally: ${e.message}", e)
            Result.failure(e)
        }
    }

    fun updateLabelsFromWebSocket(@Suppress("UNUSED_PARAMETER") labels: List<Label>) {
        syncCoordinator.syncOnce()
    }

    companion object {
        private const val TAG = "LabelRepository"
    }
}
