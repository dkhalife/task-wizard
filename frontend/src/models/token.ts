export interface APIToken {
  id: number
  name: string
  token: string
  expires_at: string
  scopes?: string[]
}

export type ApiTokenScope =
  | 'task:read'
  | 'task:write'
  | 'label:read'
  | 'label:write'
  | 'user:read'
  | 'user:write'
  | 'token:write'
  | 'dav:read'
  | 'dav:write'
