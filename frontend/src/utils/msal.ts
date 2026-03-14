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

export const initializeMsal = async () => {
  authConfig = await GetAuthConfig()
  if (!authConfig.enabled) return

  const pca = await createStandardPublicClientApplication({
    auth: {
      clientId: authConfig.client_id,
      authority: `https://login.microsoftonline.com/${authConfig.tenant_id}`,
      redirectUri: window.location.origin,
    },
    cache: {
      cacheLocation: 'localStorage',
    },
  })

  pcaPromise = Promise.resolve(pca)

  const response = await pca.handleRedirectPromise()
  if (response?.account) {
    pca.setActiveAccount(response.account)
    cachedAuthResult = response
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
  const activeAccount = pca.getActiveAccount()
  if (activeAccount) return activeAccount
  const [firstAccount] = pca.getAllAccounts()
  if (!firstAccount) throw new Error('No accounts found')
  pca.setActiveAccount(firstAccount)
  return firstAccount
}

export const isAuthEnabled = (): boolean => {
  return authConfig?.enabled ?? false
}

export const loginWithRedirect = async () => {
  if (!authConfig?.enabled || !pcaPromise) return
  const pca = await pcaPromise
  await pca.loginRedirect({ scopes: getScopes() })
}

export const loginSilently = async (): Promise<boolean> => {
  if (!authConfig?.enabled || !pcaPromise) return true
  const pca = await pcaPromise
  try {
    const account = ensureActiveAccount(pca)
    cachedAuthResult = await pca.acquireTokenSilent({ scopes: getScopes(), account })
    return true
  } catch {
    return false
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
  if (!authConfig?.enabled || !pcaPromise) {
    window.location.href = '/'
    return
  }
  const pca = await pcaPromise
  cachedAuthResult = null
  await pca.logoutRedirect({ postLogoutRedirectUri: window.location.origin })
}
