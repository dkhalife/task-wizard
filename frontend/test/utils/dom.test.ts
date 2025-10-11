import { isMobile, setTitle, applyTheme } from '@/utils/dom'
import { Theme } from '@/constants/theme'

describe('dom utils', () => {
  beforeEach(() => {
    // Clear any existing root elements
    document.body.innerHTML = ''
    // Reset window size
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: 1024,
    })
  })

  describe('isMobile', () => {
    it('should return true when window width is 768 or less', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 768,
      })
      expect(isMobile()).toBe(true)
    })

    it('should return true when window width is less than 768', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 500,
      })
      expect(isMobile()).toBe(true)
    })

    it('should return false when window width is greater than 768', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1024,
      })
      expect(isMobile()).toBe(false)
    })

    it('should return false for typical desktop width', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1920,
      })
      expect(isMobile()).toBe(false)
    })
  })

  describe('setTitle', () => {
    it('should set document title with prefix', () => {
      setTitle('Dashboard')
      expect(document.title).toBe('Task Wizard - Dashboard')
    })

    it('should set document title with another value', () => {
      setTitle('Settings')
      expect(document.title).toBe('Task Wizard - Settings')
    })

    it('should handle empty string', () => {
      setTitle('')
      expect(document.title).toBe('Task Wizard -')
    })

    it('should update document title when called multiple times', () => {
      setTitle('First')
      expect(document.title).toBe('Task Wizard - First')
      setTitle('Second')
      expect(document.title).toBe('Task Wizard - Second')
    })
  })



  describe('applyTheme', () => {
    it('should set theme data attribute to light', () => {
      applyTheme('light' as Theme)
      expect(document.body.dataset.theme).toBe('light')
    })

    it('should set theme data attribute to dark', () => {
      applyTheme('dark' as Theme)
      expect(document.body.dataset.theme).toBe('dark')
    })

    it('should update theme when called multiple times', () => {
      applyTheme('light' as Theme)
      expect(document.body.dataset.theme).toBe('light')
      applyTheme('dark' as Theme)
      expect(document.body.dataset.theme).toBe('dark')
    })
  })
})
