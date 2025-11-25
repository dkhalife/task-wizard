import { ApiTokenScopesList } from "@/models/token";
import { AuthenticationResult, createStandardPublicClientApplication } from "@azure/msal-browser";

export const TENANT_ID = import.meta.env.VITE_APP_TENANT_ID;

const REDIRECT_URL = import.meta.env.VITE_APP_REDIRECT_URL;
const clientId = import.meta.env.VITE_APP_CLIENT_ID;

const msalConfig = {
  auth: {
    clientId: clientId,
    authority: `https://login.microsoftonline.com/${TENANT_ID}`,
    redirectUri: REDIRECT_URL,
  }
};

const loginRequest = {
  scopes: ApiTokenScopesList
    .filter(scope => !scope.startsWith("Dav."))
    .map(scope => `api://task-wizard/${scope}`),
};

let cachedAuthResult: AuthenticationResult | null = null;
export const loginWithPopup = async () => {
    const pca = await createStandardPublicClientApplication(msalConfig)
    cachedAuthResult = await pca.loginPopup(loginRequest);
}

export const loginSilently = async () => {
    const pca = await createStandardPublicClientApplication(msalConfig)
    const accounts = pca.getAllAccounts();
    if (accounts.length === 0) {
        throw new Error("No accounts found for silent login");
    }

    cachedAuthResult = await pca.acquireTokenSilent(loginRequest);
}

export const isTokenValid = (): boolean => {
  if (cachedAuthResult === null) {
    return false
  }

  if (!cachedAuthResult.accessToken) {
    return false
  }

  if (!cachedAuthResult.expiresOn) {
    return false
  }

  return cachedAuthResult.expiresOn.getTime() > Date.now();
}

export const getAccessToken = (): string => {
    if (!isTokenValid()) {
        throw new Error("Token is not valid");
    }

  const accessToken = cachedAuthResult?.accessToken;
  if (!accessToken) {
    throw new Error("Token is not valid");
  }

  return accessToken;
}

export const logout = async () => {
    const pca = await createStandardPublicClientApplication(msalConfig)
    const accounts = pca.getAllAccounts();
    if (accounts.length > 0) {
        await pca.logout()
    }

    window.location.href = '/login';
}
