package com.dkhalife.tasks.ui.screen

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.CalendarMonth
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.model.*
import java.time.ZoneId
import java.time.ZonedDateTime
import java.time.format.DateTimeFormatter
import java.util.Locale

private fun parseIsoDateTime(iso: String?): ZonedDateTime? = iso?.let {
    runCatching { ZonedDateTime.parse(it).withZoneSameInstant(ZoneId.systemDefault()) }.getOrNull()
}

private fun ZonedDateTime.toIsoString(): String =
    format(DateTimeFormatter.ISO_OFFSET_DATE_TIME)

private fun ZonedDateTime.toDisplayString(): String =
    format(DateTimeFormatter.ofPattern("MM/dd/yyyy, hh:mm a", Locale.ENGLISH))

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun DateTimePickerRow(
    label: String,
    value: ZonedDateTime?,
    onValueSelected: (ZonedDateTime) -> Unit
) {
    var showDatePicker by remember { mutableStateOf(false) }
    var showTimePicker by remember { mutableStateOf(false) }
    var pendingDate by remember { mutableStateOf<ZonedDateTime?>(null) }

    val datePickerState = rememberDatePickerState(
        initialSelectedDateMillis = (value ?: ZonedDateTime.now(ZoneId.systemDefault()))
            .toInstant().toEpochMilli()
    )
    val timePickerState = rememberTimePickerState(
        initialHour = value?.hour ?: ZonedDateTime.now(ZoneId.systemDefault()).hour,
        initialMinute = value?.minute ?: 0,
        is24Hour = false
    )

    OutlinedButton(
        onClick = { showDatePicker = true },
        modifier = Modifier.fillMaxWidth(),
        contentPadding = PaddingValues(horizontal = 16.dp, vertical = 12.dp)
    ) {
        Icon(Icons.Default.CalendarMonth, contentDescription = null)
        Spacer(Modifier.width(8.dp))
        Text(
            text = value?.toDisplayString() ?: label,
            modifier = Modifier.weight(1f)
        )
    }

    if (showDatePicker) {
        DatePickerDialog(
            onDismissRequest = { showDatePicker = false },
            confirmButton = {
                TextButton(onClick = {
                    showDatePicker = false
                    val millis = datePickerState.selectedDateMillis
                    if (millis != null) {
                        pendingDate = ZonedDateTime.ofInstant(
                            java.time.Instant.ofEpochMilli(millis),
                            ZoneId.systemDefault()
                        )
                        showTimePicker = true
                    }
                }) { Text("Next") }
            },
            dismissButton = {
                TextButton(onClick = { showDatePicker = false }) { Text("Cancel") }
            }
        ) {
            DatePicker(state = datePickerState)
        }
    }

    if (showTimePicker) {
        AlertDialog(
            onDismissRequest = { showTimePicker = false },
            title = { Text("Select time") },
            text = {
                Box(Modifier.fillMaxWidth(), contentAlignment = Alignment.Center) {
                    TimePicker(state = timePickerState)
                }
            },
            confirmButton = {
                TextButton(onClick = {
                    showTimePicker = false
                    pendingDate?.let { date ->
                        val result = date
                            .withHour(timePickerState.hour)
                            .withMinute(timePickerState.minute)
                            .withSecond(0)
                            .withNano(0)
                        onValueSelected(result)
                    }
                }) { Text("OK") }
            },
            dismissButton = {
                TextButton(onClick = { showTimePicker = false }) { Text("Cancel") }
            }
        )
    }
}

// ── Main TaskFormScreen ────────────────────────────────────────────────────────

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
        mutableStateOf(existingTask?.labels?.map { it.id }?.toSet() ?: emptySet<Int>())
    }

    var hasDueDate by remember(existingTask) { mutableStateOf(existingTask?.nextDueDate != null) }
    var dueDate by remember(existingTask) { mutableStateOf(parseIsoDateTime(existingTask?.nextDueDate)) }

    val isRecurring = frequencyType != FrequencyType.ONCE
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

            Text("Next due date", style = MaterialTheme.typography.titleSmall)
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                Checkbox(
                    checked = hasDueDate,
                    onCheckedChange = { checked ->
                        hasDueDate = checked
                        if (!checked) {
                            dueDate = null
                            frequencyType = FrequencyType.ONCE
                        } else if (dueDate == null) {
                            dueDate = ZonedDateTime.now(ZoneId.systemDefault())
                                .plusDays(1)
                                .withHour(9).withMinute(0).withSecond(0).withNano(0)
                        }
                    }
                )
                Text(
                    text = if (isRecurring) "When is the next first time this task is due?"
                           else "Give this task a due date",
                    style = MaterialTheme.typography.bodyMedium
                )
            }
            if (hasDueDate) {
                DateTimePickerRow(
                    label = "Select due date & time",
                    value = dueDate,
                    onValueSelected = { dueDate = it }
                )
            }

            if (hasDueDate) {
                Text("Repeat", style = MaterialTheme.typography.titleSmall)

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
            }

            if (isRecurring) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text("Rolling deadline", style = MaterialTheme.typography.bodyLarge)
                    Switch(checked = isRolling, onCheckedChange = { isRolling = it })
                }
            }

            Button(
                onClick = {
                    val frequency = Frequency(type = frequencyType)
                    onSave(title, dueDate?.toIsoString(), frequency, selectedLabelIds.toList(), isRolling)
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
