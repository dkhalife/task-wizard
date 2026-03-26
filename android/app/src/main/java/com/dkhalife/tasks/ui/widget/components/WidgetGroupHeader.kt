package com.dkhalife.tasks.ui.widget.components

import androidx.compose.runtime.Composable
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.glance.GlanceModifier
import androidx.glance.GlanceTheme
import androidx.glance.layout.Row
import androidx.glance.layout.fillMaxWidth
import androidx.glance.layout.padding
import androidx.glance.text.FontWeight
import androidx.glance.text.Text
import androidx.glance.text.TextStyle
import com.dkhalife.tasks.ui.widget.WidgetTheme

@Composable
fun WidgetGroupHeader(
    name: String,
    count: Int,
    groupKey: String,
    modifier: GlanceModifier = GlanceModifier,
) {
    val color = WidgetTheme.groupColor(groupKey)

    Row(
        modifier = modifier
            .fillMaxWidth()
            .padding(horizontal = 12.dp, vertical = 6.dp)
    ) {
        Text(
            text = "$name ($count)",
            style = TextStyle(
                color = color,
                fontSize = 13.sp,
                fontWeight = FontWeight.Bold
            )
        )
    }
}
