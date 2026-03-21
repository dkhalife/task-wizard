package com.dkhalife.tasks.auth

interface AuthTokenProvider {
    suspend fun getAccessToken(): String?
    fun getCachedAccessToken(): String?
}
