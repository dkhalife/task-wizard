import {
  getFeatureFlag,
  setFeatureFlag,
  FEATURE_FLAG_PREFIX,
  FeatureFlag,
} from '@/constants/featureFlags'

describe('featureFlags', () => {
  beforeEach(() => {
    // Clear localStorage before each test
    localStorage.clear()
  })

  describe('setFeatureFlag', () => {
    it('should store feature flag with correct prefix', () => {
      setFeatureFlag('useWebsockets', true)
      const storedValue = localStorage.getItem(FEATURE_FLAG_PREFIX + 'useWebsockets')
      expect(storedValue).toBe('true')
    })

    it('should store false value correctly', () => {
      setFeatureFlag('refreshStaleData', false)
      const storedValue = localStorage.getItem(FEATURE_FLAG_PREFIX + 'refreshStaleData')
      expect(storedValue).toBe('false')
    })

    it('should update existing feature flag', () => {
      setFeatureFlag('useWebsockets', true)
      expect(localStorage.getItem(FEATURE_FLAG_PREFIX + 'useWebsockets')).toBe('true')
      
      setFeatureFlag('useWebsockets', false)
      expect(localStorage.getItem(FEATURE_FLAG_PREFIX + 'useWebsockets')).toBe('false')
    })

    it('should handle multiple feature flags independently', () => {
      setFeatureFlag('useWebsockets', true)
      setFeatureFlag('refreshStaleData', false)
      
      expect(localStorage.getItem(FEATURE_FLAG_PREFIX + 'useWebsockets')).toBe('true')
      expect(localStorage.getItem(FEATURE_FLAG_PREFIX + 'refreshStaleData')).toBe('false')
    })
  })

  describe('getFeatureFlag', () => {
    it('should return stored value when flag exists', () => {
      localStorage.setItem(FEATURE_FLAG_PREFIX + 'useWebsockets', 'true')
      const result = getFeatureFlag('useWebsockets', false)
      expect(result).toBe(true)
    })

    it('should return default value when flag does not exist', () => {
      const result = getFeatureFlag('useWebsockets', false)
      expect(result).toBe(false)
    })

    it('should return different default values correctly', () => {
      const result1 = getFeatureFlag('useWebsockets', true)
      expect(result1).toBe(true)
      
      const result2 = getFeatureFlag('refreshStaleData', false)
      expect(result2).toBe(false)
    })

    it('should retrieve false value correctly', () => {
      localStorage.setItem(FEATURE_FLAG_PREFIX + 'refreshStaleData', 'false')
      const result = getFeatureFlag('refreshStaleData', true)
      expect(result).toBe(false)
    })

    it('should retrieve true value correctly', () => {
      localStorage.setItem(FEATURE_FLAG_PREFIX + 'useWebsockets', 'true')
      const result = getFeatureFlag('useWebsockets', false)
      expect(result).toBe(true)
    })

    it('should handle all feature flag types', () => {
      const flags: FeatureFlag[] = ['useWebsockets', 'sendViaWebsocket', 'refreshStaleData']
      
      flags.forEach(flag => {
        setFeatureFlag(flag, true)
        expect(getFeatureFlag(flag, false)).toBe(true)
      })
    })
  })

  describe('getFeatureFlag and setFeatureFlag integration', () => {
    it('should store and retrieve feature flag correctly', () => {
      setFeatureFlag('useWebsockets', true)
      expect(getFeatureFlag('useWebsockets', false)).toBe(true)
    })

    it('should update and retrieve updated value', () => {
      setFeatureFlag('refreshStaleData', true)
      expect(getFeatureFlag('refreshStaleData', false)).toBe(true)
      
      setFeatureFlag('refreshStaleData', false)
      expect(getFeatureFlag('refreshStaleData', true)).toBe(false)
    })

    it('should maintain feature flag values independently', () => {
      setFeatureFlag('useWebsockets', true)
      setFeatureFlag('refreshStaleData', false)
      
      expect(getFeatureFlag('useWebsockets', false)).toBe(true)
      expect(getFeatureFlag('refreshStaleData', true)).toBe(false)
    })

    it('should handle toggling feature flags', () => {
      // Initial state
      setFeatureFlag('useWebsockets', false)
      expect(getFeatureFlag('useWebsockets', false)).toBe(false)
      
      // Toggle on
      setFeatureFlag('useWebsockets', true)
      expect(getFeatureFlag('useWebsockets', false)).toBe(true)
      
      // Toggle off
      setFeatureFlag('useWebsockets', false)
      expect(getFeatureFlag('useWebsockets', false)).toBe(false)
    })
  })

  describe('FEATURE_FLAG_PREFIX', () => {
    it('should be correctly defined', () => {
      expect(FEATURE_FLAG_PREFIX).toBe('featureFlags.')
    })

    it('should be used as prefix for all feature flags', () => {
      setFeatureFlag('useWebsockets', true)
      const keys = Object.keys(localStorage)
      const flagKeys = keys.filter(key => key.startsWith(FEATURE_FLAG_PREFIX))
      expect(flagKeys.length).toBeGreaterThan(0)
      expect(flagKeys[0]).toContain(FEATURE_FLAG_PREFIX)
    })
  })
})
