export interface AuthConfig {
  enabled: boolean
  tenant_id: string
  client_id: string
  audience: string
}

export const GetAuthConfig = async () => {
  const API_URL = import.meta.env.VITE_APP_API_URL
  const response = await fetch(`${API_URL}/api/v1/auth/config`)
  return (await response.json()) as AuthConfig
}
