package com.dkhalife.tasks.ui.components

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.expandVertically
import androidx.compose.animation.shrinkVertically
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.Checkbox
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.dkhalife.tasks.R

@Composable
fun NotificationsSection(
    notificationsEnabled: Boolean,
    onEnabledChange: (Boolean) -> Unit,
    notifyDueDate: Boolean,
    onNotifyDueDateChange: (Boolean) -> Unit,
    notifyPreDue: Boolean,
    onNotifyPreDueChange: (Boolean) -> Unit,
    notifyOverdue: Boolean,
    onNotifyOverdueChange: (Boolean) -> Unit
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(stringResource(R.string.notifications_enable_label), style = MaterialTheme.typography.bodyMedium)
        Switch(
            checked = notificationsEnabled,
            onCheckedChange = onEnabledChange
        )
    }

    AnimatedVisibility(
        visible = notificationsEnabled,
        enter = expandVertically(),
        exit = shrinkVertically()
    ) {
        Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
            Text(
                stringResource(R.string.notifications_trigger_description),
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
                    onCheckedChange = onNotifyDueDateChange
                )
                Column {
                    Text(stringResource(R.string.notification_trigger_due_date), style = MaterialTheme.typography.bodyMedium)
                    Text(
                        stringResource(R.string.notification_trigger_due_date_description),
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
                    onCheckedChange = onNotifyPreDueChange
                )
                Column {
                    Text(stringResource(R.string.notification_trigger_pre_due), style = MaterialTheme.typography.bodyMedium)
                    Text(
                        stringResource(R.string.notification_trigger_pre_due_description),
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
                    onCheckedChange = onNotifyOverdueChange
                )
                Column {
                    Text(stringResource(R.string.notification_trigger_overdue), style = MaterialTheme.typography.bodyMedium)
                    Text(
                        stringResource(R.string.notification_trigger_overdue_description),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
        }
    }
}
