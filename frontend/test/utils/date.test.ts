import { dayOfMonthSuffix } from '@/utils/date'

describe('date utils', () => {
  describe('dayOfMonthSuffix', () => {
    it('should return "st" for 1', () => {
      expect(dayOfMonthSuffix(1)).toBe('st')
    })

    it('should return "nd" for 2', () => {
      expect(dayOfMonthSuffix(2)).toBe('nd')
    })

    it('should return "rd" for 3', () => {
      expect(dayOfMonthSuffix(3)).toBe('rd')
    })

    it('should return "th" for 4-10', () => {
      expect(dayOfMonthSuffix(4)).toBe('th')
      expect(dayOfMonthSuffix(5)).toBe('th')
      expect(dayOfMonthSuffix(6)).toBe('th')
      expect(dayOfMonthSuffix(7)).toBe('th')
      expect(dayOfMonthSuffix(8)).toBe('th')
      expect(dayOfMonthSuffix(9)).toBe('th')
      expect(dayOfMonthSuffix(10)).toBe('th')
    })

    it('should return "th" for 11, 12, 13 (special cases)', () => {
      expect(dayOfMonthSuffix(11)).toBe('th')
      expect(dayOfMonthSuffix(12)).toBe('th')
      expect(dayOfMonthSuffix(13)).toBe('th')
    })

    it('should return "st" for 21', () => {
      expect(dayOfMonthSuffix(21)).toBe('st')
    })

    it('should return "nd" for 22', () => {
      expect(dayOfMonthSuffix(22)).toBe('nd')
    })

    it('should return "rd" for 23', () => {
      expect(dayOfMonthSuffix(23)).toBe('rd')
    })

    it('should return "th" for 24-30', () => {
      expect(dayOfMonthSuffix(24)).toBe('th')
      expect(dayOfMonthSuffix(25)).toBe('th')
      expect(dayOfMonthSuffix(26)).toBe('th')
      expect(dayOfMonthSuffix(27)).toBe('th')
      expect(dayOfMonthSuffix(28)).toBe('th')
      expect(dayOfMonthSuffix(29)).toBe('th')
      expect(dayOfMonthSuffix(30)).toBe('th')
    })

    it('should return "st" for 31', () => {
      expect(dayOfMonthSuffix(31)).toBe('st')
    })

    it('should handle edge cases correctly', () => {
      // Numbers ending in 1 (except 11)
      expect(dayOfMonthSuffix(1)).toBe('st')
      expect(dayOfMonthSuffix(21)).toBe('st')
      expect(dayOfMonthSuffix(31)).toBe('st')

      // Numbers ending in 2 (except 12)
      expect(dayOfMonthSuffix(2)).toBe('nd')
      expect(dayOfMonthSuffix(22)).toBe('nd')

      // Numbers ending in 3 (except 13)
      expect(dayOfMonthSuffix(3)).toBe('rd')
      expect(dayOfMonthSuffix(23)).toBe('rd')
    })

    it('should handle 0 correctly', () => {
      expect(dayOfMonthSuffix(0)).toBe('th')
    })

    it('should handle numbers greater than 31', () => {
      expect(dayOfMonthSuffix(101)).toBe('st')
      expect(dayOfMonthSuffix(112)).toBe('nd')
      expect(dayOfMonthSuffix(122)).toBe('nd')
    })
  })
})
