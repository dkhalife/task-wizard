import {
  getTextColorFromBackgroundColor,
  colorOptionFromColor,
  LABEL_COLORS,
} from '@/utils/colors'

describe('colors utils', () => {
  describe('getTextColorFromBackgroundColor', () => {
    it('should return black for light backgrounds', () => {
      expect(getTextColorFromBackgroundColor('#FFFFFF')).toBe('#000000')
      expect(getTextColorFromBackgroundColor('#ffee58')).toBe('#000000')
      expect(getTextColorFromBackgroundColor('#a7ffeb')).toBe('#000000')
    })

    it('should return white for dark backgrounds', () => {
      expect(getTextColorFromBackgroundColor('#000000')).toBe('#ffffff')
      expect(getTextColorFromBackgroundColor('#616161')).toBe('#ffffff')
      expect(getTextColorFromBackgroundColor('#3f51b5')).toBe('#ffffff')
    })

    it('should return empty string for undefined background color', () => {
      expect(getTextColorFromBackgroundColor(undefined)).toBe('')
    })

    it('should handle various color formats correctly', () => {
      // Light colors
      expect(getTextColorFromBackgroundColor('#ffc107')).toBe('#000000')
      expect(getTextColorFromBackgroundColor('#ffca28')).toBe('#000000')
      expect(getTextColorFromBackgroundColor('#ffab91')).toBe('#000000')

      // Dark colors
      expect(getTextColorFromBackgroundColor('#d32f2f')).toBe('#ffffff')
      expect(getTextColorFromBackgroundColor('#0288d1')).toBe('#ffffff')
      expect(getTextColorFromBackgroundColor('#8d6e63')).toBe('#ffffff')
    })

    it('should handle edge case colors near the threshold', () => {
      // Colors near the luminance threshold of 186
      expect(getTextColorFromBackgroundColor('#808080')).toBe('#ffffff')
      expect(getTextColorFromBackgroundColor('#90a4ae')).toBe('#ffffff')
    })
  })

  describe('colorOptionFromColor', () => {
    it('should return the correct color option for a valid color', () => {
      const result = colorOptionFromColor('#FFFFFF')
      expect(result).toEqual({ name: 'Default', value: '#FFFFFF' })
    })

    it('should return the correct color option for another valid color', () => {
      const result = colorOptionFromColor('#ff7961')
      expect(result).toEqual({ name: 'Salmon', value: '#ff7961' })
    })

    it('should return undefined for an invalid color', () => {
      const result = colorOptionFromColor('#invalid')
      expect(result).toBeUndefined()
    })

    it('should return undefined for empty string', () => {
      const result = colorOptionFromColor('')
      expect(result).toBeUndefined()
    })

    it('should handle all predefined label colors', () => {
      LABEL_COLORS.forEach(labelColor => {
        const result = colorOptionFromColor(labelColor.value)
        expect(result).toEqual(labelColor)
      })
    })

    it('should be case sensitive for color matching', () => {
      // Assuming colors are stored in lowercase
      const result = colorOptionFromColor('#ffffff')
      expect(result).toBeUndefined()
    })

    it('should return first match for duplicate colors if any exist', () => {
      const result = colorOptionFromColor('#26a69a')
      expect(result).toEqual({ name: 'Teal', value: '#26a69a' })
    })
  })
})
