import React from 'react'
import ReactDOM from 'react-dom/client'
import { useRoot } from './utils/dom'
import { RouterContext } from './contexts/RouterContext'
import { ErrorBoundary } from './components/ErrorBoundary'
import { LogError, LogWarning } from './api/log'
import { Provider } from 'react-redux'
import { store } from './store/store'
import { isTokenValid } from './utils/api'

window.onerror = (message, source, lineno, colno) => {
  // Telemetry APIs are behind auth wall, if the user isn't authenticated, we can't do anything
  if (!isTokenValid()) {
    return
  }

  try {  
    LogError(`${source}:${lineno}:${colno} ${message}`, window.location.pathname)
  } catch {
    console.debug('Fatal error: ', message, source, lineno, colno)
  }

  return true
}

window.onunhandledrejection = async (event) => {
  event.preventDefault()
  event.stopImmediatePropagation()

  // Telemetry APIs are behind auth wall, if the user isn't authenticated, we can't do anything
  if (!isTokenValid()) {
    return
  }

  try {
    await LogWarning(event.reason, window.location.pathname)
  } catch {
    console.debug('Fatal error: ', event.reason)
  }
}

ReactDOM.createRoot(useRoot()).render(
  <React.StrictMode>
    <ErrorBoundary>
      <Provider store={store}>
        <RouterContext />
      </Provider>
    </ErrorBoundary>
  </React.StrictMode>,
)
