package com.dkhalife.tasks.data.sync

import android.content.Context
import com.dkhalife.tasks.model.Task

interface SyncEngine {
    suspend fun sync(context: Context, tasks: List<Task>)
}
