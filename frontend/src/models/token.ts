export interface APIToken {
  id: string
  name: string
  token: string
  expires_at: string
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
