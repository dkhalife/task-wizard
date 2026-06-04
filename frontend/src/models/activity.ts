export interface ActivityEntry {
  id: number
  task_id: number
  task_title: string
  completed_date: string | null
  due_date: string | null
  is_latest: boolean
}
