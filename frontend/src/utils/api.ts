import { acquireAccessToken, isAuthEnabled } from '@/utils/msal'

const API_URL = import.meta.env.VITE_APP_API_URL

type RequestMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'

type FailureResponse = {
  error: string
}

export async function Request<SuccessfulResponse>(
  url: string,
  method: RequestMethod = 'GET',
  body: unknown = {},
  requiresAuth: boolean = true,
): Promise<SuccessfulResponse> {
  const fullURL = `${API_URL}/api/v1${url}`

  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    'Cache-Control': 'no-store',
  }

  if (requiresAuth && isAuthEnabled()) {
    const token = await acquireAccessToken()
    headers['Authorization'] = 'Bearer ' + token
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
