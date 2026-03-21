package com.dkhalife.tasks.model

data class TasksResponse(val tasks: List<Task>)
data class TaskResponse(val task: Task)
data class TaskCreatedResponse(val task: Int)
data class TaskHistoryResponse(val history: List<TaskHistory>)
data class LabelsResponse(val labels: List<Label>)
data class LabelCreatedResponse(val label: Int)
data class UserResponse(val user: UserProfile)
data class ErrorResponse(val error: String)
