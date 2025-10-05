import React from 'react'
import {
  Box,
  Card,
  Checkbox,
  FormControl,
  FormHelperText,
  Typography,
} from '@mui/joy'
import {
  NotificationTriggerOption,
  NotificationTriggerOptions,
} from '@/models/notifications'

type NotificationOptionsProps = {
  notification: NotificationTriggerOptions
  onChange: (notification: NotificationTriggerOptions) => void
}

type NotificationOptionsState = {
  notification: NotificationTriggerOptions
}

export class NotificationOptions extends React.Component<
  NotificationOptionsProps,
  NotificationOptionsState
> {
  constructor(props: NotificationOptionsProps) {
    super(props)

    this.state = {
      notification: props.notification,
    }
  }

  componentDidUpdate(prevProps: NotificationOptionsProps) {
    if (prevProps.notification !== this.props.notification) {
      this.setState({
        notification: this.props.notification,
      })
    }
  }

  private onNotificationOptionChange = (item: NotificationTriggerOption) => {
    const { notification } = this.state

    const newNotification = {
      ...notification,
      [item.type]: !notification[item.type],
    }

    this.setState({
      notification: newNotification,
    })

    this.props.onChange(newNotification)
  }

  render() {
    const { notification } = this.state

    return (
      <Box
        mt={1}
        sx={{
          display: 'flex',
          flexDirection: 'column',
          gap: 2,

          '& > div': { p: 2, borderRadius: 'md', display: 'flex' },
        }}
      >
        <Card variant='outlined'>
          <Typography>When should notifications trigger?</Typography>
          {NotificationTriggerOptions.map(item => (
            <FormControl
              sx={{ mb: 1 }}
              key={item.type}
            >
              <Checkbox
                overlay
                onClick={() => this.onNotificationOptionChange(item)}
                checked={notification[item.type]}
                label={item.title}
                key={item.title}
              />
              <FormHelperText>{item.description}</FormHelperText>
            </FormControl>
          ))}
        </Card>
      </Box>
    )
  }
}
