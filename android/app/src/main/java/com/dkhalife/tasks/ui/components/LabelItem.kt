package com.dkhalife.tasks.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.SwipeToDismissBox
import androidx.compose.material3.SwipeToDismissBoxValue
import androidx.compose.material3.Text
import androidx.compose.material3.rememberSwipeToDismissBoxState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.CustomAccessibilityAction
import androidx.compose.ui.semantics.customActions
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.unit.dp
import androidx.compose.ui.res.stringResource
import com.dkhalife.tasks.R
import com.dkhalife.tasks.model.Label

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LabelItem(
    label: Label,
    onEdit: () -> Unit,
    onDelete: () -> Unit
) {
    val color = try {
        Color(android.graphics.Color.parseColor(label.color))
    } catch (_: Exception) {
        MaterialTheme.colorScheme.primary
    }

    val accessibilityActions = listOf(
        CustomAccessibilityAction(stringResource(R.string.action_edit)) { onEdit(); true },
        CustomAccessibilityAction(stringResource(R.string.action_delete)) { onDelete(); true }
    )

    val dismissState = rememberSwipeToDismissBoxState()

    LaunchedEffect(dismissState.currentValue) {
        when (dismissState.currentValue) {
            SwipeToDismissBoxValue.StartToEnd -> {
                onEdit()
                dismissState.snapTo(SwipeToDismissBoxValue.Settled)
            }
            SwipeToDismissBoxValue.EndToStart -> onDelete()
            SwipeToDismissBoxValue.Settled -> {}
        }
    }

    SwipeToDismissBox(
        state = dismissState,
        backgroundContent = {
            val alignment = when (dismissState.dismissDirection) {
                SwipeToDismissBoxValue.StartToEnd -> Alignment.CenterStart
                else -> Alignment.CenterEnd
            }
            val icon = when (dismissState.dismissDirection) {
                SwipeToDismissBoxValue.StartToEnd -> Icons.Default.Edit
                else -> Icons.Default.Delete
            }
            val containerColor = when (dismissState.dismissDirection) {
                SwipeToDismissBoxValue.StartToEnd -> MaterialTheme.colorScheme.primaryContainer
                else -> MaterialTheme.colorScheme.errorContainer
            }
            val contentColor = when (dismissState.dismissDirection) {
                SwipeToDismissBoxValue.StartToEnd -> MaterialTheme.colorScheme.onPrimaryContainer
                else -> MaterialTheme.colorScheme.onErrorContainer
            }

            Card(
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = containerColor)
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(horizontal = 20.dp),
                    contentAlignment = alignment
                ) {
                    Icon(icon, contentDescription = null, tint = contentColor)
                }
            }
        }
    ) {
        Card(
            shape = RoundedCornerShape(16.dp),
            colors = CardDefaults.cardColors(
                containerColor = MaterialTheme.colorScheme.surface
            ),
            elevation = CardDefaults.cardElevation(defaultElevation = 1.dp),
            modifier = Modifier
                .fillMaxWidth()
                .semantics { customActions = accessibilityActions }
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Box(
                    modifier = Modifier
                        .width(5.dp)
                        .fillMaxHeight()
                        .clip(RoundedCornerShape(topStart = 16.dp, bottomStart = 16.dp))
                        .background(color)
                )

                Spacer(modifier = Modifier.width(16.dp))

                Text(
                    text = label.name,
                    style = MaterialTheme.typography.bodyLarge,
                    modifier = Modifier.weight(1f)
                )

                Spacer(modifier = Modifier.width(16.dp))
            }
        }
    }
}
