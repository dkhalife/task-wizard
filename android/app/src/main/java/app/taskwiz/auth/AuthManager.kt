package app.taskwiz.auth

import android.app.Activity
import com.microsoft.identity.client.AcquireTokenParameters
import com.microsoft.identity.client.AcquireTokenSilentParameters
import com.microsoft.identity.client.AuthenticationCallback
import com.microsoft.identity.client.IAccount
import com.microsoft.identity.client.IAuthenticationResult
import com.microsoft.identity.client.ISingleAccountPublicClientApplication
import com.microsoft.identity.client.SilentAuthenticationCallback
import com.microsoft.identity.client.exception.MsalException
import com.microsoft.identity.client.exception.MsalUiRequiredException
import app.taskwiz.telemetry.TelemetryManager
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthManager @Inject constructor(
    private val telemetryManager: TelemetryManager
) : AuthTokenProvider {

    private var singleAccountApp: ISingleAccountPublicClientApplication? = null

    @Volatile
    private var currentAccount: IAccount? = null

    @Volatile
    private var cachedAccessToken: String? = null

    @Volatile
    private var cachedExpiryTimeMs: Long = 0

    @Volatile
    private var isAccountLoaded = false

    private val _sessionExpired = MutableStateFlow(false)
    /**
     * Emits true when a silent token refresh fails non-recoverably (the MSAL refresh token is gone
     * or revoked), meaning interactive sign-in is required. The account record still exists so
     * [isSignedIn] stays true; this flag is how the UI learns the session has effectively expired.
     */
    val sessionExpired: StateFlow<Boolean> = _sessionExpired

    private val userChangeListeners = mutableListOf<UserChangeListener>()

    fun interface UserChangeListener {
        fun onUserChanged(isSignedIn: Boolean)
    }

    fun addUserChangeListener(listener: UserChangeListener) {
        synchronized(userChangeListeners) {
            userChangeListeners.add(listener)
        }
    }

    fun removeUserChangeListener(listener: UserChangeListener) {
        synchronized(userChangeListeners) {
            userChangeListeners.remove(listener)
        }
    }

    private fun notifyUserChanged(isSignedIn: Boolean) {
        val snapshot = synchronized(userChangeListeners) {
            userChangeListeners.toList()
        }
        snapshot.forEach { it.onUserChanged(isSignedIn) }
    }

    fun registerSingleAccountApp(app: ISingleAccountPublicClientApplication) {
        singleAccountApp = app
    }

    fun getSingleAccountApp(): ISingleAccountPublicClientApplication? = singleAccountApp

    fun updateAccount(account: IAccount?) {
        currentAccount = account
        if (account == null) {
            cachedAccessToken = null
            cachedExpiryTimeMs = 0
            _sessionExpired.value = false
        }
        isAccountLoaded = true
        notifyUserChanged(account != null)
    }

    fun isSignedIn(): Boolean = currentAccount != null

    fun getAccountName(): String? = currentAccount?.username

    fun isLoaded(): Boolean = isAccountLoaded

    override fun getCachedAccessToken(): String? {
        val token = cachedAccessToken
        if (token != null && System.currentTimeMillis() < cachedExpiryTimeMs - TOKEN_SKEW_MS) {
            return token
        }
        return null
    }

    override suspend fun getAccessToken(forceRefresh: Boolean): String? {
        if (!forceRefresh) {
            val token = cachedAccessToken
            if (token != null && System.currentTimeMillis() < cachedExpiryTimeMs - TOKEN_SKEW_MS) {
                return token
            }
        }

        val account = currentAccount ?: return null
        val app = singleAccountApp ?: return null

        return try {
            val result = acquireTokenSilent(app, account)
            cacheToken(result)
            result.accessToken
        } catch (e: MsalException) {
            telemetryManager.logWarning(TAG, "Silent token acquire failed: ${e.message}", e)
            if (e is MsalUiRequiredException) {
                cachedAccessToken = null
                cachedExpiryTimeMs = 0
                _sessionExpired.value = true
            }
            null
        }
    }

    private fun acquireTokenSilent(
        app: ISingleAccountPublicClientApplication,
        account: IAccount
    ): IAuthenticationResult {
        val params = AcquireTokenSilentParameters.Builder()
            .forAccount(account)
            .fromAuthority(account.authority)
            .withScopes(REQUIRED_SCOPES)
            .build()

        return app.acquireTokenSilent(params)
    }

    fun signIn(activity: Activity, callback: AuthenticationCallback) {
        val app = singleAccountApp ?: run {
            telemetryManager.logError(TAG, "Sign-in failed: singleAccountApp not initialized")
            return
        }

        val builder = AcquireTokenParameters.Builder()
            .startAuthorizationFromActivity(activity)
            .withScopes(REQUIRED_SCOPES)
            .withCallback(object : AuthenticationCallback {
                override fun onSuccess(authenticationResult: IAuthenticationResult) {
                    cacheToken(authenticationResult)
                    updateAccount(authenticationResult.account)
                    callback.onSuccess(authenticationResult)
                }

                override fun onError(exception: MsalException) {
                    callback.onError(exception)
                }

                override fun onCancel() {
                    callback.onCancel()
                }
            })

        // In single-account mode an interactive acquireToken while an account is already persisted
        // (re-authenticating after a session expired) must target that same account, otherwise MSAL
        // rejects it with CURRENT_ACCOUNT_MISMATCH before any sign-in UI is shown.
        currentAccount?.let { builder.forAccount(it) }

        app.acquireToken(builder.build())
    }

    fun signOut(callback: ISingleAccountPublicClientApplication.SignOutCallback) {
        val app = singleAccountApp ?: run {
            telemetryManager.logError(TAG, "Sign-out failed: singleAccountApp not initialized")
            return
        }
        app.signOut(object : ISingleAccountPublicClientApplication.SignOutCallback {
            override fun onSignOut() {
                updateAccount(null)
                callback.onSignOut()
            }

            override fun onError(exception: MsalException) {
                callback.onError(exception)
            }
        })
    }

    private fun cacheToken(result: IAuthenticationResult) {
        cachedAccessToken = result.accessToken
        cachedExpiryTimeMs = result.expiresOn.time
        _sessionExpired.value = false
    }

    companion object {
        private const val TAG = "AuthManager"
        private const val TOKEN_SKEW_MS = 120_000L

        val REQUIRED_SCOPES: List<String>
            get() {
                return listOf(
                    "api://task-wizard/User.Read",
                    "api://task-wizard/User.Write",
                    "api://task-wizard/Labels.Read",
                    "api://task-wizard/Labels.Write",
                    "api://task-wizard/Tasks.Read",
                    "api://task-wizard/Tasks.Write"
                )
            }
    }
}
