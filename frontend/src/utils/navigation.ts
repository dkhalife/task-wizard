import { isMobile } from './dom'
import { retrieveValue } from './storage'

export const getPathName = () => document.location.pathname

export const getQuery = (key: string): string => {
  return new URLSearchParams(document.location.search).get(key) ?? ''
}

export const NavigationPaths = {
  Login: '/login',
  Register: '/signup',
  ResetPassword: '/forgot-password',
  MyTasks: '/my/tasks',
  TasksOverview: '/tasks',
  HomeView: () => (getHomeView() === 'my_tasks' ? '/my/tasks' : '/tasks'),
  Labels: '/labels',
  Settings: '/settings',
  TaskCreate: '/tasks/create',
  TaskEdit: (taskId: number) => `/tasks/${taskId}/edit`,
  TaskHistory: (taskId: number) => `/tasks/${taskId}/history`,
}

export interface WithNavigate {
  navigate: (path: string) => void
}

export type HomeView = 'my_tasks' | 'tasks_overview'

export const getHomeView = (): HomeView => {
  return retrieveValue<HomeView>(
    'home_view',
    isMobile() ? 'my_tasks' : 'tasks_overview',
  )
}
