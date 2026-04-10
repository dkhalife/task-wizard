import type {
  AccountInfo,
  AuthenticationResult,
  IPublicClientApplication,
} from '@azure/msal-browser'
import { createStandardPublicClientApplication } from '@azure/msal-browser'
import { GetAuthConfig } from '@/api/auth'
import type { AuthConfig } from '@/api/auth'

let authConfig: AuthConfig | null = null
let cachedAuthResult: AuthenticationResult | null = null
let pcaPromise: Promise<IPublicClientApplication> | null = null
let initPromise: Promise<void> | null = null

export const initializeMsal = () => {
  if (!initPromise) {
    initPromise = doInitializeMsal()
  }
  return initPromise
}

const doInitializeMsal = async () => {
  authConfig = await GetAuthConfig()
  if (!authConfig.enabled) return

  pcaPromise = createStandardPublicClientApplication({
    auth: {
      clientId: authConfig.client_id,
      authority: `https://login.microsoftonline.com/${authConfig.tenant_id}`,
      redirectUri: `${window.location.origin}/login`,
    },
    cache: {
      cacheLocation: 'localStorage',
    },
  })

  const pca = await pcaPromise

  try {
    const response = await pca.handleRedirectPromise()
    if (response?.account) {
      pca.setActiveAccount(response.account)
      cachedAuthResult = response
    }
  } catch {
    // Allow the app to continue in unauthenticated state
  }
}

const getScopes = (): string[] => {
  if (!authConfig) return []
  return [
    `${authConfig.audience}/Tasks.Read`,
    `${authConfig.audience}/Tasks.Write`,
  ]
}

const ensureActiveAccount = (pca: IPublicClientApplication): AccountInfo => {
  if (!authConfig) {
    throw new Error('Authentication is not configured')
  }

  const activeAccount = pca.getActiveAccount()
  if (activeAccount?.tenantId === authConfig.tenant_id) {
    return activeAccount
  }

  const tenantAccount = pca.getAllAccounts().find(a => a.tenantId === authConfig!.tenant_id)
  if (!tenantAccount) {
    if (activeAccount) {
      pca.setActiveAccount(null)
    }
    throw new Error('No accounts found for configured tenant')
  }
  pca.setActiveAccount(tenantAccount)
  return tenantAccount
}

export const isAuthEnabled = (): boolean => {
  return authConfig?.enabled ?? false
}

export const loginWithRedirect = async () => {
  if (!authConfig?.enabled || !pcaPromise) return
  const pca = await pcaPromise
  await pca.loginRedirect({ scopes: getScopes() })
}

export const hasCachedAccounts = async (): Promise<boolean> => {
  if (!authConfig?.enabled || !pcaPromise) return false
  const pca = await pcaPromise
  return pca.getAllAccounts().some(a => a.tenantId === authConfig?.tenant_id)
}

const AUTH_COOKIE = 'tw_auth'

const setAuthCookie = () => {
  document.cookie = `${AUTH_COOKIE}=1; path=/; SameSite=Strict; max-age=31536000`
}

const clearAuthCookie = () => {
  document.cookie = `${AUTH_COOKIE}=; path=/; SameSite=Strict; max-age=0`
}

export const loginSilently = async (): Promise<boolean> => {
  if (!authConfig?.enabled || !pcaPromise) return true
  const pca = await pcaPromise
  try {
    const account = ensureActiveAccount(pca)
    cachedAuthResult = await pca.acquireTokenSilent({ scopes: getScopes(), account })
    setAuthCookie()
    return true
  } catch {
    try {
      const account = pca.getActiveAccount() ?? pca.getAllAccounts().find(a => a.tenantId === authConfig?.tenant_id)
      cachedAuthResult = await pca.ssoSilent({
        scopes: getScopes(),
        loginHint: account?.username,
      })
      if (cachedAuthResult.account) {
        pca.setActiveAccount(cachedAuthResult.account)
      }
      setAuthCookie()
      return true
    } catch {
      clearAuthCookie()
      return false
    }
  }
}

export const acquireAccessToken = async (): Promise<string> => {
  if (!authConfig?.enabled) return ''
  if (!pcaPromise) throw new Error('MSAL not initialized')
  if (cachedAuthResult?.accessToken && cachedAuthResult.expiresOn && cachedAuthResult.expiresOn.getTime() > Date.now()) {
    return cachedAuthResult.accessToken
  }
  const pca = await pcaPromise
  const account = ensureActiveAccount(pca)
  cachedAuthResult = await pca.acquireTokenSilent({ scopes: getScopes(), account })
  return cachedAuthResult.accessToken
}

export const logout = async () => {
  clearAuthCookie()
  if (!authConfig?.enabled || !pcaPromise) {
    window.location.href = '/'
    return
  }
  const pca = await pcaPromise
  cachedAuthResult = null
  await pca.logoutRedirect({ postLogoutRedirectUri: `${window.location.origin}/login` })
}
