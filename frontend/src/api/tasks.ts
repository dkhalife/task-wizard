import { Task } from '@/models/task'
import { Request } from '../utils/api'
import { HistoryEntry } from '@/models/history'
import { MarshallLabels } from '@/utils/marshalling'

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
  await Request<TasksResponse>(`/tasks/`)

export const MarkTaskComplete = async (id: number): Promise<SingleTaskResponse> =>
    await Request<SingleTaskResponse>(`/tasks/${id}/do`, 'POST')

export const SkipTask = async (id: number): Promise<SingleTaskResponse> =>
    await Request<SingleTaskResponse>(`/tasks/${id}/skip`, 'POST')

export const CreateTask = async (task: Omit<Task, 'id'>) =>
  await Request<TaskIdResponse>(`/tasks/`, 'POST', MarshallLabels(task as Task))

export const DeleteTask = async (id: number) =>
  await Request<void>(`/tasks/${id}`, 'DELETE')

export const SaveTask = async (task: Task) =>
  await Request<void>(`/tasks/`, 'PUT', MarshallLabels(task))

export const GetTaskHistory = async (taskId: number) =>
  await Request<TaskHistoryResponse>(`/tasks/${taskId}/history`)

export const UpdateDueDate = async (
  id: number,
  due_date: string,
): Promise<SingleTaskResponse> =>
  await Request<SingleTaskResponse>(`/tasks/${id}/dueDate`, 'PUT', {
    due_date
  })
