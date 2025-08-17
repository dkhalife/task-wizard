export function retrieveValue<T>(key: string, defaultValue: T): T {
  const stickyValue = window.localStorage.getItem(key)
  if (stickyValue === null) return defaultValue
  try {
    return JSON.parse(stickyValue) as T
  } catch (error) {
    console.error('Failed to parse localStorage key', key, error)
    window.localStorage.removeItem(key)
    return defaultValue
  }
}

export function storeValue<T>(key: string, value: T) {
  window.localStorage.setItem(key, JSON.stringify(value))
}
