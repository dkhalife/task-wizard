package com.dkhalife.tasks.data.sync

import com.dkhalife.tasks.api.TaskWizardApi
import com.dkhalife.tasks.data.db.LocalState
import com.dkhalife.tasks.data.db.dao.LabelDao
import com.dkhalife.tasks.data.db.dao.OutboxDao
import com.dkhalife.tasks.data.db.dao.TaskDao
import com.dkhalife.tasks.data.db.entity.LabelEntity
import com.dkhalife.tasks.data.db.entity.OutboxEntity
import com.dkhalife.tasks.data.db.entity.OutboxEntityType
import com.dkhalife.tasks.data.db.entity.OutboxOpType
import com.dkhalife.tasks.data.db.entity.TaskEntity
import com.dkhalife.tasks.data.db.entity.TaskLabelCrossRef
import com.dkhalife.tasks.data.db.toEntity
import com.dkhalife.tasks.data.network.NetworkMonitor
import com.dkhalife.tasks.model.CreateLabelReq
import com.dkhalife.tasks.model.CreateTaskReq
import com.dkhalife.tasks.model.UpdateDueDateReq
import com.dkhalife.tasks.model.UpdateLabelReq
import com.dkhalife.tasks.model.UpdateTaskReq
import com.dkhalife.tasks.telemetry.TelemetryManager
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Drains the outbox of pending mutations against the server, remaps local placeholder ids to
 * server-assigned ids, then refreshes tasks/labels caches so the local DB matches the server.
 *
 * Ordering guarantees: processes outbox rows strictly in insertion order (by id). After a
 * successful CREATE, any later outbox rows and DB rows still referencing the local placeholder
 * id are rewritten to the new server id before the next op runs.
 */
@Singleton
class SyncCoordinator @Inject constructor(
    private val api: TaskWizardApi,
    private val taskDao: TaskDao,
    private val labelDao: LabelDao,
    private val outboxDao: OutboxDao,
    private val networkMonitor: NetworkMonitor,
    private val telemetryManager: TelemetryManager,
) {
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    private val mutex = Mutex()

    @Volatile
    private var activeJob: Job? = null

    fun syncOnce() {
        if (!networkMonitor.isOnline.value) return
        if (activeJob?.isActive == true) return
        activeJob = scope.launch {
            mutex.withLock {
                try {
                    flushOutbox()
                    refreshAll()
                } catch (e: Exception) {
                    telemetryManager.logError(TAG, "Sync cycle failed: ${e.message}", e)
                }
            }
        }
    }

    suspend fun syncOnceBlocking(): Boolean {
        if (!networkMonitor.isOnline.value) return false
        return mutex.withLock {
            try {
                flushOutbox()
                refreshAll()
                true
            } catch (e: Exception) {
                telemetryManager.logError(TAG, "Sync cycle failed: ${e.message}", e)
                false
            }
        }
    }

    private suspend fun flushOutbox() {
        while (true) {
            val op = outboxDao.peekNext() ?: return
            val ok = try {
                processOp(op)
            } catch (e: Exception) {
                telemetryManager.logError(TAG, "Outbox op ${op.opType} ${op.entityType} failed: ${e.message}", e)
                outboxDao.update(op.copy(attempts = op.attempts + 1, lastError = e.message))
                return
            }
            if (!ok) {
                // Non-retriable failure already recorded; stop to avoid blocking forever.
                return
            }
        }
    }

    private suspend fun processOp(op: OutboxEntity): Boolean {
        return when (op.entityType) {
            OutboxEntityType.TASK -> processTaskOp(op)
            OutboxEntityType.LABEL -> processLabelOp(op)
            else -> {
                outboxDao.deleteById(op.id)
                true
            }
        }
    }

    private suspend fun processTaskOp(op: OutboxEntity): Boolean {
        when (op.opType) {
            OutboxOpType.CREATE -> {
                val oldId = op.targetServerId ?: return consume(op)
                val current = taskDao.getTaskById(oldId) ?: run {
                    // Row was deleted before sync — drop this op.
                    outboxDao.deleteById(op.id)
                    return true
                }
                if (current.labels.any { it.id < 0 }) {
                    // Referenced label still pending creation; wait for it to sync first.
                    recordFailure(op, "Waiting for pending label creates")
                    return false
                }
                val payload = CreateTaskReq(
                    title = current.task.title,
                    nextDueDate = current.task.nextDueDate,
                    endDate = current.task.endDate,
                    isRolling = current.task.isRolling,
                    frequency = current.task.frequency,
                    notification = current.task.notification,
                    labels = current.labels.map { it.id },
                )
                val response = api.createTask(payload)
                if (!response.isSuccessful) {
                    recordFailure(op, "HTTP ${response.code()}")
                    return false
                }
                val newId = response.body()?.task ?: run {
                    recordFailure(op, "Empty create response")
                    return false
                }
                taskDao.remapId(oldId, newId)
                op.targetLocalId?.let { local ->
                    outboxDao.remapLocalIdToServer(OutboxEntityType.TASK, local, newId)
                }
                outboxDao.remapServerId(OutboxEntityType.TASK, oldId, newId)
                outboxDao.deleteById(op.id)
            }
            OutboxOpType.UPDATE -> {
                val id = op.targetServerId ?: return consume(op)
                if (id < 0) {
                    // Still a placeholder — preceding CREATE must run first; skip this cycle.
                    recordFailure(op, "Awaiting preceding CREATE")
                    return false
                }
                val current = taskDao.getTaskById(id) ?: run {
                    outboxDao.deleteById(op.id)
                    return true
                }
                if (current.labels.any { it.id < 0 }) {
                    recordFailure(op, "Waiting for pending label creates")
                    return false
                }
                val payload = UpdateTaskReq(
                    id = id,
                    title = current.task.title,
                    nextDueDate = current.task.nextDueDate,
                    endDate = current.task.endDate,
                    isRolling = current.task.isRolling,
                    frequency = current.task.frequency,
                    notification = current.task.notification,
                    labels = current.labels.map { it.id },
                )
                val response = api.updateTask(payload)
                if (!response.isSuccessful) {
                    recordFailure(op, "HTTP ${response.code()}")
                    return false
                }
                taskDao.setState(id, LocalState.SYNCED)
                outboxDao.deleteById(op.id)
            }
            OutboxOpType.DELETE -> {
                val id = op.targetServerId ?: return consume(op)
                if (id < 0) return consume(op)
                val response = api.deleteTask(id)
                if (!response.isSuccessful && response.code() != 404) {
                    recordFailure(op, "HTTP ${response.code()}")
                    return false
                }
                taskDao.deleteById(id)
                outboxDao.deleteById(op.id)
            }
            OutboxOpType.COMPLETE -> {
                val id = op.targetServerId ?: return consume(op)
                if (id < 0) {
                    recordFailure(op, "Awaiting preceding CREATE")
                    return false
                }
                val endRecurrence = op.payloadJson?.toBooleanStrictOrNull() ?: false
                val response = api.completeTask(id, endRecurrence)
                if (!response.isSuccessful) {
                    recordFailure(op, "HTTP ${response.code()}")
                    return false
                }
                taskDao.setState(id, LocalState.SYNCED)
                outboxDao.deleteById(op.id)
            }
            OutboxOpType.UNCOMPLETE -> {
                val id = op.targetServerId ?: return consume(op)
                if (id < 0) {
                    recordFailure(op, "Awaiting preceding CREATE")
                    return false
                }
                val response = api.uncompleteTask(id)
                if (!response.isSuccessful) {
                    recordFailure(op, "HTTP ${response.code()}")
                    return false
                }
                taskDao.setState(id, LocalState.SYNCED)
                outboxDao.deleteById(op.id)
            }
            OutboxOpType.SKIP -> {
                val id = op.targetServerId ?: return consume(op)
                if (id < 0) {
                    recordFailure(op, "Awaiting preceding CREATE")
                    return false
                }
                val response = api.skipTask(id)
                if (!response.isSuccessful) {
                    recordFailure(op, "HTTP ${response.code()}")
                    return false
                }
                taskDao.setState(id, LocalState.SYNCED)
                outboxDao.deleteById(op.id)
            }
            OutboxOpType.DUE_DATE -> {
                val id = op.targetServerId ?: return consume(op)
                if (id < 0) {
                    recordFailure(op, "Awaiting preceding CREATE")
                    return false
                }
                val dueDate = op.payloadJson ?: return consume(op)
                val response = api.updateDueDate(id, UpdateDueDateReq(dueDate))
                if (!response.isSuccessful) {
                    recordFailure(op, "HTTP ${response.code()}")
                    return false
                }
                taskDao.setState(id, LocalState.SYNCED)
                outboxDao.deleteById(op.id)
            }
        }
        return true
    }

    private suspend fun processLabelOp(op: OutboxEntity): Boolean {
        when (op.opType) {
            OutboxOpType.CREATE -> {
                val oldId = op.targetServerId ?: return consume(op)
                val current = labelDao.getById(oldId) ?: run {
                    outboxDao.deleteById(op.id)
                    return true
                }
                val payload = CreateLabelReq(current.name, current.color)
                val response = api.createLabel(payload)
                if (!response.isSuccessful) {
                    recordFailure(op, "HTTP ${response.code()}")
                    return false
                }
                val newId = response.body()?.label ?: run {
                    recordFailure(op, "Empty create response")
                    return false
                }
                labelDao.remapId(oldId, newId)
                op.targetLocalId?.let { local ->
                    outboxDao.remapLocalIdToServer(OutboxEntityType.LABEL, local, newId)
                }
                outboxDao.remapServerId(OutboxEntityType.LABEL, oldId, newId)
                outboxDao.deleteById(op.id)
            }
            OutboxOpType.UPDATE -> {
                val id = op.targetServerId ?: return consume(op)
                if (id < 0) {
                    recordFailure(op, "Awaiting preceding CREATE")
                    return false
                }
                val current = labelDao.getById(id) ?: run {
                    outboxDao.deleteById(op.id)
                    return true
                }
                val response = api.updateLabel(UpdateLabelReq(id, current.name, current.color))
                if (!response.isSuccessful) {
                    recordFailure(op, "HTTP ${response.code()}")
                    return false
                }
                labelDao.upsert(current.copy(localState = LocalState.SYNCED))
                outboxDao.deleteById(op.id)
            }
            OutboxOpType.DELETE -> {
                val id = op.targetServerId ?: return consume(op)
                if (id < 0) return consume(op)
                val response = api.deleteLabel(id)
                if (!response.isSuccessful && response.code() != 404) {
                    recordFailure(op, "HTTP ${response.code()}")
                    return false
                }
                labelDao.deleteById(id)
                outboxDao.deleteById(op.id)
            }
            else -> {
                outboxDao.deleteById(op.id)
            }
        }
        return true
    }

    private suspend fun consume(op: OutboxEntity): Boolean {
        outboxDao.deleteById(op.id)
        return true
    }

    private suspend fun recordFailure(op: OutboxEntity, message: String) {
        outboxDao.update(op.copy(attempts = op.attempts + 1, lastError = message))
        telemetryManager.logWarning(TAG, "Outbox op ${op.opType} ${op.entityType} id=${op.id}: $message")
    }

    private suspend fun refreshAll() {
        refreshLabels()
        refreshTasks()
    }

    private suspend fun refreshTasks() {
        val response = api.getTasks()
        if (!response.isSuccessful) {
            telemetryManager.logWarning(TAG, "Refresh tasks failed: HTTP ${response.code()}")
            return
        }
        val tasks = response.body()?.tasks ?: emptyList()
        val dirtyIds = taskDao.dirtyIds().toSet()
        val serverIds = tasks.map { it.id }.toSet()

        taskDao.allIds()
            .filter { it !in serverIds && it !in dirtyIds }
            .forEach { taskDao.deleteById(it) }

        for (t in tasks) {
            if (t.id in dirtyIds) continue
            taskDao.upsert(t.toEntity(localState = LocalState.SYNCED))
            taskDao.replaceLabels(t.id, t.labels.map { it.id })
        }
    }

    private suspend fun refreshLabels() {
        val response = api.getLabels()
        if (!response.isSuccessful) {
            telemetryManager.logWarning(TAG, "Refresh labels failed: HTTP ${response.code()}")
            return
        }
        val labels = response.body()?.labels ?: emptyList()
        val dirtyIds = labelDao.dirtyIds().toSet()
        val serverIds = labels.map { it.id }.toSet()
        labelDao.allIds()
            .filter { it !in serverIds && it !in dirtyIds }
            .forEach { labelDao.deleteById(it) }
        for (l in labels) {
            if (l.id in dirtyIds) continue
            labelDao.upsert(
                LabelEntity(
                    id = l.id,
                    localId = null,
                    name = l.name,
                    color = l.color,
                    createdAt = l.createdAt,
                    updatedAt = l.updatedAt,
                    localState = LocalState.SYNCED,
                )
            )
        }
    }

    companion object {
        private const val TAG = "SyncCoordinator"
    }
}
