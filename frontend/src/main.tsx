import React from 'react'
import ReactDOM from 'react-dom/client'
import { useRoot } from './utils/dom'
import { RouterContext } from './contexts/RouterContext'
import { ErrorBoundary } from './components/ErrorBoundary'
import { LogError, LogWarning } from './api/log'
import { Provider } from 'react-redux'
import { store } from './store/store'
import { isAuthEnabled } from './utils/msal'

window.onerror = (message, source, lineno, colno) => {
  if (!isAuthEnabled()) {
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

  if (!isAuthEnabled()) {
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
