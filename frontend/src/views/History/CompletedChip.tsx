import { CalendarViewDay, Check, SkipNext, Timelapse } from '@mui/icons-material'
import { Chip, ColorPaletteProp } from '@mui/joy'
import React, { ReactElement } from 'react'

interface CompletedChipProps {
  dueDate: Date | null
  completedDate: Date | null
}

export class CompletedChip extends React.Component<CompletedChipProps> {
  render(): React.ReactNode {
    const { dueDate, completedDate } = this.props

    let text = 'No Due Date'
    let color: ColorPaletteProp = 'neutral'
    let icon: ReactElement = <CalendarViewDay />

    if (completedDate === null) {
      text = 'Skipped'
      color = 'neutral'
      icon = <SkipNext />
    } else if (dueDate) {
      const oneHour = 1 * 60 * 60 * 1000
      const onTime = completedDate.getTime() - dueDate.getTime() <= oneHour

      if (onTime) {
        text = 'On Time'
        color = 'success'
        icon = <Check />
      } else {
        text = 'Late'
        color = 'warning'
        icon = <Timelapse />
      }
    }

    return (
      <Chip
        startDecorator={icon}
        color={color}
      >
        {text}
      </Chip>
    )
  }
}
