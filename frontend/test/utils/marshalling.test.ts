import { MakeActivityUI } from '@/utils/marshalling'
import { ActivityEntry } from '@/models/activity'

describe('marshalling', () => {
  describe('MakeActivityUI', () => {
    it('converts date strings to Date objects for a completion', () => {
      const entry: ActivityEntry = {
        id: 5,
        task_id: 2,
        task_title: 'Water plants',
        completed_date: '2024-01-02T10:00:00.000Z',
        due_date: '2024-01-01T09:00:00.000Z',
        is_latest: true,
      }

      const ui = MakeActivityUI(entry)

      expect(ui.id).toBe(5)
      expect(ui.task_id).toBe(2)
      expect(ui.task_title).toBe('Water plants')
      expect(ui.is_latest).toBe(true)
      expect(ui.completed_date).toBeInstanceOf(Date)
      expect(ui.due_date).toBeInstanceOf(Date)
      expect(ui.completed_date?.toISOString()).toBe('2024-01-02T10:00:00.000Z')
      expect(ui.due_date?.toISOString()).toBe('2024-01-01T09:00:00.000Z')
    })

    it('keeps null completed_date for a skip', () => {
      const entry: ActivityEntry = {
        id: 6,
        task_id: 3,
        task_title: 'Take out trash',
        completed_date: null,
        due_date: '2024-01-01T09:00:00.000Z',
        is_latest: false,
      }

      const ui = MakeActivityUI(entry)

      expect(ui.completed_date).toBeNull()
      expect(ui.due_date).toBeInstanceOf(Date)
      expect(ui.is_latest).toBe(false)
    })

    it('handles a missing due date', () => {
      const entry: ActivityEntry = {
        id: 7,
        task_id: 4,
        task_title: 'One off task',
        completed_date: '2024-01-02T10:00:00.000Z',
        due_date: null,
        is_latest: true,
      }

      const ui = MakeActivityUI(entry)

      expect(ui.due_date).toBeNull()
      expect(ui.completed_date).toBeInstanceOf(Date)
    })
  })
})
