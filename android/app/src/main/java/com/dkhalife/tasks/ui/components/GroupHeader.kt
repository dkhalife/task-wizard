package com.dkhalife.tasks.ui.components

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ExpandMore
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.R
import com.dkhalife.tasks.data.TaskGroup

@Composable
fun GroupHeader(group: TaskGroup, isExpanded: Boolean, onToggle: () -> Unit) {
    val rotation by animateFloatAsState(
        targetValue = if (isExpanded) 0f else -90f,
        label = "chevron-rotation"
    )

    val headerColor = if (group.color == Color.Unspecified) {
        MaterialTheme.colorScheme.onSurface
    } else {
        group.color
    }

    val indicatorColor = if (group.color == Color.Unspecified) {
        MaterialTheme.colorScheme.outline
    } else {
        group.color
    }

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onToggle)
            .padding(vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Surface(
            shape = RoundedCornerShape(4.dp),
            color = indicatorColor,
            modifier = Modifier
                .width(4.dp)
                .height(20.dp)
        ) {}

        Spacer(modifier = Modifier.width(12.dp))

        Text(
            text = group.name,
            style = MaterialTheme.typography.titleSmall,
            color = headerColor
        )

        Spacer(modifier = Modifier.width(8.dp))

        Surface(
            shape = RoundedCornerShape(12.dp),
            color = indicatorColor.copy(alpha = 0.12f)
        ) {
            Text(
                text = "${group.tasks.size}",
                style = MaterialTheme.typography.labelSmall,
                color = indicatorColor,
                modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp)
            )
        }

        Spacer(modifier = Modifier.weight(1f))

        Icon(
            imageVector = Icons.Default.ExpandMore,
            contentDescription = if (isExpanded) stringResource(R.string.action_collapse) else stringResource(R.string.action_expand),
            tint = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier
                .size(20.dp)
                .rotate(rotation)
        )
    }
}
