package com.dkhalife.tasks.ui.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Block
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.SkipNext
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.LargeTopAppBar
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.input.nestedscroll.nestedScroll
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.R
import com.dkhalife.tasks.data.SwipeAction
import com.dkhalife.tasks.data.SwipeSettings

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SwipeActionsSettingsScreen(
    swipeSettings: SwipeSettings,
    onStartToEndActionChanged: (SwipeAction) -> Unit,
    onEndToStartActionChanged: (SwipeAction) -> Unit,
    onBack: () -> Unit
) {
    val scrollBehavior = TopAppBarDefaults.exitUntilCollapsedScrollBehavior()

    Scaffold(
        modifier = Modifier.nestedScroll(scrollBehavior.nestedScrollConnection),
        topBar = {
            LargeTopAppBar(
                title = { Text(stringResource(R.string.swipe_settings_title)) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(
                            Icons.AutoMirrored.Filled.ArrowBack,
                            contentDescription = stringResource(R.string.btn_back)
                        )
                    }
                },
                scrollBehavior = scrollBehavior,
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.background,
                    scrolledContainerColor = MaterialTheme.colorScheme.surface
                )
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(horizontal = 16.dp)
                .verticalScroll(rememberScrollState()),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Spacer(modifier = Modifier.height(4.dp))

            SwipePreviewCard(swipeSettings = swipeSettings)

            SwipeActionPickerCard(
                title = stringResource(R.string.swipe_settings_right_action_title),
                selectedAction = swipeSettings.startToEndAction,
                availableActions = listOf(SwipeAction.COMPLETE, SwipeAction.SKIP, SwipeAction.NONE),
                onActionSelected = onStartToEndActionChanged
            )

            SwipeActionPickerCard(
                title = stringResource(R.string.swipe_settings_left_action_title),
                selectedAction = swipeSettings.endToStartAction,
                availableActions = listOf(SwipeAction.DELETE, SwipeAction.COMPLETE, SwipeAction.SKIP, SwipeAction.NONE),
                onActionSelected = onEndToStartActionChanged
            )

            Spacer(modifier = Modifier.height(16.dp))
        }
    }
}

@Composable
private fun SwipePreviewCard(swipeSettings: SwipeSettings) {
    Card(
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f)
        )
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = stringResource(R.string.swipe_settings_preview_title),
                style = MaterialTheme.typography.titleMedium
            )
            Spacer(modifier = Modifier.height(12.dp))
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                SwipeDirectionPreview(
                    label = stringResource(R.string.swipe_settings_preview_swipe_right),
                    action = swipeSettings.startToEndAction,
                    isStartToEnd = true,
                    modifier = Modifier.weight(1f)
                )
                SwipeDirectionPreview(
                    label = stringResource(R.string.swipe_settings_preview_swipe_left),
                    action = swipeSettings.endToStartAction,
                    isStartToEnd = false,
                    modifier = Modifier.weight(1f)
                )
            }
        }
    }
}

@Composable
private fun SwipeDirectionPreview(
    label: String,
    action: SwipeAction,
    isStartToEnd: Boolean,
    modifier: Modifier = Modifier
) {
    val (icon, containerColor, contentColor) = swipeActionVisuals(action)

    Column(
        modifier = modifier,
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        Text(
            text = label,
            style = MaterialTheme.typography.labelMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            textAlign = TextAlign.Center
        )

        Card(
            shape = RoundedCornerShape(12.dp),
            colors = CardDefaults.cardColors(containerColor = containerColor)
        ) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp)
                    .padding(horizontal = 12.dp),
                contentAlignment = if (isStartToEnd) Alignment.CenterStart else Alignment.CenterEnd
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                    if (isStartToEnd) {
                        Icon(icon, contentDescription = null, tint = contentColor, modifier = Modifier.size(20.dp))
                        Text(
                            text = swipeActionLabel(action),
                            style = MaterialTheme.typography.labelSmall,
                            color = contentColor
                        )
                    } else {
                        Text(
                            text = swipeActionLabel(action),
                            style = MaterialTheme.typography.labelSmall,
                            color = contentColor
                        )
                        Icon(icon, contentDescription = null, tint = contentColor, modifier = Modifier.size(20.dp))
                    }
                }
            }
        }
    }
}

@Composable
private fun SwipeActionPickerCard(
    title: String,
    selectedAction: SwipeAction,
    availableActions: List<SwipeAction>,
    onActionSelected: (SwipeAction) -> Unit
) {
    Card(
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f)
        )
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(title, style = MaterialTheme.typography.titleMedium)
            Spacer(modifier = Modifier.height(12.dp))
            Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                availableActions.forEach { action ->
                    SwipeActionOption(
                        action = action,
                        isSelected = action == selectedAction,
                        onClick = { onActionSelected(action) }
                    )
                }
            }
        }
    }
}

@Composable
private fun SwipeActionOption(
    action: SwipeAction,
    isSelected: Boolean,
    onClick: () -> Unit
) {
    val (icon, containerColor, contentColor) = swipeActionVisuals(action)
    val selectedContainerColor = if (isSelected)
        MaterialTheme.colorScheme.secondaryContainer
    else
        MaterialTheme.colorScheme.surface

    Card(
        shape = RoundedCornerShape(10.dp),
        colors = CardDefaults.cardColors(containerColor = selectedContainerColor),
        onClick = onClick
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 12.dp, vertical = 10.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Card(
                shape = RoundedCornerShape(8.dp),
                colors = CardDefaults.cardColors(containerColor = containerColor),
                modifier = Modifier.size(36.dp)
            ) {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    Icon(icon, contentDescription = null, tint = contentColor, modifier = Modifier.size(18.dp))
                }
            }

            Text(
                text = swipeActionLabel(action),
                style = MaterialTheme.typography.bodyMedium,
                modifier = Modifier.weight(1f)
            )

            if (isSelected) {
                Icon(
                    Icons.Default.Check,
                    contentDescription = null,
                    tint = MaterialTheme.colorScheme.secondary,
                    modifier = Modifier.size(18.dp)
                )
            } else {
                Spacer(modifier = Modifier.width(18.dp))
            }
        }
    }
}

private data class SwipeActionVisuals(
    val icon: ImageVector,
    val containerColor: Color,
    val contentColor: Color
)

@Composable
private fun swipeActionVisuals(action: SwipeAction): SwipeActionVisuals = when (action) {
    SwipeAction.COMPLETE -> SwipeActionVisuals(
        icon = Icons.Default.Check,
        containerColor = MaterialTheme.colorScheme.primaryContainer,
        contentColor = MaterialTheme.colorScheme.onPrimaryContainer
    )
    SwipeAction.DELETE -> SwipeActionVisuals(
        icon = Icons.Default.Delete,
        containerColor = MaterialTheme.colorScheme.errorContainer,
        contentColor = MaterialTheme.colorScheme.onErrorContainer
    )
    SwipeAction.SKIP -> SwipeActionVisuals(
        icon = Icons.Default.SkipNext,
        containerColor = MaterialTheme.colorScheme.tertiaryContainer,
        contentColor = MaterialTheme.colorScheme.onTertiaryContainer
    )
    SwipeAction.NONE -> SwipeActionVisuals(
        icon = Icons.Default.Block,
        containerColor = MaterialTheme.colorScheme.surfaceVariant,
        contentColor = MaterialTheme.colorScheme.onSurfaceVariant
    )
}

@Composable
private fun swipeActionLabel(action: SwipeAction): String = when (action) {
    SwipeAction.COMPLETE -> stringResource(R.string.swipe_action_complete)
    SwipeAction.DELETE -> stringResource(R.string.swipe_action_delete)
    SwipeAction.SKIP -> stringResource(R.string.swipe_action_skip)
    SwipeAction.NONE -> stringResource(R.string.swipe_action_none)
}
