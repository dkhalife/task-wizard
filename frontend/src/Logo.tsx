import { Typography } from '@mui/joy'
import { SxProps, TypographySystem } from '@mui/joy/styles/types'
import React from 'react'

interface LogoProps {
  level?: keyof TypographySystem
  sx?: SxProps
}

export class Logo extends React.Component<LogoProps> {
  render() {
    const { level, sx } = this.props

    return (
      <Typography
        level={level ?? 'h2'}
        sx={sx}
      >
        Task
        <span
          style={{
            color: '#9b59b6',
          }}
        >
          Wizard 🪄
        </span>
      </Typography>
    )
  }
}
