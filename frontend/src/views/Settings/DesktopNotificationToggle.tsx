import React from 'react'
import { Box, Typography, Divider, FormControl, Checkbox } from '@mui/joy'
import { connect } from 'react-redux'
import { retrieveValue, storeValue } from '@/utils/storage'
import { AppDispatch } from '@/store/store'
import { pushStatus } from '@/store/statusSlice'
import { Status } from '@/models/status'

type DesktopNotificationToggleProps = {
  pushStatus: (status: Status) => void,
}

interface DesktopNotificationToggleState {
  enabled: boolean,
}

class DesktopNotificationToggleImpl extends React.Component<
  DesktopNotificationToggleProps,
  DesktopNotificationToggleState
> {
  constructor(props: DesktopNotificationToggleProps) {
    super(props)

    const enabled = retrieveValue('desktop_notifications', false)
    if (enabled && Notification.permission !== 'granted') {
      storeValue('desktop_notifications', false)
    }

    this.state = {
      enabled: enabled && Notification.permission === 'granted',
    }
  }

  private onToggle = async () => {
    const next = !this.state.enabled
    if (next) {
      if (Notification.permission === 'granted') {
        storeValue('desktop_notifications', true)
        this.setState({
          enabled: true,
        })
        return
      }

      const perm = await Notification.requestPermission()
      if (perm === 'granted') {
        storeValue('desktop_notifications', true)
        this.setState({
          enabled: true,
        })
      } else {
        storeValue('desktop_notifications', false)
        this.setState({
          enabled: false,
        })
        this.props.pushStatus({
          message: 'Notification permissions denied',
          severity: 'error',
          timeout: 3000,
        })
      }
    } else {
      storeValue('desktop_notifications', false)
      this.setState({
        enabled: false,
      })
    }
  }

  render(): React.ReactNode {
    const { enabled } = this.state

    return (
      <Box sx={{ mt: 2 }}>
        <Typography level='h3'>Desktop Notifications</Typography>
        <Divider />
        <FormControl sx={{ mt: 1 }}>
          <Checkbox
            overlay
            checked={enabled}
            onChange={this.onToggle}
            label='Enable desktop notifications'
          />
        </FormControl>
      </Box>
    )
  }
}

const mapDispatchToProps = (dispatch: AppDispatch) => ({
  pushStatus: (status: Status) => dispatch(pushStatus(status)),
})

export const DesktopNotificationToggle = connect(
  null,
  mapDispatchToProps,
)(DesktopNotificationToggleImpl)
