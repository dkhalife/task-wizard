import type { AccountInfo } from '@azure/msal-browser'

const TENANT_ID = 'tenant-123'
const OTHER_TENANT = 'other-tenant'

const makeAccount = (tenantId: string, username: string): AccountInfo =>
  ({
    homeAccountId: `${username}.${tenantId}`,
    environment: 'login.microsoftonline.com',
    tenantId,
    username,
    localAccountId: username,
  }) as AccountInfo

const tenantAccount = makeAccount(TENANT_ID, 'user@example.com')
const otherAccount = makeAccount(OTHER_TENANT, 'other@example.com')
const authResult = {
  accessToken: 'access-token',
  account: tenantAccount,
  expiresOn: new Date(Date.now() + 3_600_000),
}

describe('msal utils', () => {
  let mockPca: {
    handleRedirectPromise: jest.Mock
    getActiveAccount: jest.Mock
    getAllAccounts: jest.Mock
    setActiveAccount: jest.Mock
    acquireTokenSilent: jest.Mock
    ssoSilent: jest.Mock
    loginRedirect: jest.Mock
    logoutRedirect: jest.Mock
  }

  let initializeMsal: () => Promise<void>
  let loginSilently: () => Promise<boolean>
  let hasCachedAccounts: () => Promise<boolean>

  beforeEach(async () => {
    jest.resetModules()

    mockPca = {
      handleRedirectPromise: jest.fn().mockResolvedValue(null),
      getActiveAccount: jest.fn().mockReturnValue(null),
      getAllAccounts: jest.fn().mockReturnValue([]),
      setActiveAccount: jest.fn(),
      acquireTokenSilent: jest.fn(),
      ssoSilent: jest.fn(),
      loginRedirect: jest.fn(),
      logoutRedirect: jest.fn(),
    }

    jest.doMock('@azure/msal-browser', () => ({
      createStandardPublicClientApplication: jest.fn().mockResolvedValue(mockPca),
    }))

    jest.doMock('@/api/auth', () => ({
      GetAuthConfig: jest.fn().mockResolvedValue({
        enabled: true,
        client_id: 'client-id',
        tenant_id: TENANT_ID,
        audience: 'https://audience',
      }),
    }))

    const module = await import('@/utils/msal')
    initializeMsal = module.initializeMsal
    loginSilently = module.loginSilently
    hasCachedAccounts = module.hasCachedAccounts
  })

  describe('hasCachedAccounts', () => {
    it('returns false when no accounts are cached', async () => {
      await initializeMsal()
      expect(await hasCachedAccounts()).toBe(false)
    })

    it('returns false when only accounts from other tenants are cached', async () => {
      await initializeMsal()
      mockPca.getAllAccounts.mockReturnValue([otherAccount])
      expect(await hasCachedAccounts()).toBe(false)
    })

    it('returns true when a cached account matches the configured tenant', async () => {
      await initializeMsal()
      mockPca.getAllAccounts.mockReturnValue([tenantAccount])
      expect(await hasCachedAccounts()).toBe(true)
    })

    it('returns true when mixed accounts include one for the configured tenant', async () => {
      await initializeMsal()
      mockPca.getAllAccounts.mockReturnValue([otherAccount, tenantAccount])
      expect(await hasCachedAccounts()).toBe(true)
    })
  })

  describe('loginSilently', () => {
    it('returns true when acquireTokenSilent succeeds', async () => {
      await initializeMsal()
      mockPca.getAllAccounts.mockReturnValue([tenantAccount])
      mockPca.acquireTokenSilent.mockResolvedValue(authResult)

      expect(await loginSilently()).toBe(true)
      expect(mockPca.ssoSilent).not.toHaveBeenCalled()
    })

    it('falls back to ssoSilent when acquireTokenSilent fails', async () => {
      await initializeMsal()
      mockPca.getAllAccounts.mockReturnValue([tenantAccount])
      mockPca.acquireTokenSilent.mockRejectedValue(new Error('token expired'))
      mockPca.ssoSilent.mockResolvedValue({ ...authResult })

      expect(await loginSilently()).toBe(true)
      expect(mockPca.ssoSilent).toHaveBeenCalledWith(
        expect.objectContaining({ loginHint: tenantAccount.username }),
      )
    })

    it('returns false when both acquireTokenSilent and ssoSilent fail', async () => {
      await initializeMsal()
      mockPca.getAllAccounts.mockReturnValue([tenantAccount])
      mockPca.acquireTokenSilent.mockRejectedValue(new Error('token expired'))
      mockPca.ssoSilent.mockRejectedValue(new Error('sso failed'))

      expect(await loginSilently()).toBe(false)
    })

    it('selects the tenant-matching account when multiple accounts are cached', async () => {
      await initializeMsal()
      mockPca.getAllAccounts.mockReturnValue([otherAccount, tenantAccount])
      mockPca.acquireTokenSilent.mockResolvedValue(authResult)

      await loginSilently()

      expect(mockPca.setActiveAccount).toHaveBeenCalledWith(tenantAccount)
      expect(mockPca.setActiveAccount).not.toHaveBeenCalledWith(otherAccount)
    })

    it('clears a wrong active account before throwing when no tenant account exists', async () => {
      await initializeMsal()
      mockPca.getActiveAccount.mockReturnValue(otherAccount)
      mockPca.getAllAccounts.mockReturnValue([otherAccount])

      expect(await loginSilently()).toBe(false)
      expect(mockPca.setActiveAccount).toHaveBeenCalledWith(null)
    })
  })
})
