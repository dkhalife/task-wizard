package com.dkhalife.tasks.repo

import com.dkhalife.tasks.api.TaskWizardApi
import com.dkhalife.tasks.model.*
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TaskRepository @Inject constructor(
    private val api: TaskWizardApi
) {
    private val _tasks = MutableStateFlow<List<Task>>(emptyList())
    val tasks: StateFlow<List<Task>> = _tasks

    private val _completedTasks = MutableStateFlow<List<Task>>(emptyList())
    val completedTasks: StateFlow<List<Task>> = _completedTasks

    suspend fun refreshTasks(): Result<List<Task>> {
        return try {
            val response = api.getTasks()
            if (response.isSuccessful) {
                val tasks = response.body()?.tasks ?: emptyList()
                _tasks.value = tasks
                Result.success(tasks)
            } else {
                Result.failure(Exception("Failed to fetch tasks: ${response.code()}"))
            }
        } catch (e: Exception) {
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
                Result.failure(Exception("Failed to fetch completed tasks: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getTask(id: Int): Result<Task> {
        return try {
            val response = api.getTask(id)
            if (response.isSuccessful) {
                Result.success(response.body()!!.task)
            } else {
                Result.failure(Exception("Failed to fetch task: ${response.code()}"))
            }
        } catch (e: Exception) {
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
            Result.failure(e)
        }
    }

    suspend fun createTask(req: CreateTaskReq): Result<Int> {
        return try {
            val response = api.createTask(req)
            if (response.isSuccessful) {
                refreshTasks()
                Result.success(response.body()!!.task)
            } else {
                Result.failure(Exception("Failed to create task: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun updateTask(req: UpdateTaskReq): Result<Unit> {
        return try {
            val response = api.updateTask(req)
            if (response.isSuccessful) {
                refreshTasks()
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to update task: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun deleteTask(id: Int): Result<Unit> {
        return try {
            val response = api.deleteTask(id)
            if (response.isSuccessful) {
                refreshTasks()
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to delete task: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun completeTask(id: Int, endRecurrence: Boolean = false): Result<Unit> {
        return try {
            val response = api.completeTask(id, endRecurrence)
            if (response.isSuccessful) {
                refreshTasks()
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to complete task: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun uncompleteTask(id: Int): Result<Unit> {
        return try {
            val response = api.uncompleteTask(id)
            if (response.isSuccessful) {
                refreshTasks()
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to uncomplete task: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun skipTask(id: Int): Result<Unit> {
        return try {
            val response = api.skipTask(id)
            if (response.isSuccessful) {
                refreshTasks()
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to skip task: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun updateDueDate(id: Int, dueDate: String): Result<Unit> {
        return try {
            val response = api.updateDueDate(id, UpdateDueDateReq(dueDate))
            if (response.isSuccessful) {
                refreshTasks()
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to update due date: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    fun updateTasksFromWebSocket(tasks: List<Task>) {
        _tasks.value = tasks
    }
}
