import { PayloadAction, createAsyncThunk, createSlice } from '@reduxjs/toolkit'
import {
  GetTasks,
  MarkTaskComplete,
  DeleteTask,
  SkipTask,
  CreateTask,
  SaveTask,
  UpdateDueDate,
} from '@/api/tasks'
import { newTask, Task } from '@/models/task'
import { RootState, store } from './store'
import { SyncState } from '@/models/sync'
import { WSEventPayloads } from '@/models/websocket'
import {
  getDefaultExpandedState,
  GROUP_BY,
  groupTaskBy,
  groupTasksBy,
  sortTasksByDueDate,
  TaskGroups,
} from '@/utils/grouping'
import { retrieveValue, storeValue } from '@/utils/storage'
import WebSocketManager from '@/utils/websocket'

export interface TasksState {
  draft: Task
  items: Task[]

  searchQuery: string
  filteredItems: Task[]

  groupBy: GROUP_BY
  groupedItems: TaskGroups<Task>
  expandedGroups: Record<keyof TaskGroups<Task>, boolean>

  status: SyncState
  lastFetched: number | null
  error: string | null
}

const initialGroupBy = retrieveValue<GROUP_BY>('group_by', 'due_date')
const initialExpandedGroups = retrieveValue<Record<keyof TaskGroups<Task>, boolean>>(
  'expanded_groups',
  getDefaultExpandedState(initialGroupBy, []),
)

const initialState: TasksState = {
  draft: newTask(),
  items: [],

  searchQuery: '',
  filteredItems: [],

  groupBy: initialGroupBy,
  expandedGroups: initialExpandedGroups,
  groupedItems: {},

  status: 'loading',
  lastFetched: null,
  error: null,
}

export const fetchTasks = createAsyncThunk('tasks/fetchTasks', async () => {
  const data = await GetTasks()
  return data.tasks
})

export const completeTask = createAsyncThunk(
  'tasks/completeTask',
  async (taskId: number) => {
    const response = await MarkTaskComplete(taskId)
    return response.task
  },
)

export const skipTask = createAsyncThunk(
  'tasks/skipTask',
  async (taskId: number) => {
    const response = await SkipTask(taskId)
    return response.task
  },
)

export const deleteTask = createAsyncThunk(
  'tasks/deleteTask',
  async (taskId: number) => await DeleteTask(taskId),
)

export const createTask = createAsyncThunk(
  'tasks/createTask',
  async (task: Omit<Task, 'id'>) => await CreateTask(task),
)

export const saveTask = createAsyncThunk(
  'tasks/saveTask',
  async (task: Task) => await SaveTask(task),
)

export const updateDueDate = createAsyncThunk(
  'tasks/updateDueDate',
  async ({ taskId, dueDate }: { taskId: number; dueDate: string }) => {
    const response = await UpdateDueDate(taskId, dueDate)
    return response.task
  },
)

export const setGroupBy = createAsyncThunk(
  'tasks/setGroupBy',
  async (groupBy: GROUP_BY, thunkAPI) => {
      const state = thunkAPI.getState() as RootState

      const userLabels = state.labels.items
      const tasks = state.tasks.items
      const nextExpanded = getDefaultExpandedState(groupBy, userLabels)
      const groupedItems = groupTasksBy(tasks, userLabels, groupBy)

      storeValue('group_by', groupBy)
      storeValue('expanded_groups', nextExpanded)

      return {
        groupBy,
        groupedItems: groupedItems,
        expandedGroups: nextExpanded,
      }
  },
)

export const initGroups = createAsyncThunk(
  'tasks/initGroups',
  async (_, thunkAPI) => {
    const state = thunkAPI.getState() as RootState
    const userLabels = state.labels.items
    const tasks = state.tasks.items

    return groupTasksBy(tasks, userLabels, initialGroupBy)
  },
)

function taskMatchesQuery(task: Task, query: string): boolean {
  return task.title.toLowerCase().includes(query)
}

function filterItems(items: Task[], query: string): Task[] {
  if (query === '') {
    return items
  }

  const lowerQuery = query.toLowerCase()
  return items.filter(task => taskMatchesQuery(task, lowerQuery))
}

function deleteTaskFromGroups(taskId: number, groups: TaskGroups<Task>) {
  const keys = Object.keys(groups) as (keyof TaskGroups<Task>)[]

  keys.forEach(groupKey => {
    const group = groups[groupKey]
    group.content = group.content.filter(t => t.id !== taskId)
  })
}

const tasksSlice = createSlice({
  name: 'tasks',
  initialState,
  reducers: {
    setDraft: (state, action: PayloadAction<Task>) => {
      state.draft = action.payload
    },
    filterTasks: (state, action: PayloadAction<string>) => {
      state.searchQuery = action.payload
      state.filteredItems = filterItems(state.items, action.payload)
    },
    toggleGroup: (state, action: PayloadAction<keyof TaskGroups<Task>>) => {
      const groupKey = action.payload
      const isExpanded = state.expandedGroups[groupKey]
      state.expandedGroups[groupKey] = !isExpanded

      storeValue('expanded_groups', state.expandedGroups)
    },
    taskUpserted: (state, action: PayloadAction<Task>) => {
      const task = action.payload

      const index = state.items.findIndex(t => t.id === task.id)
      if (index >= 0) {
        state.items[index] = task
      } else {
        state.items.push(task)
      }

      const filteredIndex = state.filteredItems.findIndex(t => t.id === task.id)
      const matchesQuery = state.searchQuery === '' || taskMatchesQuery(task, state.searchQuery.toLowerCase())

      if (filteredIndex >= 0) {
        if (matchesQuery) {
          state.filteredItems[filteredIndex] = task
        } else {
          state.filteredItems.splice(filteredIndex, 1)
        }
      } else if (matchesQuery) {
        state.filteredItems.push(task)
      }

      deleteTaskFromGroups(task.id, state.groupedItems)
      groupTaskBy(task, state.groupedItems, state.groupBy)
      sortTasksByDueDate(state.items)
      sortTasksByDueDate(state.filteredItems)
    },
    taskDeleted: (state, action: PayloadAction<number>) => {
      const taskId = action.payload
      state.items = state.items.filter(t => t.id !== taskId)
      state.filteredItems = state.filteredItems.filter(t => t.id !== taskId)
      deleteTaskFromGroups(taskId, state.groupedItems)
    },
  },
  extraReducers: builder => {
    builder
      // Fetch tasks
      .addCase(fetchTasks.pending, state => {
        state.status = 'loading'
        state.error = null
      })
      .addCase(fetchTasks.fulfilled, (state, action) => {
        state.status = 'succeeded'
        state.items = action.payload
        state.filteredItems = filterItems(action.payload, state.searchQuery)
        state.lastFetched = Date.now()
        state.error = null
      })
      .addCase(fetchTasks.rejected, (state, action) => {
        state.status = 'failed'
        state.error = action.error.message ?? null
      })
      // Initialize groups
      .addCase(initGroups.pending, state => {
        state.status = 'loading'
        state.error = null
      })
      .addCase(initGroups.fulfilled, (state, action) => {
        state.groupedItems = action.payload
        state.status = 'succeeded'
        state.error = null
      })
      .addCase(initGroups.rejected, (state, action) => {
        state.status = 'failed'
        state.error = action.error.message ?? null
      })
      // Change group by
      .addCase(setGroupBy.pending, state => {
        state.status = 'loading'
        state.error = null
      })
      .addCase(setGroupBy.fulfilled, (state, action) => {
        state.groupBy = action.payload.groupBy
        state.groupedItems = action.payload.groupedItems
        state.expandedGroups = action.payload.expandedGroups
        state.status = 'succeeded'
        state.error = null
      })
      .addCase(setGroupBy.rejected, (state, action) => {
        state.status = 'failed'
        state.error = action.error.message ?? null
      })
      // Create tasks
      .addCase(createTask.pending, state => {
        state.status = 'loading'
      })
      .addCase(createTask.fulfilled, (state, action) => {
        state.status = 'succeeded'

        const taskId = action.payload.task
        const task: Task = {
          ...action.meta.arg,
          id: taskId,
        }

        tasksSlice.caseReducers.taskUpserted(state, {
          payload: task,
          type: 'tasks/taskUpserted',
        })

        state.draft = newTask()

        state.error = null
      })
      .addCase(createTask.rejected, (state, action) => {
        state.status = 'failed'
        state.error = action.error.message ?? null
      })
      // Edit tasks
      .addCase(saveTask.pending, state => {
        state.status = 'loading'
      })
      .addCase(saveTask.fulfilled, (state, action) => {
        state.status = 'succeeded'

        const updatedTask = action.meta.arg
        tasksSlice.caseReducers.taskUpserted(state, {
          payload: updatedTask,
          type: 'tasks/taskUpserted',
        })

        state.draft = newTask()

        state.error = null
      })
      .addCase(saveTask.rejected, (state, action) => {
        state.status = 'failed'
        state.error = action.error.message ?? null
      })
      // Mark tasks as complete
      .addCase(completeTask.pending, state => {
        state.status = 'loading'
        state.error = null
      })
      .addCase(completeTask.fulfilled, (state, action) => {

        // For recurring tasks, we might have a new task with the same ID but updated details
        const newTask = action.payload
        if (newTask.next_due_date) {
          tasksSlice.caseReducers.taskUpserted(state, {
            payload: newTask,
            type: 'tasks/taskUpserted',
          })
        } else {
          tasksSlice.caseReducers.taskDeleted(state, {
            payload: newTask.id,
            type: 'tasks/taskDeleted',
          })
        }

        state.status = 'succeeded'
        state.error = null
      })
      .addCase(completeTask.rejected, (state, action) => {
        state.status = 'failed'
        state.error = action.error.message ?? null
      })
      // Skip tasks
      .addCase(skipTask.pending, state => {
        state.status = 'loading'
        state.error = null
      })
      .addCase(skipTask.fulfilled, (state, action) => {
        const newTask = action.payload
        if (newTask.next_due_date) {
          tasksSlice.caseReducers.taskUpserted(state, {
            payload: newTask,
            type: 'tasks/taskUpserted',
          })
        }

        state.status = 'succeeded'
        state.error = null
      })
      .addCase(skipTask.rejected, (state, action) => {
        state.status = 'failed'
        state.error = action.error.message ?? null
      })
      // Update due date
      .addCase(updateDueDate.pending, state => {
        state.status = 'loading'
        state.error = null
      })
      .addCase(updateDueDate.fulfilled, (state, action) => {
        const updatedTask = action.payload

        tasksSlice.caseReducers.taskUpserted(state, {
          payload: updatedTask,
          type: 'tasks/taskUpserted',
        })

        state.status = 'succeeded'
        state.error = null
      })
      .addCase(updateDueDate.rejected, (state, action) => {
        state.status = 'failed'
        state.error = action.error.message ?? null
      })
      // Deleting tasks
      .addCase(deleteTask.pending, state => {
        state.status = 'loading'
      })
      .addCase(deleteTask.fulfilled, (state, action) => {
        state.status = 'succeeded'
        tasksSlice.caseReducers.taskDeleted(state, {
          payload: action.meta.arg,
          type: 'tasks/taskDeleted',
        })
        state.error = null
      })
      .addCase(deleteTask.rejected, (state, action) => {
        state.status = 'failed'
        state.error = action.error.message ?? null
      })
  },
})

export const { setDraft, filterTasks, toggleGroup } = tasksSlice.actions

export const tasksReducer = tasksSlice.reducer

const { taskUpserted, taskDeleted } = tasksSlice.actions

const onTaskCreated = (data: WSEventPayloads['task_created']) => {
  store.dispatch(taskUpserted(data))
}

const onTaskUpdated = (data: WSEventPayloads['task_updated']) => {
  store.dispatch(taskUpserted(data))
}

const onTaskDeletedEvent = (data: WSEventPayloads['task_deleted']) => {
  store.dispatch(taskDeleted(data.id))
}

const onTaskCompleted = (data: WSEventPayloads['task_completed']) => {
  if (data.next_due_date) {
    store.dispatch(taskUpserted(data))
  } else {
    store.dispatch(taskDeleted(data.id))
  }
}

const onTaskUncompleted = (data: WSEventPayloads['task_uncompleted']) => {
  store.dispatch(taskUpserted(data))
}

const onTaskSkipped = (data: WSEventPayloads['task_skipped']) => {
  if (data.next_due_date) {
    store.dispatch(taskUpserted(data))
  } else {
    store.dispatch(taskDeleted(data.id))
  }
}

const onNotification = (data: WSEventPayloads['notification']) => {
  const enabled = retrieveValue<boolean>('desktop_notifications', false)
  if (!enabled || Notification.permission !== 'granted') {
    return
  }

  const state = store.getState()
  const task = state.tasks.items.find(t => t.id === data.task_id)
  if (!task) {
    return
  }

  let body = ''
  switch (data.type) {
    case 'pre_due':
      body = 'Task is due soon'
      break
    case 'overdue':
      body = 'Task is overdue'
      break
    case 'due_date':
    default:
      body = 'Task is due'
      break
  }

  try {
    // Sanitize task.title to prevent XSS
    const safeTitle = task.title.replace(/</g, "&lt;").replace(/>/g, "&gt;");
    new Notification(safeTitle, { body })
  } catch (e) {
    console.debug('Failed to show notification', e)
  }
}

export const registerWebSocketListeners = (ws: WebSocketManager) => {
  ws.on('task_created', onTaskCreated)
  ws.on('task_updated', onTaskUpdated)
  ws.on('task_deleted', onTaskDeletedEvent)
  ws.on('task_completed', onTaskCompleted)
  ws.on('task_uncompleted', onTaskUncompleted)
  ws.on('task_skipped', onTaskSkipped)
  ws.on('notification', onNotification)
}

export const unregisterWebSocketListeners = (ws: WebSocketManager) => {
  ws.off('task_created', onTaskCreated)
  ws.off('task_updated', onTaskUpdated)
  ws.off('task_deleted', onTaskDeletedEvent)
  ws.off('task_completed', onTaskCompleted)
  ws.off('task_uncompleted', onTaskUncompleted)
  ws.off('task_skipped', onTaskSkipped)
  ws.off('notification', onNotification)
}
