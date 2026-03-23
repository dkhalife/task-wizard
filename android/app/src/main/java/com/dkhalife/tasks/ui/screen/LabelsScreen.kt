package com.dkhalife.tasks.ui.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
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

    Scaffold(
        floatingActionButton = {
            FloatingActionButton(onClick = {
                editingLabel = null
                showDialog = true
            }) {
                Icon(Icons.Default.Add, contentDescription = "Create label")
            }
        }
    ) { padding ->
        PullToRefreshBox(
            isRefreshing = isRefreshing,
            onRefresh = onRefresh,
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            if (labels.isEmpty() && !isRefreshing) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "No labels yet. Tap + to create one.",
                        style = MaterialTheme.typography.bodyLarge,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            } else {
                LazyColumn(
                    modifier = Modifier.fillMaxSize(),
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
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
