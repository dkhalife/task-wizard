package com.dkhalife.tasks.data.network

import android.content.Context
import android.net.ConnectivityManager
import android.net.Network
import android.net.NetworkCapabilities
import android.net.NetworkRequest
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import java.util.concurrent.CopyOnWriteArrayList
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class NetworkMonitor @Inject constructor(
    @ApplicationContext context: Context,
) {
    private val cm = context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager

    private val _isOnline = MutableStateFlow(currentlyOnline())
    val isOnline: StateFlow<Boolean> = _isOnline.asStateFlow()

    private val listeners = CopyOnWriteArrayList<() -> Unit>()

    init {
        val request = NetworkRequest.Builder()
            .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
            .build()
        cm.registerNetworkCallback(request, object : ConnectivityManager.NetworkCallback() {
            override fun onAvailable(network: Network) {
                val previouslyOffline = !_isOnline.value
                val onlineNow = currentlyOnline()
                _isOnline.value = onlineNow
                if (previouslyOffline && onlineNow) {
                    listeners.forEach { runCatching { it() } }
                }
            }

            override fun onLost(network: Network) {
                _isOnline.value = currentlyOnline()
            }

            override fun onCapabilitiesChanged(network: Network, caps: NetworkCapabilities) {
                val previouslyOffline = !_isOnline.value
                val onlineNow = currentlyOnline()
                _isOnline.value = onlineNow
                if (previouslyOffline && onlineNow) {
                    listeners.forEach { runCatching { it() } }
                }
            }
        })
    }

    fun addOnAvailableListener(block: () -> Unit) {
        listeners.add(block)
    }

    private fun currentlyOnline(): Boolean {
        val active = cm.activeNetwork ?: return false
        val caps = cm.getNetworkCapabilities(active) ?: return false
        return caps.hasCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET) &&
            caps.hasCapability(NetworkCapabilities.NET_CAPABILITY_VALIDATED)
    }
}
