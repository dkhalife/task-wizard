import {
  getFeatureFlag,
  setFeatureFlag,
  FEATURE_FLAG_PREFIX,
} from '@/constants/featureFlags'

describe('featureFlags', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  describe('setFeatureFlag', () => {
    it('should store feature flag with correct prefix', () => {
      setFeatureFlag('someFlag', true)
      expect(localStorage.getItem(FEATURE_FLAG_PREFIX + 'someFlag')).toBe('true')
    })

    it('should store false value correctly', () => {
      setFeatureFlag('someFlag', false)
      expect(localStorage.getItem(FEATURE_FLAG_PREFIX + 'someFlag')).toBe('false')
    })

    it('should update existing feature flag', () => {
      setFeatureFlag('someFlag', true)
      expect(localStorage.getItem(FEATURE_FLAG_PREFIX + 'someFlag')).toBe('true')

      setFeatureFlag('someFlag', false)
      expect(localStorage.getItem(FEATURE_FLAG_PREFIX + 'someFlag')).toBe('false')
    })

    it('should handle multiple feature flags independently', () => {
      setFeatureFlag('flagA', true)
      setFeatureFlag('flagB', false)

      expect(localStorage.getItem(FEATURE_FLAG_PREFIX + 'flagA')).toBe('true')
      expect(localStorage.getItem(FEATURE_FLAG_PREFIX + 'flagB')).toBe('false')
    })
  })

  describe('getFeatureFlag', () => {
    it('should return stored value when flag exists', () => {
      localStorage.setItem(FEATURE_FLAG_PREFIX + 'someFlag', 'true')
      expect(getFeatureFlag('someFlag', false)).toBe(true)
    })

    it('should return default value when flag does not exist', () => {
      expect(getFeatureFlag('someFlag', false)).toBe(false)
    })

    it('should return different default values correctly', () => {
      expect(getFeatureFlag('someFlag', true)).toBe(true)
      expect(getFeatureFlag('otherFlag', false)).toBe(false)
    })

    it('should retrieve false value correctly', () => {
      localStorage.setItem(FEATURE_FLAG_PREFIX + 'someFlag', 'false')
      expect(getFeatureFlag('someFlag', true)).toBe(false)
    })

    it('should retrieve true value correctly', () => {
      localStorage.setItem(FEATURE_FLAG_PREFIX + 'someFlag', 'true')
      expect(getFeatureFlag('someFlag', false)).toBe(true)
    })
  })

  describe('getFeatureFlag and setFeatureFlag integration', () => {
    it('should store and retrieve feature flag correctly', () => {
      setFeatureFlag('someFlag', true)
      expect(getFeatureFlag('someFlag', false)).toBe(true)
    })

    it('should update and retrieve updated value', () => {
      setFeatureFlag('someFlag', true)
      expect(getFeatureFlag('someFlag', false)).toBe(true)

      setFeatureFlag('someFlag', false)
      expect(getFeatureFlag('someFlag', true)).toBe(false)
    })
  })
})
