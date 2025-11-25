import {
  Container,
  Typography,
  Divider,
  Box,
  Select,
  Option,
} from '@mui/joy'
import React from 'react'
import { APITokenSettings } from './APITokenSettings'
import { NotificationSettings } from '../Notifications/NotificationSettings'
import { ThemeToggle } from './ThemeToggle'
import { FeatureFlagSettings } from './FeatureFlagSettings'
import { DesktopNotificationToggle } from './DesktopNotificationToggle'
import { storeValue } from '@/utils/storage'
import { getHomeView, HomeView } from '@/utils/navigation'
import { SelectValue } from '@mui/base'
import { isMobile } from '@/utils/dom'

type SettingsProps = object
type SettingsState = {
  homeView: HomeView
}

export class Settings extends React.Component<SettingsProps, SettingsState> {
  constructor(props: SettingsProps) {
    super(props)

    this.state = {
      homeView: getHomeView(),
    }
  }

  private onHomeViewChange = async (
    e: React.MouseEvent | React.KeyboardEvent | React.FocusEvent | null,
    option: SelectValue<HomeView, false>,
  ) => {
    const homeView = option as HomeView
    await this.setState({
      homeView,
    })

    storeValue<HomeView>('home_view', homeView)
  }

  render(): React.ReactNode {
    const { homeView } = this.state

    return (
      <Container
        sx={{
          paddingBottom: '16px',
        }}
      >
        <div
          style={{
            display: 'grid',
            gap: '4',
            paddingTop: '4',
            paddingBottom: '4',
          }}
        >
          {!isMobile() && (
            <>
              <Typography level='h3'>Preferred view</Typography>
              <Divider />
              <Typography>Choose your default view:</Typography>
              <Select
                value={homeView}
                sx={{
                  maxWidth: '200px',
                }}
                onChange={this.onHomeViewChange}
              >
                <Option value='my_tasks'>My Tasks</Option>
                <Option value='tasks_overview'>Tasks Overview</Option>
              </Select>
            </>
          )}
        </div>
        <NotificationSettings />
        <DesktopNotificationToggle />
        <APITokenSettings />
        <FeatureFlagSettings />
        <Box
          sx={{
            mt: 2,
          }}
        >
          <Typography level='h3'>Theme preferences</Typography>
          <Divider />

          <ThemeToggle />
        </Box>
      </Container>
    )
  }
}
