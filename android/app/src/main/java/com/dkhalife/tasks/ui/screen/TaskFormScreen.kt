package com.dkhalife.tasks.ui.screen

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.model.*

@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun TaskFormScreen(
    existingTask: Task?,
    availableLabels: List<Label>,
    isSaving: Boolean,
    onSave: (title: String, nextDueDate: String?, frequency: Frequency, selectedLabelIds: List<Int>, isRolling: Boolean) -> Unit,
    onBack: () -> Unit
) {
    var title by remember(existingTask) { mutableStateOf(existingTask?.title ?: "") }
    var frequencyType by remember(existingTask) { mutableStateOf(existingTask?.frequency?.type ?: FrequencyType.ONCE) }
    var isRolling by remember(existingTask) { mutableStateOf(existingTask?.isRolling ?: false) }
    var selectedLabelIds by remember(existingTask) {
        mutableStateOf(existingTask?.labels?.map { it.id }?.toSet() ?: emptySet())
    }

    val isEditing = existingTask != null

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(if (isEditing) "Edit Task" else "New Task") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(16.dp)
                .verticalScroll(rememberScrollState()),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            OutlinedTextField(
                value = title,
                onValueChange = { title = it },
                label = { Text("Title") },
                singleLine = true,
                modifier = Modifier.fillMaxWidth()
            )

            Text("Frequency", style = MaterialTheme.typography.titleSmall)

            SingleChoiceSegmentedButtonRow(modifier = Modifier.fillMaxWidth()) {
                val options = listOf(FrequencyType.ONCE, FrequencyType.DAILY, FrequencyType.WEEKLY, FrequencyType.MONTHLY)
                options.forEachIndexed { index, option ->
                    SegmentedButton(
                        selected = frequencyType == option,
                        onClick = { frequencyType = option },
                        shape = SegmentedButtonDefaults.itemShape(index, options.size)
                    ) {
                        Text(option.replaceFirstChar { it.uppercase() })
                    }
                }
            }

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text("Rolling deadline", style = MaterialTheme.typography.bodyLarge)
                Switch(checked = isRolling, onCheckedChange = { isRolling = it })
            }

            if (availableLabels.isNotEmpty()) {
                Text("Labels", style = MaterialTheme.typography.titleSmall)
                FlowRow(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    availableLabels.forEach { label ->
                        FilterChip(
                            selected = label.id in selectedLabelIds,
                            onClick = {
                                selectedLabelIds = if (label.id in selectedLabelIds) {
                                    selectedLabelIds - label.id
                                } else {
                                    selectedLabelIds + label.id
                                }
                            },
                            label = { Text(label.name) }
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.weight(1f))

            Button(
                onClick = {
                    val frequency = Frequency(type = frequencyType)
                    onSave(title, null, frequency, selectedLabelIds.toList(), isRolling)
                },
                enabled = title.isNotBlank() && !isSaving,
                modifier = Modifier.fillMaxWidth()
            ) {
                if (isSaving) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(20.dp),
                        strokeWidth = 2.dp,
                        color = MaterialTheme.colorScheme.onPrimary
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                }
                Text(if (isEditing) "Update Task" else "Create Task")
            }
        }
    }
}
