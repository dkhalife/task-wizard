package com.dkhalife.tasks.ui.components

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ExpandLess
import androidx.compose.material.icons.filled.ExpandMore
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.data.TaskGroup

@Composable
fun GroupHeader(group: TaskGroup, isExpanded: Boolean, onToggle: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onToggle)
            .padding(vertical = 4.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        val headerColor = if (group.color == Color.Unspecified) {
            MaterialTheme.colorScheme.onSurface
        } else {
            group.color
        }

        val indicatorColor = if (group.color == Color.Unspecified) {
            MaterialTheme.colorScheme.onSurfaceVariant
        } else {
            group.color
        }

        Surface(
            shape = MaterialTheme.shapes.extraSmall,
            color = indicatorColor.copy(alpha = 0.2f),
            modifier = Modifier.size(12.dp)
        ) {}

        Spacer(modifier = Modifier.width(8.dp))

        Text(
            text = group.name,
            style = MaterialTheme.typography.titleSmall,
            color = headerColor
        )

        Spacer(modifier = Modifier.width(8.dp))

        Text(
            text = "(${group.tasks.size})",
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )

        Spacer(modifier = Modifier.weight(1f))

        Icon(
            imageVector = if (isExpanded) Icons.Default.ExpandLess else Icons.Default.ExpandMore,
            contentDescription = if (isExpanded) "Collapse" else "Expand",
            tint = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.size(20.dp)
        )
    }
}
