package com.dkhalife.tasks.ui.screen

import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.automirrored.filled.Label
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExtendedFloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.LargeTopAppBar
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.input.nestedscroll.nestedScroll
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.R
import com.dkhalife.tasks.model.Label
import com.dkhalife.tasks.ui.components.LabelDialog
import com.dkhalife.tasks.ui.components.LabelItem

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LabelsScreen(
    labels: List<Label>,
    isRefreshing: Boolean,
    onRefresh: () -> Unit,
    onCreateLabel: (name: String, color: String) -> Unit,
    onUpdateLabel: (id: Int, name: String, color: String) -> Unit,
    onDeleteLabel: (id: Int) -> Unit
) {
    var showDialog by remember { mutableStateOf(false) }
    var editingLabel by remember { mutableStateOf<Label?>(null) }
    val scrollBehavior = TopAppBarDefaults.exitUntilCollapsedScrollBehavior()
    val newLabelLabel = stringResource(R.string.btn_new_label)

    Scaffold(
        modifier = Modifier.nestedScroll(scrollBehavior.nestedScrollConnection),
        topBar = {
            LargeTopAppBar(
                title = { Text(stringResource(R.string.nav_labels)) },
                scrollBehavior = scrollBehavior,
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.background,
                    scrolledContainerColor = MaterialTheme.colorScheme.surface
                )
            )
        },
        floatingActionButton = {
            ExtendedFloatingActionButton(
                onClick = {
                    editingLabel = null
                    showDialog = true
                },
                icon = { Icon(Icons.Default.Add, contentDescription = null) },
                text = { Text(newLabelLabel) },
                containerColor = MaterialTheme.colorScheme.primaryContainer,
                contentColor = MaterialTheme.colorScheme.onPrimaryContainer
            )
        }
    ) { padding ->
        PullToRefreshBox(
            isRefreshing = isRefreshing,
            onRefresh = onRefresh,
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            AnimatedContent(
                targetState = labels.isEmpty() && !isRefreshing,
                transitionSpec = { fadeIn() togetherWith fadeOut() },
                label = "labels-content"
            ) { isEmpty ->
                if (isEmpty) {
                    Box(
                        modifier = Modifier.fillMaxSize(),
                        contentAlignment = Alignment.Center
                    ) {
                        Column(horizontalAlignment = Alignment.CenterHorizontally) {
                            Icon(
                                Icons.AutoMirrored.Filled.Label,
                                contentDescription = null,
                                modifier = Modifier.size(64.dp),
                                tint = MaterialTheme.colorScheme.primary.copy(alpha = 0.6f)
                            )
                            Spacer(modifier = Modifier.height(16.dp))
                            Text(
                                text = stringResource(R.string.labels_empty_title),
                                style = MaterialTheme.typography.titleMedium,
                                color = MaterialTheme.colorScheme.onSurface
                            )
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(
                                text = stringResource(R.string.labels_empty_hint, newLabelLabel),
                                style = MaterialTheme.typography.bodyMedium,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    }
                } else {
                    LazyColumn(
                        modifier = Modifier.fillMaxSize(),
                        contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp),
                        verticalArrangement = Arrangement.spacedBy(6.dp)
                    ) {
                        items(labels, key = { it.id }) { label ->
                            LabelItem(
                                label = label,
                                onEdit = {
                                    editingLabel = label
                                    showDialog = true
                                },
                                onDelete = { onDeleteLabel(label.id) }
                            )
                        }
                    }
                }
            }
        }
    }

    if (showDialog) {
        LabelDialog(
            existingLabel = editingLabel,
            onDismiss = { showDialog = false },
            onSave = { name, color ->
                if (editingLabel != null) {
                    onUpdateLabel(editingLabel!!.id, name, color)
                } else {
                    onCreateLabel(name, color)
                }
                showDialog = false
            }
        )
    }
}
