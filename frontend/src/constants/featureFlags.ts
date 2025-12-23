import { retrieveValue, storeValue } from '@/utils/storage'

export type FeatureFlag = 'sendViaWebsocket' | 'refreshStaleData'

export interface FeatureFlagDefinition {
  name: FeatureFlag
  description: string
  defaultValue: boolean
}

export const featureFlagDefinitions: FeatureFlagDefinition[] = [
  {
    name: 'sendViaWebsocket',
    description: 'Send requests via WebSocket',
    defaultValue: false,
  },
  {
    name: 'refreshStaleData',
    description: 'Refresh stale data when tab becomes visible',
    defaultValue: false,
  },
]

export const FEATURE_FLAG_PREFIX = 'featureFlags.'

export const getFeatureFlag = (
  name: FeatureFlag,
  defaultValue: boolean,
): boolean => retrieveValue<boolean>(FEATURE_FLAG_PREFIX + name, defaultValue)

export const setFeatureFlag = (name: FeatureFlag, value: boolean): void => {
  storeValue<boolean>(FEATURE_FLAG_PREFIX + name, value)
}
