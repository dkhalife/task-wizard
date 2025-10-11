import { ListItemContent, ListDivider } from '@mui/joy'
import { ListItem, Box, Typography } from '@mui/joy'
import React from 'react'
import { CompletedChip } from './CompletedChip'
import { format, formatDistanceToNow } from 'date-fns'
import { HistoryEntryUI } from '@/utils/marshalling'

interface HistoryCardProps {
  historyEntry: HistoryEntryUI
}

export class HistoryCard extends React.Component<HistoryCardProps> {
  render(): React.ReactNode {
    const { historyEntry } = this.props

    return (
      <>
        <ListItem
          sx={{
            gap: 1.5,
            alignItems: 'flex-start',
          }}
        >
          <ListItemContent sx={{ my: 0 }}>
            <Box
              sx={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
              }}
            >
              <Typography sx={{ fontWeight: '600' }}>
                {historyEntry.completed_date ? 'Completed' : 'Skipped'}
              </Typography>
              <CompletedChip
                dueDate={historyEntry.due_date}
                completedDate={historyEntry.completed_date}
              />
            </Box>
            {historyEntry.completed_date && (
              <Typography>
                {`on ${format(historyEntry.completed_date, 'MMMM do yyyy, h:mm a')}`}
              </Typography>
            )}
          </ListItemContent>
        </ListItem>
        <ListDivider component='li'>
          <Typography>
            {historyEntry.due_date
              ? `due ${formatDistanceToNow(historyEntry.due_date)}`
              : '-'}
          </Typography>
        </ListDivider>
      </>
    )
  }
}
