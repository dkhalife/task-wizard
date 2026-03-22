package com.dkhalife.tasks.ui.screen

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.CalendarMonth
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
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
    onSave: (title: String, nextDueDate: String?, endDate: String?, frequency: Frequency, notification: NotificationTriggerOptions, selectedLabelIds: List<Int>, isRolling: Boolean) -> Unit,
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

    var hasEndDate by remember(existingTask) { mutableStateOf(existingTask?.endDate != null) }
    var endDate by remember(existingTask) { mutableStateOf(parseIsoDateTime(existingTask?.endDate)) }

    var repeatOn by remember(existingTask) {
        mutableStateOf(existingTask?.frequency?.on ?: RepeatOn.INTERVAL)
    }
    var intervalEvery by remember(existingTask) {
        mutableStateOf(existingTask?.frequency?.every?.toString() ?: "1")
    }
    var intervalUnit by remember(existingTask) {
        mutableStateOf(existingTask?.frequency?.unit ?: IntervalUnit.WEEKS)
    }
    var selectedDays by remember(existingTask) {
        val today = ZonedDateTime.now(ZoneId.systemDefault()).dayOfWeek.value % 7
        mutableStateOf(existingTask?.frequency?.days?.toSet() ?: setOf(today))
    }
    var selectedMonths by remember(existingTask) {
        val thisMonth = ZonedDateTime.now(ZoneId.systemDefault()).monthValue - 1
        mutableStateOf(existingTask?.frequency?.months?.toSet() ?: setOf(thisMonth))
    }

    var notificationsEnabled by remember(existingTask) {
        mutableStateOf(existingTask?.notification?.enabled ?: false)
    }
    var notifyDueDate by remember(existingTask) {
        mutableStateOf(existingTask?.notification?.dueDate ?: true)
    }
    var notifyPreDue by remember(existingTask) {
        mutableStateOf(existingTask?.notification?.preDue ?: false)
    }
    var notifyOverdue by remember(existingTask) {
        mutableStateOf(existingTask?.notification?.overdue ?: false)
    }

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

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    Checkbox(
                        checked = isRecurring,
                        onCheckedChange = { checked ->
                            frequencyType = if (checked) FrequencyType.DAILY else FrequencyType.ONCE
                        }
                    )
                    Text("Repeat this task", style = MaterialTheme.typography.bodyMedium)
                }

                if (isRecurring) {
                    // Row 1: Daily, Weekly, Monthly
                    SingleChoiceSegmentedButtonRow(modifier = Modifier.fillMaxWidth()) {
                        val row1 = listOf(FrequencyType.DAILY, FrequencyType.WEEKLY, FrequencyType.MONTHLY)
                        row1.forEachIndexed { index, option ->
                            SegmentedButton(
                                selected = frequencyType == option,
                                onClick = { frequencyType = option },
                                shape = SegmentedButtonDefaults.itemShape(index, row1.size)
                            ) {
                                Text(option.replaceFirstChar { it.uppercase() })
                            }
                        }
                    }

                    // Row 2: Yearly, Custom
                    SingleChoiceSegmentedButtonRow(modifier = Modifier.fillMaxWidth()) {
                        val row2 = listOf(FrequencyType.YEARLY, FrequencyType.CUSTOM)
                        row2.forEachIndexed { index, option ->
                            SegmentedButton(
                                selected = frequencyType == option,
                                onClick = { frequencyType = option },
                                shape = SegmentedButtonDefaults.itemShape(index, row2.size)
                            ) {
                                Text(option.replaceFirstChar { it.uppercase() })
                            }
                        }
                    }

                    if (frequencyType == FrequencyType.CUSTOM) {
                        // Sub-mode selector
                        SingleChoiceSegmentedButtonRow(modifier = Modifier.fillMaxWidth()) {
                            val subModes = listOf(RepeatOn.INTERVAL, RepeatOn.DAYS_OF_THE_WEEK, RepeatOn.DAY_OF_THE_MONTHS)
                            val subModeLabels = listOf("Interval", "Days of the week", "Day of the months")
                            subModes.forEachIndexed { index, mode ->
                                SegmentedButton(
                                    selected = repeatOn == mode,
                                    onClick = { repeatOn = mode },
                                    shape = SegmentedButtonDefaults.itemShape(index, subModes.size)
                                ) {
                                    Text(subModeLabels[index])
                                }
                            }
                        }

                        when (repeatOn) {
                            RepeatOn.INTERVAL -> {
                                Row(
                                    modifier = Modifier.fillMaxWidth(),
                                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    Text("Every", style = MaterialTheme.typography.bodyMedium)
                                    OutlinedTextField(
                                        value = intervalEvery,
                                        onValueChange = { v ->
                                            if (v.isEmpty() || v.toIntOrNull()?.let { it in 1..365 } == true) {
                                                intervalEvery = v
                                            }
                                        },
                                        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                                        singleLine = true,
                                        modifier = Modifier.width(80.dp)
                                    )
                                    var unitExpanded by remember { mutableStateOf(false) }
                                    ExposedDropdownMenuBox(
                                        expanded = unitExpanded,
                                        onExpandedChange = { unitExpanded = it },
                                        modifier = Modifier.weight(1f)
                                    ) {
                                        OutlinedTextField(
                                            value = intervalUnit.replaceFirstChar { it.uppercase() },
                                            onValueChange = {},
                                            readOnly = true,
                                            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = unitExpanded) },
                                            modifier = Modifier.menuAnchor(ExposedDropdownMenuAnchorType.PrimaryNotEditable)
                                        )
                                        ExposedDropdownMenu(
                                            expanded = unitExpanded,
                                            onDismissRequest = { unitExpanded = false }
                                        ) {
                                            listOf(
                                                IntervalUnit.HOURS,
                                                IntervalUnit.DAYS,
                                                IntervalUnit.WEEKS,
                                                IntervalUnit.MONTHS,
                                                IntervalUnit.YEARS
                                            ).forEach { unit ->
                                                DropdownMenuItem(
                                                    text = { Text(unit.replaceFirstChar { it.uppercase() }) },
                                                    onClick = {
                                                        intervalUnit = unit
                                                        unitExpanded = false
                                                    },
                                                    contentPadding = ExposedDropdownMenuDefaults.ItemContentPadding
                                                )
                                            }
                                        }
                                    }
                                }
                            }

                            RepeatOn.DAYS_OF_THE_WEEK -> {
                                val dayNames = listOf("Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat")
                                FlowRow(
                                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                                    verticalArrangement = Arrangement.spacedBy(8.dp)
                                ) {
                                    dayNames.forEachIndexed { i, name ->
                                        FilterChip(
                                            selected = i in selectedDays,
                                            onClick = {
                                                selectedDays = if (i in selectedDays) {
                                                    if (selectedDays.size > 1) selectedDays - i else selectedDays
                                                } else {
                                                    selectedDays + i
                                                }
                                            },
                                            label = { Text(name) }
                                        )
                                    }
                                }
                            }

                            RepeatOn.DAY_OF_THE_MONTHS -> {
                                val dayOfMonth = dueDate?.dayOfMonth ?: 1
                                val ordinalSuffix = when {
                                    dayOfMonth in 11..13 -> "th"
                                    dayOfMonth % 10 == 1 -> "st"
                                    dayOfMonth % 10 == 2 -> "nd"
                                    dayOfMonth % 10 == 3 -> "rd"
                                    else -> "th"
                                }
                                Text(
                                    "on the $dayOfMonth$ordinalSuffix of the following month(s)",
                                    style = MaterialTheme.typography.bodyMedium
                                )
                                val monthNames = listOf("Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec")
                                FlowRow(
                                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                                    verticalArrangement = Arrangement.spacedBy(8.dp)
                                ) {
                                    monthNames.forEachIndexed { i, name ->
                                        FilterChip(
                                            selected = i in selectedMonths,
                                            onClick = {
                                                selectedMonths = if (i in selectedMonths) {
                                                    if (selectedMonths.size > 1) selectedMonths - i else selectedMonths
                                                } else {
                                                    selectedMonths + i
                                                }
                                            },
                                            label = { Text(name) }
                                        )
                                    }
                                }
                            }
                        }
                    }
                }
            }

            // ── Scheduling Preferences ─────────────────────────────────
            if (isRecurring) {
                Text("Scheduling Preferences", style = MaterialTheme.typography.titleSmall)
                Text(
                    "How should the next occurrence be calculated?",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.Top,
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    RadioButton(
                        selected = !isRolling,
                        onClick = { isRolling = false }
                    )
                    Column {
                        Text("Reschedule from due date", style = MaterialTheme.typography.bodyMedium)
                        Text(
                            "The next task will be scheduled from the original due date, even if the previous task was completed late",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.Top,
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    RadioButton(
                        selected = isRolling,
                        onClick = { isRolling = true }
                    )
                    Column {
                        Text("Reschedule from completion date", style = MaterialTheme.typography.bodyMedium)
                        Text(
                            "The next task will be scheduled from the actual completion date of the previous task",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }

                // ── End Date ───────────────────────────────────────────
                Text("End Date", style = MaterialTheme.typography.titleSmall)
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    Checkbox(
                        checked = hasEndDate,
                        onCheckedChange = { checked ->
                            hasEndDate = checked
                            if (!checked) endDate = null
                            else if (endDate == null) {
                                endDate = (dueDate ?: ZonedDateTime.now(ZoneId.systemDefault()))
                                    .plusMonths(1)
                                    .withSecond(0).withNano(0)
                            }
                        }
                    )
                    Text("Give this task an end date", style = MaterialTheme.typography.bodyMedium)
                }
                if (hasEndDate) {
                    DateTimePickerRow(
                        label = "Select end date & time",
                        value = endDate,
                        onValueSelected = { endDate = it }
                    )
                }
            }

            // ── Notifications ─────────────────────────────────────
            if (hasDueDate) {
                Text("Notifications", style = MaterialTheme.typography.titleSmall)
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    Checkbox(
                        checked = notificationsEnabled,
                        onCheckedChange = { notificationsEnabled = it }
                    )
                    Text("Notify for this task", style = MaterialTheme.typography.bodyMedium)
                }

                if (notificationsEnabled) {
                    Text(
                        "When should notifications trigger?",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        verticalAlignment = Alignment.Top,
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Checkbox(
                            checked = notifyDueDate,
                            onCheckedChange = { notifyDueDate = it }
                        )
                        Column {
                            Text("Due Date/Time", style = MaterialTheme.typography.bodyMedium)
                            Text(
                                "After the due date and time has passed",
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    }
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        verticalAlignment = Alignment.Top,
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Checkbox(
                            checked = notifyPreDue,
                            onCheckedChange = { notifyPreDue = it }
                        )
                        Column {
                            Text("Pre-due", style = MaterialTheme.typography.bodyMedium)
                            Text(
                                "A few hours before the due date",
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    }
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        verticalAlignment = Alignment.Top,
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Checkbox(
                            checked = notifyOverdue,
                            onCheckedChange = { notifyOverdue = it }
                        )
                        Column {
                            Text("Overdue", style = MaterialTheme.typography.bodyMedium)
                            Text(
                                "When left uncompleted at least one day past its due date",
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    }
                }
            }

            Button(
                onClick = {
                    val frequency = when (frequencyType) {
                        FrequencyType.CUSTOM -> when (repeatOn) {
                            RepeatOn.INTERVAL -> Frequency(
                                type = FrequencyType.CUSTOM,
                                on = RepeatOn.INTERVAL,
                                every = intervalEvery.toIntOrNull()?.coerceAtLeast(1) ?: 1,
                                unit = intervalUnit
                            )
                            RepeatOn.DAYS_OF_THE_WEEK -> Frequency(
                                type = FrequencyType.CUSTOM,
                                on = RepeatOn.DAYS_OF_THE_WEEK,
                                days = selectedDays.sorted()
                            )
                            RepeatOn.DAY_OF_THE_MONTHS -> Frequency(
                                type = FrequencyType.CUSTOM,
                                on = RepeatOn.DAY_OF_THE_MONTHS,
                                months = selectedMonths.sorted()
                            )
                            else -> Frequency(type = frequencyType)
                        }
                        else -> Frequency(type = frequencyType)
                    }
                    val notification = if (notificationsEnabled) {
                        NotificationTriggerOptions(
                            enabled = true,
                            dueDate = notifyDueDate,
                            preDue = notifyPreDue,
                            overdue = notifyOverdue
                        )
                    } else {
                        NotificationTriggerOptions()
                    }
                    onSave(
                        title,
                        dueDate?.toIsoString(),
                        endDate?.toIsoString(),
                        frequency,
                        notification,
                        selectedLabelIds.toList(),
                        isRolling
                    )
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
