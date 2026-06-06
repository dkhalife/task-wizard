package app.taskwiz.data.db

import androidx.room.TypeConverter
import app.taskwiz.model.Frequency
import app.taskwiz.model.NotificationTriggerOptions
import com.google.gson.Gson

class Converters {
    private val gson = Gson()

    @TypeConverter
    fun frequencyToJson(value: Frequency?): String? = value?.let { gson.toJson(it) }

    @TypeConverter
    fun frequencyFromJson(value: String?): Frequency? =
        value?.let { gson.fromJson(it, Frequency::class.java) }

    @TypeConverter
    fun notificationToJson(value: NotificationTriggerOptions?): String? = value?.let { gson.toJson(it) }

    @TypeConverter
    fun notificationFromJson(value: String?): NotificationTriggerOptions? =
        value?.let { gson.fromJson(it, NotificationTriggerOptions::class.java) }
}
