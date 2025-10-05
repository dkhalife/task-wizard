import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import {
    FeatureFlag,
  featureFlagDefinitions,
  getFeatureFlag as getFlag,
  setFeatureFlag as persistFlag,
} from '@/constants/featureFlags'

export type FeatureFlagsState = Record<string, boolean>

const initialState: FeatureFlagsState = {}
featureFlagDefinitions.forEach(def => {
  initialState[def.name] = getFlag(def.name, def.defaultValue)
})

const featureFlagsSlice = createSlice({
  name: 'featureFlags',
  initialState,
  reducers: {
    setFlag(state, action: PayloadAction<{ name: FeatureFlag; value: boolean }>) {
      const { name, value } = action.payload
      state[name] = value
      persistFlag(name, value)
    },
  },
})

export const featureFlagsReducer = featureFlagsSlice.reducer
export const { setFlag } = featureFlagsSlice.actions
