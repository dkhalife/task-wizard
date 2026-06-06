package app.taskwiz.auth

interface AuthTokenProvider {
    suspend fun getAccessToken(forceRefresh: Boolean = false): String?
    fun getCachedAccessToken(): String?
}
