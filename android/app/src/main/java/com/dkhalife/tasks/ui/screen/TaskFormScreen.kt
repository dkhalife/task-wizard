package com.dkhalife.tasks.ui.screen

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.expandVertically
import androidx.compose.animation.shrinkVertically
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
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
import androidx.compose.material.icons.automirrored.filled.Label
import androidx.compose.material.icons.filled.CalendarToday
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.Repeat
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilterChip
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.R
import com.dkhalife.tasks.model.CreateTaskReq
import com.dkhalife.tasks.model.Frequency
import com.dkhalife.tasks.model.FrequencyType
import com.dkhalife.tasks.model.IntervalUnit
import com.dkhalife.tasks.model.Label
import com.dkhalife.tasks.model.NotificationTriggerOptions
import com.dkhalife.tasks.model.RepeatOn
import com.dkhalife.tasks.model.Task
import com.dkhalife.tasks.model.UpdateTaskReq
import com.dkhalife.tasks.ui.components.DateTimePickerRow
import com.dkhalife.tasks.ui.components.NotificationsSection
import com.dkhalife.tasks.ui.components.RecurrenceSection
import com.dkhalife.tasks.ui.components.SchedulingSection
import com.dkhalife.tasks.ui.utils.parseIsoDateTime
import com.dkhalife.tasks.ui.utils.toIsoString
import java.time.ZoneId
import java.time.ZonedDateTime

@Composable
private fun SectionHeader(icon: ImageVector, title: String) {
    Row(verticalAlignment = Alignment.CenterVertically) {
        Icon(
            icon,
            contentDescription = null,
            tint = MaterialTheme.colorScheme.primary,
            modifier = Modifier.size(20.dp)
        )
        Spacer(modifier = Modifier.width(8.dp))
        Text(title, style = MaterialTheme.typography.titleSmall)
    }
}

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
                title = { Text(if (isEditing) stringResource(R.string.title_edit_task) else stringResource(R.string.title_new_task)) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = stringResource(R.string.btn_back))
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.background
                )
            )
        },
        bottomBar = {
            Card(
                shape = RoundedCornerShape(topStart = 16.dp, topEnd = 16.dp),
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.surface
                ),
                elevation = CardDefaults.cardElevation(defaultElevation = 8.dp)
            ) {
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
                            if (hasDueDate) dueDate?.toIsoString() else null,
                            if (hasDueDate && isRecurring && hasEndDate) endDate?.toIsoString() else null,
                            frequency,
                            if (hasDueDate) notification else NotificationTriggerOptions(),
                            selectedLabelIds.toList(),
                            isRolling
                        )
                    },
                    enabled = title.isNotBlank() && !isSaving,
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp)
                        .height(52.dp),
                    shape = RoundedCornerShape(16.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = MaterialTheme.colorScheme.primary
                    )
                ) {
                    if (isSaving) {
                        CircularProgressIndicator(
                            modifier = Modifier.size(20.dp),
                            strokeWidth = 2.dp,
                            color = MaterialTheme.colorScheme.onPrimary
                        )
                        Spacer(modifier = Modifier.width(12.dp))
                    }
                    Text(
                        if (isEditing) stringResource(R.string.btn_update_task) else stringResource(R.string.btn_create_task),
                        style = MaterialTheme.typography.labelLarge
                    )
                }
            }
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

            OutlinedTextField(
                value = title,
                onValueChange = { title = it },
                label = { Text(stringResource(R.string.label_task_title)) },
                singleLine = true,
                shape = RoundedCornerShape(12.dp),
                modifier = Modifier.fillMaxWidth()
            )

            if (availableLabels.isNotEmpty()) {
                Card(
                    shape = RoundedCornerShape(16.dp),
                    colors = CardDefaults.cardColors(
                        containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f)
                    )
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        SectionHeader(Icons.AutoMirrored.Filled.Label, stringResource(R.string.section_labels))
                        Spacer(modifier = Modifier.height(8.dp))
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
                }
            }

            Card(
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f)
                )
            ) {
                Column(modifier = Modifier.padding(16.dp)) {
                    SectionHeader(Icons.Default.CalendarToday, stringResource(R.string.section_due_date))
                    Spacer(modifier = Modifier.height(8.dp))
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.SpaceBetween
                    ) {
                        Text(
                            text = if (isRecurring) stringResource(R.string.due_date_recurring_hint)
                                   else stringResource(R.string.due_date_one_time_hint),
                            style = MaterialTheme.typography.bodyMedium,
                            modifier = Modifier.weight(1f)
                        )
                        Switch(
                            checked = hasDueDate,
                            onCheckedChange = { checked ->
                                hasDueDate = checked
                                if (!checked) {
                                    dueDate = null
                                    frequencyType = FrequencyType.ONCE
                                    hasEndDate = false
                                    endDate = null
                                    notificationsEnabled = false
                                } else if (dueDate == null) {
                                    dueDate = ZonedDateTime.now(ZoneId.systemDefault())
                                        .plusDays(1)
                                        .withHour(9).withMinute(0).withSecond(0).withNano(0)
                                }
                            }
                        )
                    }
                    AnimatedVisibility(
                        visible = hasDueDate,
                        enter = expandVertically(),
                        exit = shrinkVertically()
                    ) {
                        Column {
                            Spacer(modifier = Modifier.height(8.dp))
                            DateTimePickerRow(
                                label = stringResource(R.string.picker_due_date_label),
                                value = dueDate,
                                onValueSelected = { dueDate = it }
                            )
                        }
                    }
                }
            }

            AnimatedVisibility(
                visible = hasDueDate,
                enter = expandVertically(),
                exit = shrinkVertically()
            ) {
                Card(
                    shape = RoundedCornerShape(16.dp),
                    colors = CardDefaults.cardColors(
                        containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f)
                    )
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        SectionHeader(Icons.Default.Repeat, stringResource(R.string.section_repeat))
                        Spacer(modifier = Modifier.height(8.dp))
                        RecurrenceSection(
                            isRecurring = isRecurring,
                            onIsRecurringChange = { checked ->
                                frequencyType = if (checked) FrequencyType.DAILY else FrequencyType.ONCE
                                if (!checked) {
                                    hasEndDate = false
                                    endDate = null
                                }
                            },
                            frequencyType = frequencyType,
                            onFrequencyTypeChange = { frequencyType = it },
                            repeatOn = repeatOn,
                            onRepeatOnChange = { repeatOn = it },
                            intervalEvery = intervalEvery,
                            onIntervalEveryChange = { intervalEvery = it },
                            intervalUnit = intervalUnit,
                            onIntervalUnitChange = { intervalUnit = it },
                            selectedDays = selectedDays,
                            onSelectedDaysChange = { selectedDays = it },
                            selectedMonths = selectedMonths,
                            onSelectedMonthsChange = { selectedMonths = it },
                            dueDate = dueDate
                        )
                    }
                }
            }

            AnimatedVisibility(
                visible = isRecurring,
                enter = expandVertically(),
                exit = shrinkVertically()
            ) {
                Card(
                    shape = RoundedCornerShape(16.dp),
                    colors = CardDefaults.cardColors(
                        containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f)
                    )
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        SchedulingSection(
                            isRolling = isRolling,
                            onIsRollingChange = { isRolling = it },
                            hasEndDate = hasEndDate,
                            onHasEndDateChange = { checked ->
                                hasEndDate = checked
                                if (!checked) endDate = null
                                else if (endDate == null) {
                                    endDate = (dueDate ?: ZonedDateTime.now(ZoneId.systemDefault()))
                                        .plusMonths(1)
                                        .withSecond(0).withNano(0)
                                }
                            },
                            endDate = endDate,
                            onEndDateChange = { endDate = it }
                        )
                    }
                }
            }

            AnimatedVisibility(
                visible = hasDueDate,
                enter = expandVertically(),
                exit = shrinkVertically()
            ) {
                Card(
                    shape = RoundedCornerShape(16.dp),
                    colors = CardDefaults.cardColors(
                        containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f)
                    )
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        SectionHeader(Icons.Default.Notifications, stringResource(R.string.section_notifications))
                        Spacer(modifier = Modifier.height(8.dp))
                        NotificationsSection(
                            notificationsEnabled = notificationsEnabled,
                            onEnabledChange = { notificationsEnabled = it },
                            notifyDueDate = notifyDueDate,
                            onNotifyDueDateChange = { notifyDueDate = it },
                            notifyPreDue = notifyPreDue,
                            onNotifyPreDueChange = { notifyPreDue = it },
                            notifyOverdue = notifyOverdue,
                            onNotifyOverdueChange = { notifyOverdue = it }
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(8.dp))
        }
    }
}