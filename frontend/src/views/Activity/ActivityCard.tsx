import { ListItemContent, ListDivider, Box, Typography, Button } from '@mui/joy'
import { Undo } from '@mui/icons-material'
import { ListItem } from '@mui/joy'
import React from 'react'
import { CompletedChip } from '@/views/History/CompletedChip'
import { format, formatDistanceToNow } from 'date-fns'
import { ActivityEntryUI } from '@/utils/marshalling'

interface ActivityCardProps {
  entry: ActivityEntryUI
  onRevert: (entry: ActivityEntryUI) => void
  onTitleClick: (taskId: number) => void
}

export class ActivityCard extends React.Component<ActivityCardProps> {
  render(): React.ReactNode {
    const { entry, onRevert, onTitleClick } = this.props

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
                gap: 1,
              }}
            >
              <Typography
                sx={{ fontWeight: '600', cursor: 'pointer' }}
                onClick={() => onTitleClick(entry.task_id)}
              >
                {entry.task_title || '--'}
              </Typography>
              <CompletedChip
                dueDate={entry.due_date}
                completedDate={entry.completed_date}
              />
            </Box>
            <Box
              sx={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                gap: 1,
                mt: 0.5,
              }}
            >
              <Typography level='body-sm'>
                {entry.completed_date
                  ? `Completed on ${format(entry.completed_date, 'MMMM do yyyy, h:mm a')}`
                  : 'Skipped'}
              </Typography>
              {entry.is_latest && (
                <Button
                  size='sm'
                  variant='outlined'
                  color='warning'
                  startDecorator={<Undo />}
                  onClick={() => onRevert(entry)}
                >
                  Revert
                </Button>
              )}
            </Box>
          </ListItemContent>
        </ListItem>
        <ListDivider component='li'>
          <Typography>
            {entry.due_date
              ? `due ${formatDistanceToNow(entry.due_date, { addSuffix: true })}`
              : '-'}
          </Typography>
        </ListDivider>
      </>
    )
  }
}
