package com.dkhalife.tasks.repo

import com.dkhalife.tasks.api.TaskWizardApi
import com.dkhalife.tasks.data.LocalIdGenerator
import com.dkhalife.tasks.data.db.LocalState
import com.dkhalife.tasks.data.db.dao.OutboxDao
import com.dkhalife.tasks.data.db.dao.TaskDao
import com.dkhalife.tasks.data.db.entity.OutboxEntity
import com.dkhalife.tasks.data.db.entity.OutboxEntityType
import com.dkhalife.tasks.data.db.entity.OutboxOpType
import com.dkhalife.tasks.data.db.entity.TaskEntity
import com.dkhalife.tasks.data.db.toDomain
import com.dkhalife.tasks.data.network.NetworkMonitor
import com.dkhalife.tasks.data.sync.SyncCoordinator
import com.dkhalife.tasks.model.CreateTaskReq
import com.dkhalife.tasks.model.Task
import com.dkhalife.tasks.model.TaskHistory
import com.dkhalife.tasks.model.UpdateTaskReq
import com.dkhalife.tasks.telemetry.TelemetryManager
import com.google.gson.Gson
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import java.util.UUID
import javax.inject.Inject
import javax.inject.Singleton

data class TaskWithSyncState(
    val task: Task,
    val pendingSync: Boolean,
)

@Singleton
class TaskRepository @Inject constructor(
    private val api: TaskWizardApi,
    private val taskDao: TaskDao,
    private val outboxDao: OutboxDao,
    private val localIdGenerator: LocalIdGenerator,
    private val networkMonitor: NetworkMonitor,
    private val syncCoordinator: SyncCoordinator,
    private val gson: Gson,
    private val telemetryManager: TelemetryManager,
) {
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    val taskStates: StateFlow<List<TaskWithSyncState>> = taskDao.observeTasks()
        .map { rows ->
            rows.map { r ->
                TaskWithSyncState(
                    task = r.toDomain(),
                    pendingSync = r.task.localState != LocalState.SYNCED,
                )
            }
        }
        .stateIn(scope, SharingStarted.Eagerly, emptyList())

    val tasks: StateFlow<List<Task>> = taskDao.observeTasks()
        .map { rows -> rows.map { it.toDomain() } }
        .stateIn(scope, SharingStarted.Eagerly, emptyList())

    private val _completedTasks = MutableStateFlow<List<Task>>(emptyList())
    val completedTasks: StateFlow<List<Task>> = _completedTasks

    suspend fun refreshTasks(): Result<List<Task>> {
        if (!networkMonitor.isOnline.value) {
            return Result.success(tasks.value)
        }
        return try {
            syncCoordinator.syncOnceBlocking()
            Result.success(tasks.value)
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to refresh tasks: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun refreshCompletedTasks(limit: Int = 10, page: Int = 1): Result<List<Task>> {
        return try {
            val response = api.getCompletedTasks(limit, page)
            if (response.isSuccessful) {
                val tasks = response.body()?.tasks ?: emptyList()
                _completedTasks.value = tasks
                Result.success(tasks)
            } else {
                telemetryManager.logError(TAG, "Failed to fetch completed tasks: ${response.code()}")
                Result.failure(Exception("Failed to fetch completed tasks: ${response.code()}"))
            }
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to fetch completed tasks: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun getTask(id: Int): Result<Task> {
        taskDao.getTaskById(id)?.let { return Result.success(it.toDomain()) }
        return try {
            val response = api.getTask(id)
            if (response.isSuccessful) {
                Result.success(response.body()!!.task)
            } else {
                Result.failure(Exception("Failed to fetch task: ${response.code()}"))
            }
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to fetch task: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun getTaskHistory(id: Int): Result<List<TaskHistory>> {
        return try {
            val response = api.getTaskHistory(id)
            if (response.isSuccessful) {
                Result.success(response.body()?.history ?: emptyList())
            } else {
                Result.failure(Exception("Failed to fetch task history: ${response.code()}"))
            }
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to fetch task history: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun createTask(req: CreateTaskReq): Result<Int> {
        return try {
            val localId = UUID.randomUUID().toString()
            val placeholderId = localIdGenerator.nextId()
            val entity = TaskEntity(
                id = placeholderId,
                localId = localId,
                title = req.title,
                nextDueDate = req.nextDueDate,
                endDate = req.endDate,
                isRolling = req.isRolling,
                frequency = req.frequency,
                notification = req.notification,
                localState = LocalState.PENDING_CREATE,
            )
            taskDao.upsert(entity)
            taskDao.replaceLabels(placeholderId, req.labels)
            outboxDao.insert(
                OutboxEntity(
                    entityType = OutboxEntityType.TASK,
                    opType = OutboxOpType.CREATE,
                    targetLocalId = localId,
                    targetServerId = placeholderId,
                )
            )
            syncCoordinator.syncOnce()
            Result.success(placeholderId)
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to create task locally: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun updateTask(req: UpdateTaskReq): Result<Unit> {
        return try {
            val existing = taskDao.getTaskById(req.id)?.task
            val nextState = when (existing?.localState) {
                LocalState.PENDING_CREATE -> LocalState.PENDING_CREATE
                else -> LocalState.PENDING_UPDATE
            }
            val entity = TaskEntity(
                id = req.id,
                localId = existing?.localId,
                title = req.title,
                nextDueDate = req.nextDueDate,
                endDate = req.endDate,
                isRolling = req.isRolling,
                frequency = req.frequency,
                notification = req.notification,
                createdAt = existing?.createdAt,
                updatedAt = existing?.updatedAt,
                localState = nextState,
            )
            taskDao.upsert(entity)
            taskDao.replaceLabels(req.id, req.labels)

            if (nextState == LocalState.PENDING_CREATE) {
                // The outbox CREATE row will reconstruct its payload from the latest DB state at
                // send time, so no additional outbox row is needed here.
            } else {
                outboxDao.insert(
                    OutboxEntity(
                        entityType = OutboxEntityType.TASK,
                        opType = OutboxOpType.UPDATE,
                        targetServerId = req.id,
                    )
                )
            }
            syncCoordinator.syncOnce()
            Result.success(Unit)
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to update task locally: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun deleteTask(id: Int): Result<Unit> {
        return try {
            val existing = taskDao.getTaskById(id)?.task
            if (existing?.localState == LocalState.PENDING_CREATE) {
                // Coalesce: drop the row + any pending ops for this entity, no network op needed.
                existing.localId?.let { outboxDao.deleteByLocalId(OutboxEntityType.TASK, it) }
                taskDao.deleteById(id)
            } else {
                taskDao.setState(id, LocalState.PENDING_DELETE)
                outboxDao.insert(
                    OutboxEntity(
                        entityType = OutboxEntityType.TASK,
                        opType = OutboxOpType.DELETE,
                        targetServerId = id,
                    )
                )
                syncCoordinator.syncOnce()
            }
            Result.success(Unit)
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to delete task locally: ${e.message}", e)
            Result.failure(e)
        }
    }

    suspend fun completeTask(id: Int, endRecurrence: Boolean = false): Result<Unit> =
        enqueueTaskOp(id, OutboxOpType.COMPLETE, endRecurrence.toString())

    suspend fun uncompleteTask(id: Int): Result<Unit> =
        enqueueTaskOp(id, OutboxOpType.UNCOMPLETE, null)

    suspend fun skipTask(id: Int): Result<Unit> =
        enqueueTaskOp(id, OutboxOpType.SKIP, null)

    suspend fun updateDueDate(id: Int, dueDate: String): Result<Unit> {
        return try {
            taskDao.updateDueDate(id, dueDate)
            outboxDao.insert(
                OutboxEntity(
                    entityType = OutboxEntityType.TASK,
                    opType = OutboxOpType.DUE_DATE,
                    targetServerId = id,
                    payloadJson = dueDate,
                )
            )
            syncCoordinator.syncOnce()
            Result.success(Unit)
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to update due date locally: ${e.message}", e)
            Result.failure(e)
        }
    }

    private suspend fun enqueueTaskOp(id: Int, opType: String, payload: String?): Result<Unit> {
        return try {
            taskDao.setState(id, LocalState.PENDING_UPDATE)
            outboxDao.insert(
                OutboxEntity(
                    entityType = OutboxEntityType.TASK,
                    opType = opType,
                    targetServerId = id,
                    payloadJson = payload,
                )
            )
            syncCoordinator.syncOnce()
            Result.success(Unit)
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to enqueue $opType locally: ${e.message}", e)
            Result.failure(e)
        }
    }

    fun updateTasksFromWebSocket(@Suppress("UNUSED_PARAMETER") tasks: List<Task>) {
        // WebSocket pushes trigger a full sync via SyncCoordinator; DB updates flow from there.
        syncCoordinator.syncOnce()
    }

    companion object {
        private const val TAG = "TaskRepository"
    }
}
