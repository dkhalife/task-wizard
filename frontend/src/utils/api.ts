import { RefreshToken } from '@/api/auth'
import { TOKEN_REFRESH_THRESHOLD_MS } from '@/constants/time'
import WebSocketManager from './websocket'
import { store } from '@/store/store'
import { WSAction } from '@/models/websocket'

const API_URL = import.meta.env.VITE_APP_API_URL
const HTTP_STATUS_NO_CONTENT = 204

type RequestMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'

type FailureResponse = {
  error: string
}

interface RequestToWSMapping {
  action: WSAction
  getData: (url: string, body: unknown) => any
}

let isRefreshingAccessToken = false
const isTokenNearExpiration = () => {
  const now = new Date()
  const expiration = localStorage.getItem('ca_expiration') || ''
  const expire = new Date(expiration)
  return now.getTime() + TOKEN_REFRESH_THRESHOLD_MS > expire.getTime()
}

// Map HTTP requests to WebSocket actions
const mapRequestToWebSocket = (
  url: string,
  method: RequestMethod
): RequestToWSMapping | null => {
  // Extract ID from URL if present
  const idMatch = url.match(/\/(\d+)/)
  const id = idMatch ? parseInt(idMatch[1], 10) : undefined

  // Labels endpoints
  if (url === '/labels' && method === 'GET') {
    return { action: 'get_user_labels', getData: () => undefined }
  }
  if (url === '/labels' && method === 'POST') {
    return { action: 'create_label', getData: (_url, body) => body }
  }
  if (url === '/labels' && method === 'PUT') {
    return { action: 'update_label', getData: (_url, body) => body }
  }
  if (url.startsWith('/labels/') && method === 'DELETE' && id) {
    return { action: 'delete_label', getData: () => ({ id }) }
  }

  // Tasks endpoints
  if (url === '/tasks/' && method === 'GET') {
    return { action: 'get_tasks', getData: () => undefined }
  }
  if (url === '/tasks/completed' && method === 'GET') {
    return { action: 'get_completed_tasks', getData: () => ({}) }
  }
  if (url === '/tasks/' && method === 'POST') {
    return { action: 'create_task', getData: (_url, body) => body }
  }
  if (url === '/tasks/' && method === 'PUT') {
    return { action: 'update_task', getData: (_url, body) => body }
  }
  if (url.startsWith('/tasks/') && url.endsWith('/do') && method === 'POST' && id) {
    return { action: 'complete_task', getData: () => ({ id }) }
  }
  if (url.startsWith('/tasks/') && url.endsWith('/skip') && method === 'POST' && id) {
    return { action: 'skip_task', getData: () => ({ id }) }
  }
  if (url.startsWith('/tasks/') && url.endsWith('/dueDate') && method === 'PUT' && id) {
    return { action: 'update_due_date', getData: (_url, body: any) => ({ id, due_date: body.due_date }) }
  }
  if (url.startsWith('/tasks/') && url.endsWith('/history') && method === 'GET' && id) {
    return { action: 'get_task_history', getData: () => ({ id }) }
  }
  if (url.startsWith('/tasks/') && method === 'DELETE' && id) {
    return { action: 'delete_task', getData: () => ({ id }) }
  }

  // User/Token endpoints
  if (url === '/users/tokens' && method === 'GET') {
    return { action: 'get_app_tokens', getData: () => undefined }
  }
  if (url === '/users/tokens' && method === 'POST') {
    return { action: 'create_app_token', getData: (_url, body) => body }
  }
  if (url.startsWith('/users/tokens/') && method === 'DELETE') {
    const tokenId = url.split('/').pop()
    return { action: 'delete_app_token', getData: () => ({ id: tokenId }) }
  }
  if (url === '/users/profile' && method === 'GET') {
    return { action: 'get_user_profile', getData: () => undefined }
  }
  if (url === '/users/notifications' && method === 'PUT') {
    return { action: 'update_notification_settings', getData: (_url, body) => body }
  }
  if (url === '/users/change_password' && method === 'PUT') {
    return { action: 'update_password', getData: (_url, body) => body }
  }

  return null
}

export const isTokenValid = (): boolean => {
  const token = localStorage.getItem('ca_token')
  if (token) {
    const now = new Date()
    const expiration = localStorage.getItem('ca_expiration') || ''
    const expire = new Date(expiration)
    if (now < expire) {
      return true
    }
    localStorage.removeItem('ca_token')
    localStorage.removeItem('ca_expiration')
    return false
  }
  return false
}

export const refreshAccessToken = async () => {
  isRefreshingAccessToken = true
  try {
    const data = await RefreshToken()
    localStorage.setItem('ca_token', data.token)
    localStorage.setItem('ca_expiration', data.expiration)
  } catch (error) {
    localStorage.removeItem('ca_token')
    localStorage.removeItem('ca_expiration')
    localStorage.setItem('ca_redirect', window.location.pathname)
    window.location.href = '/login'
    console.error('Failed to refresh access token', error)
  } finally {
    isRefreshingAccessToken = false
  }
}

export async function Request<SuccessfulResponse>(
  url: string,
  method: RequestMethod = 'GET',
  body: unknown = {},
  requiresAuth: boolean = true,
): Promise<SuccessfulResponse> {
  if (isTokenValid()) {
    if (!isRefreshingAccessToken && isTokenNearExpiration()) {
      await refreshAccessToken()
    }
  } else if (requiresAuth) {
    localStorage.setItem('ca_redirect', window.location.pathname)
    window.location.href = '/login'

    throw new Error('User is not authenticated')
  }

  // Check if WebSocket should be used
  const state = store.getState()
  const useWebSocket = state.featureFlags.useWebsockets && requiresAuth
  
  if (useWebSocket) {
    const wsManager = WebSocketManager.getInstance()
    if (wsManager.isConnected()) {
      const wsMapping = mapRequestToWebSocket(url, method)
      
      if (wsMapping) {
        try {
          const wsData = wsMapping.getData(url, body)
          const response = await wsManager.sendRequest(wsMapping.action, wsData)
          
          if (response.status === HTTP_STATUS_NO_CONTENT) {
            return {} as SuccessfulResponse
          }

          if (response.status >= 400) {
            const errorData = response.data as FailureResponse
            throw new Error(errorData.error || 'Request failed')
          }

          return response.data as SuccessfulResponse
        } catch (error) {
          // If WebSocket request fails, fall back to HTTP
          console.debug('WebSocket request failed, falling back to HTTP:', error)
        }
      }
    }
  }

  // Fall back to HTTP request
  const fullURL = `${API_URL}/api/v1${url}`

  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    'Cache-Control': 'no-store',
  }

  if (requiresAuth) {
    headers['Authorization'] = 'Bearer ' + localStorage.getItem('ca_token')
  }

  const options: RequestInit = {
    method,
    headers,
    cache: 'no-store',
  }

  if (method !== 'GET') {
    options.body = JSON.stringify(body)
  }

  const response: Response = await fetch(fullURL, options)
  if (response.status === HTTP_STATUS_NO_CONTENT) {
    return {} as SuccessfulResponse
  }

  const data = await response.json()

  if (!response.ok) {
    throw new Error((data as FailureResponse).error)
  }

  return data as SuccessfulResponse
}
