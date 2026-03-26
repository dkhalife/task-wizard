package com.dkhalife.tasks.ui.widget

import androidx.compose.ui.graphics.Color
import androidx.glance.material3.ColorProviders
import androidx.glance.unit.ColorProvider
import com.dkhalife.tasks.data.TaskGroupColors
import com.dkhalife.tasks.ui.theme.*

object WidgetTheme {
    val colors = ColorProviders(
        light = androidx.compose.material3.lightColorScheme(
            primary = PrimaryLight,
            onPrimary = OnPrimaryLight,
            primaryContainer = PrimaryContainerLight,
            onPrimaryContainer = OnPrimaryContainerLight,
            secondary = SecondaryLight,
            onSecondary = OnSecondaryLight,
            secondaryContainer = SecondaryContainerLight,
            onSecondaryContainer = OnSecondaryContainerLight,
            tertiary = TertiaryLight,
            onTertiary = OnTertiaryLight,
            tertiaryContainer = TertiaryContainerLight,
            onTertiaryContainer = OnTertiaryContainerLight,
            error = ErrorLight,
            onError = OnErrorLight,
            errorContainer = ErrorContainerLight,
            onErrorContainer = OnErrorContainerLight,
            background = BackgroundLight,
            onBackground = OnBackgroundLight,
            surface = SurfaceLight,
            onSurface = OnSurfaceLight,
            surfaceVariant = SurfaceVariantLight,
            onSurfaceVariant = OnSurfaceVariantLight,
            outline = OutlineLight,
            outlineVariant = OutlineVariantLight,
        ),
        dark = androidx.compose.material3.darkColorScheme(
            primary = PrimaryDark,
            onPrimary = OnPrimaryDark,
            primaryContainer = PrimaryContainerDark,
            onPrimaryContainer = OnPrimaryContainerDark,
            secondary = SecondaryDark,
            onSecondary = OnSecondaryDark,
            secondaryContainer = SecondaryContainerDark,
            onSecondaryContainer = OnSecondaryContainerDark,
            tertiary = TertiaryDark,
            onTertiary = OnTertiaryDark,
            tertiaryContainer = TertiaryContainerDark,
            onTertiaryContainer = OnTertiaryContainerDark,
            error = ErrorDark,
            onError = OnErrorDark,
            errorContainer = ErrorContainerDark,
            onErrorContainer = OnErrorContainerDark,
            background = BackgroundDark,
            onBackground = OnBackgroundDark,
            surface = SurfaceDark,
            onSurface = OnSurfaceDark,
            surfaceVariant = SurfaceVariantDark,
            onSurfaceVariant = OnSurfaceVariantDark,
            outline = OutlineDark,
            outlineVariant = OutlineVariantDark,
        )
    )

    fun groupColor(key: String): ColorProvider {
        val color = when (key) {
            "overdue" -> TaskGroupColors.OVERDUE
            "today" -> TaskGroupColors.TODAY
            "tomorrow" -> TaskGroupColors.TOMORROW
            "this_week" -> TaskGroupColors.THIS_WEEK
            "next_week" -> TaskGroupColors.NEXT_WEEK
            "later" -> TaskGroupColors.LATER
            "any_time" -> TaskGroupColors.ANY_TIME
            else -> Color.Gray
        }
        return ColorProvider(color)
    }
}
