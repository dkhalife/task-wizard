import { Task } from '@/models/task'
import { Request } from '../utils/api'
import { HistoryEntry } from '@/models/history'
import { MarshallLabels } from '@/utils/marshalling'
import { transport } from './transport'

type TaskIdResponse = {
  task: number
}

type SingleTaskResponse = {
  task: Task
}

type TasksResponse = {
  tasks: Task[]
}

type TaskHistoryResponse = {
  history: HistoryEntry[]
}

export const GetTasks = async (): Promise<TasksResponse> =>
  await transport({
    http: () => Request<TasksResponse>(`/tasks/`),
    ws: (ws) => ws.request<'get_tasks', TasksResponse>('get_tasks'),
  })

export const GetCompletedTasks = async (): Promise<TasksResponse> =>
  await transport({
    http: () => Request<TasksResponse>(`/tasks/completed`),
    // WS handler requires data to be present (at least {}).
    ws: (ws) => ws.request<'get_completed_tasks', TasksResponse>('get_completed_tasks', {}),
  })

export const MarkTaskComplete = async (id: number, endRecurrence: boolean): Promise<SingleTaskResponse> =>
    await transport({
      http: () => Request<SingleTaskResponse>(`/tasks/${id}/do?endRecurrence=${endRecurrence}`, 'POST'),
      ws: (ws) => ws.request<'complete_task', SingleTaskResponse>('complete_task', { id, endRecurrence }),
    })

export const SkipTask = async (id: number): Promise<SingleTaskResponse> =>
    await transport({
      http: () => Request<SingleTaskResponse>(`/tasks/${id}/skip`, 'POST'),
      ws: (ws) => ws.request<'skip_task', SingleTaskResponse>('skip_task', id),
    })

export const CreateTask = async (task: Omit<Task, 'id'>) =>
  await transport({
    http: () => Request<TaskIdResponse>(`/tasks/`, 'POST', MarshallLabels(task)),
    ws: (ws) => ws.request<'create_task', TaskIdResponse>('create_task', MarshallLabels(task)),
  })

export const DeleteTask = async (id: number) =>
  await transport({
    http: () => Request<void>(`/tasks/${id}`, 'DELETE'),
    ws: (ws) => ws.request<'delete_task', void>('delete_task', id),
  })

export const SaveTask = async (task: Task) =>
  await transport({
    http: () => Request<void>(`/tasks/`, 'PUT', MarshallLabels(task)),
    ws: (ws) => ws.request<'update_task', void>('update_task', MarshallLabels(task)),
  })

export const GetTaskHistory = async (taskId: number) =>
  await transport({
    http: () => Request<TaskHistoryResponse>(`/tasks/${taskId}/history`),
    ws: (ws) => ws.request<'get_task_history', TaskHistoryResponse>('get_task_history', taskId),
  })

export const UpdateDueDate = async (
  id: number,
  due_date: string,
): Promise<SingleTaskResponse> =>
  await transport({
    http: () =>
      Request<SingleTaskResponse>(`/tasks/${id}/dueDate`, 'PUT', {
        due_date,
      }),
    ws: (ws) =>
      ws.request<'update_due_date', SingleTaskResponse>('update_due_date', {
        id,
        due_date,
      }),
  })
