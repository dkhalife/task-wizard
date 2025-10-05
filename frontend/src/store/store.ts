import { configureStore } from '@reduxjs/toolkit'
import { TypedUseSelectorHook, useDispatch, useSelector } from 'react-redux'
import { tasksReducer } from './tasksSlice'
import { labelsReducer } from './labelsSlice'
import { userReducer } from './userSlice'
import { tokensReducer } from './tokensSlice'
import { statusReducer } from './statusSlice'
import { wsReducer } from './wsSlice'
import { featureFlagsReducer } from './featureFlagsSlice'

export const store = configureStore({
  reducer: {
    featureFlags: featureFlagsReducer,
    tasks: tasksReducer,
    labels: labelsReducer,
    user: userReducer,
    tokens: tokensReducer,
    status: statusReducer,
    ws: wsReducer,
  },
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch

export const useAppDispatch: () => AppDispatch = useDispatch
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector
