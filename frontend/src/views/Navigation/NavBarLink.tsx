import {
  ListItem,
  ListItemButton,
  ListItemContent,
  ListItemDecorator,
} from '@mui/joy'
import React from 'react'
import { Link } from 'react-router-dom'

interface NavBarLinkProps {
  to: string
  label: string
  icon: React.ReactNode
}

export class NavBarLink extends React.Component<NavBarLinkProps> {
  render() {
    return (
      <ListItem>
        <ListItemButton
          component={Link}
          to={this.props.to}
          variant='plain'
          color='neutral'
          sx={{
            borderRadius: 4,
            py: 1.2,
          }}
        >
          <ListItemDecorator>{this.props.icon}</ListItemDecorator>
          <ListItemContent>{this.props.label}</ListItemContent>
        </ListItemButton>
      </ListItem>
    )
  }
}
