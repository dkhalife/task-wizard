package com.dkhalife.tasks.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.Checkbox
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.RadioButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import java.time.ZonedDateTime

@Composable
fun SchedulingSection(
    isRolling: Boolean,
    onIsRollingChange: (Boolean) -> Unit,
    hasEndDate: Boolean,
    onHasEndDateChange: (Boolean) -> Unit,
    endDate: ZonedDateTime?,
    onEndDateChange: (ZonedDateTime) -> Unit
) {
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
            onClick = { onIsRollingChange(false) }
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
            onClick = { onIsRollingChange(true) }
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

    Text("End Date", style = MaterialTheme.typography.titleSmall)
    Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        Checkbox(
            checked = hasEndDate,
            onCheckedChange = onHasEndDateChange
        )
        Text("Give this task an end date", style = MaterialTheme.typography.bodyMedium)
    }
    if (hasEndDate) {
        DateTimePickerRow(
            label = "Select end date & time",
            value = endDate,
            onValueSelected = onEndDateChange
        )
    }
}
