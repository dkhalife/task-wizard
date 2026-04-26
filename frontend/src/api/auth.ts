import { acquireAccessToken } from '@/utils/msal'

export interface AuthConfig {
  enabled: boolean
  tenant_id: string
  client_id: string
  audience: string
}

export const GetAuthConfig = async () => {
  const API_URL = import.meta.env.VITE_APP_API_URL
  const response = await fetch(`${API_URL}/api/v1/auth/config`, { credentials: 'include' })
  return (await response.json()) as AuthConfig
}

export const CreateSession = async () => {
  const API_URL = import.meta.env.VITE_APP_API_URL
  const headers: HeadersInit = { 'Content-Type': 'application/json' }

  try {
    const token = await acquireAccessToken()
    if (token) {
      headers['Authorization'] = 'Bearer ' + token
    }
  } catch {
    return
  }

  const response = await fetch(`${API_URL}/api/v1/auth/session`, {
    method: 'POST',
    headers,
    credentials: 'include',
  })

  if (!response.ok) {
    throw new Error('Failed to create session')
  }
}

export const DeleteSession = async () => {
  const API_URL = import.meta.env.VITE_APP_API_URL
  await fetch(`${API_URL}/api/v1/auth/session`, { method: 'DELETE', credentials: 'include' })
}
