import { Theme } from '@/constants/theme'

let root: HTMLDivElement | null = null

export const useRoot = (): HTMLDivElement => {
  if (root) {
    return root
  }

  root = document.createElement('div')
  document.body.appendChild(root)
  return root
}

export const setTitle = (title: string): void => {
  document.title = `Task Wizard - ${title}`
}

export const isMobile = (): boolean => {
  return window.innerWidth <= 768
}

export const applyTheme = (theme: Theme): void => {
  document.body.dataset.theme = theme
}
