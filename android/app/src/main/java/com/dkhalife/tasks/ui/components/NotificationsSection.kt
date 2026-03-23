package com.dkhalife.tasks.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.Checkbox
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

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
        horizontalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        Checkbox(
            checked = notificationsEnabled,
            onCheckedChange = onEnabledChange
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
                onCheckedChange = onNotifyDueDateChange
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
                onCheckedChange = onNotifyPreDueChange
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
                onCheckedChange = onNotifyOverdueChange
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
