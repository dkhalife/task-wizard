package com.dkhalife.tasks.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExposedDropdownMenuAnchorType
import androidx.compose.material3.ExposedDropdownMenuBox
import androidx.compose.material3.ExposedDropdownMenuDefaults
import androidx.compose.material3.FilterChip
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.SegmentedButton
import androidx.compose.material3.SegmentedButtonDefaults
import androidx.compose.material3.SingleChoiceSegmentedButtonRow
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.model.FrequencyType
import com.dkhalife.tasks.model.IntervalUnit
import com.dkhalife.tasks.model.RepeatOn
import com.dkhalife.tasks.ui.utils.getDayOfMonthSuffix
import java.time.ZonedDateTime

@OptIn(ExperimentalMaterial3Api::class, androidx.compose.foundation.layout.ExperimentalLayoutApi::class)
@Composable
fun RecurrenceSection(
    isRecurring: Boolean,
    onIsRecurringChange: (Boolean) -> Unit,
    frequencyType: String,
    onFrequencyTypeChange: (String) -> Unit,
    repeatOn: String,
    onRepeatOnChange: (String) -> Unit,
    intervalEvery: String,
    onIntervalEveryChange: (String) -> Unit,
    intervalUnit: String,
    onIntervalUnitChange: (String) -> Unit,
    selectedDays: Set<Int>,
    onSelectedDaysChange: (Set<Int>) -> Unit,
    selectedMonths: Set<Int>,
    onSelectedMonthsChange: (Set<Int>) -> Unit,
    dueDate: ZonedDateTime?
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text("Repeat this task", style = MaterialTheme.typography.bodyMedium)
        Switch(
            checked = isRecurring,
            onCheckedChange = onIsRecurringChange
        )
    }

    if (isRecurring) {
        SingleChoiceSegmentedButtonRow(modifier = Modifier.fillMaxWidth()) {
            val row1 = listOf(FrequencyType.DAILY, FrequencyType.WEEKLY, FrequencyType.MONTHLY)
            row1.forEachIndexed { index, option ->
                SegmentedButton(
                    selected = frequencyType == option,
                    onClick = { onFrequencyTypeChange(option) },
                    shape = SegmentedButtonDefaults.itemShape(index, row1.size)
                ) {
                    Text(option.replaceFirstChar { it.uppercase() })
                }
            }
        }

        SingleChoiceSegmentedButtonRow(modifier = Modifier.fillMaxWidth()) {
            val row2 = listOf(FrequencyType.YEARLY, FrequencyType.CUSTOM)
            row2.forEachIndexed { index, option ->
                SegmentedButton(
                    selected = frequencyType == option,
                    onClick = { onFrequencyTypeChange(option) },
                    shape = SegmentedButtonDefaults.itemShape(index, row2.size)
                ) {
                    Text(option.replaceFirstChar { it.uppercase() })
                }
            }
        }

        if (frequencyType == FrequencyType.CUSTOM) {
            SingleChoiceSegmentedButtonRow(modifier = Modifier.fillMaxWidth()) {
                val subModes = listOf(RepeatOn.INTERVAL, RepeatOn.DAYS_OF_THE_WEEK, RepeatOn.DAY_OF_THE_MONTHS)
                val subModeLabels = listOf("Interval", "Days of the week", "Day of the months")
                subModes.forEachIndexed { index, mode ->
                    SegmentedButton(
                        selected = repeatOn == mode,
                        onClick = { onRepeatOnChange(mode) },
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
                                    onIntervalEveryChange(v)
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
                                            onIntervalUnitChange(unit)
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
                                    onSelectedDaysChange(
                                        if (i in selectedDays) {
                                            if (selectedDays.size > 1) selectedDays - i else selectedDays
                                        } else {
                                            selectedDays + i
                                        }
                                    )
                                },
                                label = { Text(name) }
                            )
                        }
                    }
                }

                RepeatOn.DAY_OF_THE_MONTHS -> {
                    val dayOfMonth = dueDate?.dayOfMonth ?: 1
                    val ordinalSuffix = getDayOfMonthSuffix(dayOfMonth)
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
                                    onSelectedMonthsChange(
                                        if (i in selectedMonths) {
                                            if (selectedMonths.size > 1) selectedMonths - i else selectedMonths
                                        } else {
                                            selectedMonths + i
                                        }
                                    )
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
