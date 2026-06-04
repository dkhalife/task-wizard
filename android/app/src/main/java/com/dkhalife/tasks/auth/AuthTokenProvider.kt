package com.dkhalife.tasks.auth

interface AuthTokenProvider {
    suspend fun getAccessToken(forceRefresh: Boolean = false): String?
    fun getCachedAccessToken(): String?
}
