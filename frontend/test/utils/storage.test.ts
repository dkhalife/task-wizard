import { retrieveValue, storeValue } from '@/utils/storage'

describe('storage utils', () => {
  beforeEach(() => {
    // Clear localStorage before each test
    localStorage.clear()
    jest.clearAllMocks()
  })

  describe('storeValue', () => {
    it('should store a string value in localStorage', () => {
      storeValue('testKey', 'testValue')
      expect(localStorage.getItem('testKey')).toBe('"testValue"')
    })

    it('should store a number value in localStorage', () => {
      storeValue('testKey', 42)
      expect(localStorage.getItem('testKey')).toBe('42')
    })

    it('should store a boolean value in localStorage', () => {
      storeValue('testKey', true)
      expect(localStorage.getItem('testKey')).toBe('true')
    })

    it('should store an object value in localStorage', () => {
      const testObj = { foo: 'bar', num: 123 }
      storeValue('testKey', testObj)
      expect(localStorage.getItem('testKey')).toBe(JSON.stringify(testObj))
    })

    it('should store an array value in localStorage', () => {
      const testArray = [1, 2, 3, 'test']
      storeValue('testKey', testArray)
      expect(localStorage.getItem('testKey')).toBe(JSON.stringify(testArray))
    })
  })

  describe('retrieveValue', () => {
    it('should retrieve a stored string value from localStorage', () => {
      localStorage.setItem('testKey', '"testValue"')
      const result = retrieveValue('testKey', 'default')
      expect(result).toBe('testValue')
    })

    it('should retrieve a stored number value from localStorage', () => {
      localStorage.setItem('testKey', '42')
      const result = retrieveValue('testKey', 0)
      expect(result).toBe(42)
    })

    it('should retrieve a stored boolean value from localStorage', () => {
      localStorage.setItem('testKey', 'true')
      const result = retrieveValue('testKey', false)
      expect(result).toBe(true)
    })

    it('should retrieve a stored object from localStorage', () => {
      const testObj = { foo: 'bar', num: 123 }
      localStorage.setItem('testKey', JSON.stringify(testObj))
      const result = retrieveValue('testKey', {})
      expect(result).toEqual(testObj)
    })

    it('should retrieve a stored array from localStorage', () => {
      const testArray = [1, 2, 3, 'test']
      localStorage.setItem('testKey', JSON.stringify(testArray))
      const result = retrieveValue('testKey', [])
      expect(result).toEqual(testArray)
    })

    it('should return default value when key does not exist', () => {
      const result = retrieveValue('nonExistentKey', 'defaultValue')
      expect(result).toBe('defaultValue')
    })

    it('should return default value when stored value is null', () => {
      localStorage.setItem('testKey', 'null')
      const result = retrieveValue('testKey', 'defaultValue')
      expect(result).toBe(null)
    })

    it('should handle invalid JSON and return default value', () => {
      const consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation()
      localStorage.setItem('testKey', 'invalid JSON {')
      const result = retrieveValue('testKey', 'defaultValue')
      expect(result).toBe('defaultValue')
      expect(localStorage.getItem('testKey')).toBeNull()
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        'Failed to parse localStorage key',
        'testKey',
        expect.any(Error)
      )
      consoleErrorSpy.mockRestore()
    })

    it('should remove invalid JSON from localStorage', () => {
      const consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation()
      localStorage.setItem('testKey', 'invalid JSON')
      retrieveValue('testKey', 'defaultValue')
      expect(localStorage.getItem('testKey')).toBeNull()
      consoleErrorSpy.mockRestore()
    })
  })

  describe('storeValue and retrieveValue integration', () => {
    it('should store and retrieve a value correctly', () => {
      const testValue = { name: 'Test', count: 5, active: true }
      storeValue('integrationKey', testValue)
      const retrieved = retrieveValue('integrationKey', {})
      expect(retrieved).toEqual(testValue)
    })

    it('should handle multiple keys independently', () => {
      storeValue('key1', 'value1')
      storeValue('key2', 42)
      storeValue('key3', true)

      expect(retrieveValue('key1', '')).toBe('value1')
      expect(retrieveValue('key2', 0)).toBe(42)
      expect(retrieveValue('key3', false)).toBe(true)
    })
  })
})
