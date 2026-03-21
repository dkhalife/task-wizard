package com.dkhalife.tasks.auth

import android.app.Activity
import com.microsoft.identity.client.AcquireTokenParameters
import com.microsoft.identity.client.AcquireTokenSilentParameters
import com.microsoft.identity.client.AuthenticationCallback
import com.microsoft.identity.client.IAccount
import com.microsoft.identity.client.IAuthenticationResult
import com.microsoft.identity.client.ISingleAccountPublicClientApplication
import com.microsoft.identity.client.SilentAuthenticationCallback
import com.microsoft.identity.client.exception.MsalException
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthManager @Inject constructor() : AuthTokenProvider {

    private var singleAccountApp: ISingleAccountPublicClientApplication? = null

    @Volatile
    private var currentAccount: IAccount? = null

    @Volatile
    private var cachedAccessToken: String? = null

    @Volatile
    private var cachedExpiryTimeMs: Long = 0

    @Volatile
    private var isAccountLoaded = false

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
        }
        isAccountLoaded = true
        notifyUserChanged(account != null)
    }

    fun isSignedIn(): Boolean = currentAccount != null

    fun isLoaded(): Boolean = isAccountLoaded

    override fun getCachedAccessToken(): String? {
        val token = cachedAccessToken
        if (token != null && System.currentTimeMillis() < cachedExpiryTimeMs - TOKEN_SKEW_MS) {
            return token
        }
        return null
    }

    override suspend fun getAccessToken(): String? {
        val token = cachedAccessToken
        if (token != null && System.currentTimeMillis() < cachedExpiryTimeMs - TOKEN_SKEW_MS) {
            return token
        }

        val account = currentAccount ?: return null
        val app = singleAccountApp ?: return null

        return try {
            val result = acquireTokenSilent(app, account)
            cacheToken(result)
            result.accessToken
        } catch (e: MsalException) {
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
        val app = singleAccountApp ?: return

        val params = AcquireTokenParameters.Builder()
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
            .build()

        app.acquireToken(params)
    }

    fun signOut(callback: ISingleAccountPublicClientApplication.SignOutCallback) {
        val app = singleAccountApp ?: return
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
    }

    companion object {
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
