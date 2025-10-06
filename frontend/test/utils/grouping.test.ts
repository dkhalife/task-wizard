import { getDefaultExpandedState, GROUP_BY } from '@/utils/grouping'
import { Label } from '@/models/label'

describe('grouping utils', () => {
  describe('getDefaultExpandedState', () => {
    describe('when groupBy is "due_date"', () => {
      it('should return default expanded state for due date groups', () => {
        const result = getDefaultExpandedState('due_date', [])
        expect(result).toEqual({
          overdue: false,
          today: false,
          tomorrow: false,
          this_week: false,
          next_week: false,
          later: false,
          any_time: false,
        })
      })

      it('should return same structure regardless of labels', () => {
        const labels: Label[] = [
          { id: 1, name: 'Label 1', color: '#ff0000' },
          { id: 2, name: 'Label 2', color: '#00ff00' },
        ]
        const result = getDefaultExpandedState('due_date', labels)
        expect(result).toEqual({
          overdue: false,
          today: false,
          tomorrow: false,
          this_week: false,
          next_week: false,
          later: false,
          any_time: false,
        })
      })
    })

    describe('when groupBy is "labels"', () => {
      it('should return empty object for empty labels array', () => {
        const result = getDefaultExpandedState('labels', [])
        expect(result).toEqual({})
      })

      it('should return expanded state for single label', () => {
        const labels: Label[] = [
          { id: 1, name: 'Label 1', color: '#ff0000' },
        ]
        const result = getDefaultExpandedState('labels', labels)
        expect(result).toEqual({
          '1': false,
        })
      })

      it('should return expanded state for multiple labels', () => {
        const labels: Label[] = [
          { id: 1, name: 'Label 1', color: '#ff0000' },
          { id: 2, name: 'Label 2', color: '#00ff00' },
          { id: 3, name: 'Label 3', color: '#0000ff' },
        ]
        const result = getDefaultExpandedState('labels', labels)
        expect(result).toEqual({
          '1': false,
          '2': false,
          '3': false,
        })
      })

      it('should handle labels with different IDs', () => {
        const labels: Label[] = [
          { id: 10, name: 'Label A', color: '#ff0000' },
          { id: 25, name: 'Label B', color: '#00ff00' },
          { id: 99, name: 'Label C', color: '#0000ff' },
        ]
        const result = getDefaultExpandedState('labels', labels)
        expect(result).toEqual({
          '10': false,
          '25': false,
          '99': false,
        })
      })

      it('should set all labels to collapsed (false) by default', () => {
        const labels: Label[] = [
          { id: 1, name: 'Label 1', color: '#ff0000' },
          { id: 2, name: 'Label 2', color: '#00ff00' },
        ]
        const result = getDefaultExpandedState('labels', labels)
        Object.values(result).forEach(value => {
          expect(value).toBe(false)
        })
      })
    })

    describe('edge cases', () => {
      it('should handle invalid groupBy gracefully', () => {
        // TypeScript would prevent this, but in runtime it would fall through
        const result = getDefaultExpandedState('invalid' as GROUP_BY, [])
        expect(result).toEqual({})
      })
    })
  })
})
