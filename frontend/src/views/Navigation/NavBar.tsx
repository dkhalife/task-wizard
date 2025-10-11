import {
  MenuRounded,
  Logout,
  FormatListBulleted,
  LocalLibrary,
  Label,
  Settings,
} from '@mui/icons-material'
import {
  IconButton,
  Box,
  Typography,
  Drawer,
  List,
  ListItemButton,
  ListItemDecorator,
  ListItemContent,
} from '@mui/joy'
import React from 'react'
import { ThemeToggleButton } from '../Settings/ThemeToggleButton'
import { SyncStatus } from './SyncStatus'
import { NavBarLink } from './NavBarLink'
import { getPathName, NavigationPaths, WithNavigate } from '@/utils/navigation'
import { isMobile } from '@/utils/dom'
import { Logo } from '@/Logo'

type NavBarProps = WithNavigate

interface NavBarState {
  drawerOpen: boolean
}

export class NavBar extends React.Component<NavBarProps, NavBarState> {
  constructor(props: NavBarProps) {
    super(props)
    this.state = {
      drawerOpen: false,
    }
  }

  private frontend_version = import.meta.env.PACKAGE_VERSION
  private backend_version = import.meta.env.VITE_BACKEND_VERSION

  private openDrawer = () => {
    this.setState({ drawerOpen: true })
  }

  private closeDrawer = () => {
    this.setState({ drawerOpen: false })
  }

  private logout = () => {
    localStorage.removeItem('ca_token')
    localStorage.removeItem('ca_expiration')
    this.props.navigate(NavigationPaths.Login)
  }

  render(): React.ReactNode {
    if (['/signup', '/login', '/forgot-password'].includes(getPathName())) {
      return null
    }

    return (
      <nav
        style={{
          display: 'flex',
          gap: 2,
          padding: '3px',
        }}
      >
        <IconButton
          size='sm'
          variant='plain'
          onClick={this.openDrawer}
        >
          <MenuRounded />
        </IconButton>
        <Box
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 2,
          }}
        >
          <Logo
            level='title-lg'
            sx={{
              fontWeight: 700,
              fontSize: 24,
            }}
          />
          <SyncStatus
            style={{
              position: 'absolute',
              right: 50,
            }}
          />
          <ThemeToggleButton
            style={{
              position: 'absolute',
              right: 10,
            }}
          />
        </Box>
        <Drawer
          open={this.state.drawerOpen}
          onClick={this.closeDrawer}
          onClose={this.closeDrawer}
          size='sm'
        >
          <div>
            <List
              size='md'
              onClick={this.openDrawer}
              sx={{ borderRadius: 4, width: '100%', padding: 1 }}
            >
              <NavBarLink
                to={NavigationPaths.MyTasks}
                icon={<LocalLibrary />}
                label='My tasks'
              />
              {!isMobile() && (
                <NavBarLink
                  to={NavigationPaths.TasksOverview}
                  icon={<FormatListBulleted />}
                  label='Tasks Overview'
                />
              )}
              <NavBarLink
                to={NavigationPaths.Labels}
                icon={<Label />}
                label='Labels'
              />
              <NavBarLink
                to={NavigationPaths.Settings}
                icon={<Settings />}
                label='Settings'
              />
            </List>
          </div>
          <div>
            <List
              sx={{
                p: 2,
                height: 'min-content',
                position: 'absolute',
                bottom: 0,
                borderRadius: 4,
                width: '100%',
                padding: 2,
              }}
              size='md'
              onClick={this.openDrawer}
            >
              <ListItemButton
                onClick={this.logout}
                sx={{
                  py: 1.2,
                }}
              >
                <ListItemDecorator>
                  <Logout />
                </ListItemDecorator>
                <ListItemContent>Logout</ListItemContent>
              </ListItemButton>
              <Typography
                level='body-xs'
                sx={{
                  p: 1,
                  color: 'text.tertiary',
                  textAlign: 'center',
                  bottom: 0,
                }}
              >
                Frontend: v{this.frontend_version}<br />
                Backend: v{this.backend_version}
              </Typography>
            </List>
          </div>
        </Drawer>
      </nav>
    )
  }
}
