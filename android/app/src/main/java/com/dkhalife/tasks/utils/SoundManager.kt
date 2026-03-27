package com.dkhalife.tasks.utils

import android.content.Context
import android.media.MediaPlayer
import androidx.annotation.RawRes
import com.dkhalife.tasks.R
import com.dkhalife.tasks.telemetry.TelemetryManager
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

enum class SoundEffect(@RawRes val resourceId: Int) {
    TASK_COMPLETE(R.raw.ding)
}

@Singleton
class SoundManager @Inject constructor(
    @ApplicationContext private val context: Context,
    private val telemetryManager: TelemetryManager
) {
    private val mediaPlayers = mutableMapOf<SoundEffect, MediaPlayer>()

    init {
        preloadSounds()
    }

    private fun preloadSounds() {
        SoundEffect.entries.forEach { effect ->
            try {
                val mediaPlayer = MediaPlayer.create(context, effect.resourceId)
                if (mediaPlayer != null) {
                    mediaPlayers[effect] = mediaPlayer
                } else {
                    telemetryManager.logError(TAG, "Failed to preload sound ${effect.name}: MediaPlayer.create returned null")
                }
            } catch (e: Exception) {
                telemetryManager.logError(TAG, "Failed to preload sound ${effect.name}: ${e.message}", e)
            }
        }
    }

    fun playSound(effect: SoundEffect) {
        try {
            val mediaPlayer = mediaPlayers[effect] ?: MediaPlayer.create(context, effect.resourceId)

            if (mediaPlayer == null) {
                telemetryManager.logError(TAG, "Failed to create player for ${effect.name}: MediaPlayer.create returned null")
                return
            }

            if (!mediaPlayers.containsKey(effect)) {
                mediaPlayers[effect] = mediaPlayer
            }
            
            if (mediaPlayer.isPlaying) {
                mediaPlayer.seekTo(0)
            } else {
                mediaPlayer.start()
            }
        } catch (e: Exception) {
            telemetryManager.logError(TAG, "Failed to play sound ${effect.name}: ${e.message}", e)
        }
    }

    fun release() {
        mediaPlayers.values.forEach { player ->
            try {
                if (player.isPlaying) {
                    player.stop()
                }
                player.release()
            } catch (e: Exception) {
                telemetryManager.logError(TAG, "Failed to release media player: ${e.message}", e)
            }
        }
        mediaPlayers.clear()
    }

    companion object {
        private const val TAG = "SoundManager"
    }
}
