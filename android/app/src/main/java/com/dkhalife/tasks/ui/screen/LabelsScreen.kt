package com.dkhalife.tasks.ui.screen

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material3.*
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.model.Label

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

@Composable
private fun LabelItem(
    label: Label,
    onEdit: () -> Unit,
    onDelete: () -> Unit
) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            val color = try {
                Color(android.graphics.Color.parseColor(label.color))
            } catch (_: Exception) {
                MaterialTheme.colorScheme.primary
            }

            Box(
                modifier = Modifier
                    .size(24.dp)
                    .clip(CircleShape)
                    .background(color)
            )

            Spacer(modifier = Modifier.width(16.dp))

            Text(
                text = label.name,
                style = MaterialTheme.typography.bodyLarge,
                modifier = Modifier.weight(1f)
            )

            IconButton(onClick = onEdit) {
                Icon(Icons.Default.Edit, contentDescription = "Edit")
            }

            IconButton(onClick = onDelete) {
                Icon(
                    Icons.Default.Delete,
                    contentDescription = "Delete",
                    tint = MaterialTheme.colorScheme.error
                )
            }
        }
    }
}

@Composable
private fun LabelDialog(
    existingLabel: Label?,
    onDismiss: () -> Unit,
    onSave: (name: String, color: String) -> Unit
) {
    var name by remember { mutableStateOf(existingLabel?.name ?: "") }
    var color by remember { mutableStateOf(existingLabel?.color ?: "#6650a4") }

    val presetColors = listOf("#6650a4", "#625b71", "#7D5260", "#B3261E", "#006D3B", "#0061A4", "#984061", "#7D5700")

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(if (existingLabel != null) "Edit Label" else "New Label") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                OutlinedTextField(
                    value = name,
                    onValueChange = { name = it },
                    label = { Text("Name") },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth()
                )

                Text("Color", style = MaterialTheme.typography.titleSmall)

                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    presetColors.forEach { preset ->
                        val presetColor = try {
                            Color(android.graphics.Color.parseColor(preset))
                        } catch (_: Exception) {
                            MaterialTheme.colorScheme.primary
                        }

                        Surface(
                            onClick = { color = preset },
                            shape = CircleShape,
                            color = presetColor,
                            modifier = Modifier.size(32.dp),
                        ) {
                            if (color.equals(preset, ignoreCase = true)) {
                                Box(
                                    modifier = Modifier.fillMaxSize(),
                                    contentAlignment = Alignment.Center
                                ) {
                                    Icon(
                                        Icons.Default.Check,
                                        contentDescription = "Selected",
                                        tint = Color.White,
                                        modifier = Modifier.size(18.dp)
                                    )
                                }
                            }
                        }
                    }
                }
            }
        },
        confirmButton = {
            TextButton(
                onClick = { onSave(name, color) },
                enabled = name.isNotBlank()
            ) {
                Text("Save")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("Cancel")
            }
        }
    )
}
