export type StatusSeverity = 'error' | 'success' | 'info' | 'warning'

export type Status = {
  message: string
  severity: StatusSeverity
  timeout?: number
}

export type TrackedStatus = Status & {
  id: string
  createdAt: number
}
