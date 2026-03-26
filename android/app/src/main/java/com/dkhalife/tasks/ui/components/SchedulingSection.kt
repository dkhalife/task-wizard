package com.dkhalife.tasks.ui.components

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.expandVertically
import androidx.compose.animation.shrinkVertically
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.RadioButton
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.R
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
    Text(stringResource(R.string.section_scheduling), style = MaterialTheme.typography.titleSmall)
    Text(
        stringResource(R.string.scheduling_description),
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
            Text(stringResource(R.string.scheduling_from_due_date), style = MaterialTheme.typography.bodyMedium)
            Text(
                stringResource(R.string.scheduling_from_due_date_description),
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
            Text(stringResource(R.string.scheduling_from_completion_date), style = MaterialTheme.typography.bodyMedium)
            Text(
                stringResource(R.string.scheduling_from_completion_date_description),
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }

    Spacer(modifier = Modifier.height(8.dp))
    Text(stringResource(R.string.section_end_date), style = MaterialTheme.typography.titleSmall)
    Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(stringResource(R.string.end_date_hint), style = MaterialTheme.typography.bodyMedium)
        Switch(
            checked = hasEndDate,
            onCheckedChange = onHasEndDateChange
        )
    }
    AnimatedVisibility(
        visible = hasEndDate,
        enter = expandVertically(),
        exit = shrinkVertically()
    ) {
        DateTimePickerRow(
            label = stringResource(R.string.picker_end_date_label),
            value = endDate,
            onValueSelected = onEndDateChange
        )
    }
}
