package com.dkhalife.tasks.data.sync

import com.dkhalife.tasks.auth.AuthManager
import com.dkhalife.tasks.ws.WebSocketManager
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class WebSocketLifecycleManager @Inject constructor(
    private val webSocketManager: WebSocketManager,
    private val webSocketSyncBridge: WebSocketSyncBridge,
    private val authManager: AuthManager
) {
    private val authListener = AuthManager.UserChangeListener { isSignedIn ->
        if (isSignedIn) {
            webSocketManager.connect()
        } else {
            webSocketManager.disconnect()
        }
    }

    fun start() {
        webSocketSyncBridge.start()
        authManager.addUserChangeListener(authListener)
        if (authManager.isSignedIn()) {
            webSocketManager.connect()
        }
    }

    fun stop() {
        webSocketSyncBridge.stop()
        authManager.removeUserChangeListener(authListener)
        webSocketManager.disconnect()
    }
}
