package com.dkhalife.tasks.utils

import android.content.Context
import android.media.MediaPlayer
import android.util.Log
import androidx.annotation.RawRes
import com.dkhalife.tasks.R
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

enum class SoundEffect(@RawRes val resourceId: Int) {
    TASK_COMPLETE(R.raw.ding)
}

@Singleton
class SoundManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    private val mediaPlayers = mutableMapOf<SoundEffect, MediaPlayer>()

    init {
        preloadSounds()
    }

    private fun preloadSounds() {
        SoundEffect.entries.forEach { effect ->
            try {
                val mediaPlayer = MediaPlayer.create(context, effect.resourceId)
                mediaPlayers[effect] = mediaPlayer
            } catch (e: Exception) {
                Log.e("SoundManager", "Failed to preload sound ${effect.name}", e)
            }
        }
    }

    fun playSound(effect: SoundEffect) {
        try {
            val mediaPlayer = mediaPlayers[effect] ?: MediaPlayer.create(context, effect.resourceId)
            
            if (!mediaPlayers.containsKey(effect)) {
                mediaPlayers[effect] = mediaPlayer
            }
            
            if (mediaPlayer.isPlaying) {
                mediaPlayer.seekTo(0)
            } else {
                mediaPlayer.start()
            }
        } catch (e: Exception) {
            Log.e("SoundManager", "Failed to play sound ${effect.name}", e)
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
                Log.e("SoundManager", "Failed to release media player", e)
            }
        }
        mediaPlayers.clear()
    }
}
