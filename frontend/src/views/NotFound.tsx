import { Typography } from '@mui/joy'
import React from 'react'

export class NotFound extends React.Component {
  render(): React.ReactNode {
    return (
      <div
        style={{
          display: 'grid',
          minHeight: '100vh',
          placeItems: 'center',
        }}
      >
        <Typography level='h1'>404</Typography>
      </div>
    )
  }
}
