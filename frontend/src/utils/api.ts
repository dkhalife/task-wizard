import { getAccessToken, loginSilently, logout } from "./msal"

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
  let accessToken: string | null = null
  if (requiresAuth) {
    try {
      accessToken = getAccessToken();
    } catch {
      try {
        await loginSilently();
        accessToken = getAccessToken();
      } catch (error) {
        await logout();
        throw error;
      }
    }
  }

  const fullURL = `${API_URL}/api/v1${url}`

  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    'Cache-Control': 'no-store',
  }

  if (requiresAuth) {
    headers['Authorization'] = 'Bearer ' + accessToken
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
