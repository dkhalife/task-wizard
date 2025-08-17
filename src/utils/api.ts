import { RefreshToken } from '@/api/auth'
import { TOKEN_REFRESH_THRESHOLD_MS } from '@/constants/time'

const API_URL = import.meta.env.VITE_APP_API_URL

type RequestMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'

type FailureResponse = {
  error: string
}

let isRefreshingAccessToken = false
const isTokenNearExpiration = () => {
  const now = new Date()
  const expiration = localStorage.getItem('ca_expiration') || ''
  const expire = new Date(expiration)
  return now.getTime() + TOKEN_REFRESH_THRESHOLD_MS > expire.getTime()
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

  const fullURL = `${API_URL}/api/v1${url}`

  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  }

  if (requiresAuth) {
    headers['Authorization'] = 'Bearer ' + localStorage.getItem('ca_token')
  }

  const options: RequestInit = {
    method,
    headers,
  }

  if (method !== 'GET') {
    options.body = JSON.stringify(body)
  }

  const response: Response = await fetch(fullURL, options)
  const HTTP_STATUS_NO_CONTENT = 204
  if (response.status === HTTP_STATUS_NO_CONTENT) {
    return {} as SuccessfulResponse
  }

  const data = await response.json()

  if (!response.ok) {
    throw new Error((data as FailureResponse).error)
  }

  return data as SuccessfulResponse
}
