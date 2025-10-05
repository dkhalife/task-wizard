import { applyTheme } from '@/utils/dom'
import { Mode } from '@mui/system/cssVars/useCurrentColorScheme'

export type Theme = 'light' | 'dark'

export const getCurrentThemeMode = (): Mode => {
  const mode = localStorage.getItem('themeMode')
  if (mode === 'light' || mode === 'dark' || mode === 'system') {
    return mode
  }

  return 'system'
}

export const getNextThemeMode = (currentThemeMode: Mode): Mode => {
  switch (currentThemeMode) {
    case 'light':
      return 'dark'
    case 'dark':
      return 'system'
    case 'system':
    default:
      return 'light'
  }
}

function prefersDarkMode() {
  return (
    window.matchMedia &&
    window.matchMedia('(prefers-color-scheme: dark)').matches
  )
}

function themeForMode(mode: Mode): Theme {
  switch (mode) {
    case 'light':
    case 'dark':
      return mode

    case 'system':
    default:
      return prefersDarkMode() ? 'dark' : 'light'
  }
}

export const setThemeMode = (mode: Mode): void => {
  localStorage.setItem('themeMode', mode)
  applyTheme(themeForMode(mode))
}
