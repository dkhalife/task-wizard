package app.taskwiz.data.sync

import android.content.Context
import app.taskwiz.model.Task

interface SyncEngine {
    suspend fun sync(context: Context, tasks: List<Task>)
}
