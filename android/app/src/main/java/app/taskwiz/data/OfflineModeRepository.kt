package app.taskwiz.data

import android.content.SharedPreferences
import androidx.core.content.edit
import app.taskwiz.model.AuthConfig
import com.google.gson.Gson
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class OfflineModeRepository @Inject constructor(
    private val sharedPreferences: SharedPreferences,
    private val gson: Gson,
) {
    fun hasSignedIn(): Boolean {
        return sharedPreferences.getBoolean(AppPreferences.KEY_HAS_SIGNED_IN, false)
    }

    fun markSignedIn() {
        sharedPreferences.edit { putBoolean(AppPreferences.KEY_HAS_SIGNED_IN, true) }
    }

    fun isSyncPromptShown(): Boolean {
        return sharedPreferences.getBoolean(AppPreferences.KEY_SYNC_PROMPT_SHOWN, false)
    }

    fun markSyncPromptShown() {
        sharedPreferences.edit { putBoolean(AppPreferences.KEY_SYNC_PROMPT_SHOWN, true) }
    }

    fun getCachedAuthConfig(): AuthConfig? {
        val json = sharedPreferences.getString(AppPreferences.KEY_CACHED_AUTH_CONFIG, null)
            ?: return null
        return try {
            gson.fromJson(json, AuthConfig::class.java)
        } catch (_: Exception) {
            null
        }
    }

    fun setCachedAuthConfig(config: AuthConfig) {
        sharedPreferences.edit {
            putString(AppPreferences.KEY_CACHED_AUTH_CONFIG, gson.toJson(config))
        }
    }

    fun clearCachedAuthConfig() {
        sharedPreferences.edit { remove(AppPreferences.KEY_CACHED_AUTH_CONFIG) }
    }
}
