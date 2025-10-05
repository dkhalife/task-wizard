import {
  Button,
  Input,
  Typography,
  Divider,
  Box,
  Select,
  Option,
  Snackbar,
} from '@mui/joy'
import React from 'react'
import { SelectValue } from '@mui/base/useSelect/useSelect.types'
import { NotificationOptions } from './NotificationOptions'
import {
  getDefaultTypeForProvider,
  NotificationProvider,
  NotificationTriggerOptions,
  NotificationType,
  NotificationTypeGotify,
  NotificationTypeWebhook,
  WebhookMethod,
} from '@/models/notifications'
import { RootState, AppDispatch } from '@/store/store'
import { connect } from 'react-redux'
import { updateNotificationSettings, setNotificationSettingsDraft } from '@/store/userSlice'

type NotificationSettingProps = {
  draftType: NotificationType
  draftOptions: NotificationTriggerOptions

  setNotificationSettingsDraft: (provider: NotificationType, triggers: NotificationTriggerOptions) => void
  updateNotificationSettings: (type: NotificationType, options: NotificationTriggerOptions) => Promise<any>
}

interface NotificationSettingState {
  saved: boolean
  error: string
}

class NotificationSettingsImpl extends React.Component<
  NotificationSettingProps,
  NotificationSettingState
> {
  constructor(props: NotificationSettingProps) {
    super(props)

    this.state = {
      saved: true,
      error: '',
    }
  }

  private onNotificationProviderChange = (
    e: React.MouseEvent | React.KeyboardEvent | React.FocusEvent | null,
    option: SelectValue<NotificationProvider, false>,
  ) => {
    const provider = option as NotificationProvider
    const type = getDefaultTypeForProvider(provider)

    this.props.setNotificationSettingsDraft(type, this.props.draftOptions)
    this.setState({
      saved: false,
    })
  }

  private onWebhookMethodChanged = (
    e: React.MouseEvent | React.KeyboardEvent | React.FocusEvent | null,
    option: SelectValue<string, false>,
  ) => {
    const method = option as WebhookMethod
    const type = this.props.draftType as NotificationTypeWebhook

    this.props.setNotificationSettingsDraft(
      {
        ...type,
        method,
      },
      this.props.draftOptions
    )
    this.setState({
      saved: false,
    })
  }

  private onURLChanged = (e: React.ChangeEvent<HTMLInputElement>) => {
    const url = e.target.value

    const type = this.props.draftType as
      | NotificationTypeWebhook
      | NotificationTypeGotify

    this.props.setNotificationSettingsDraft(
      {
        ...type,
        url,
      },
      this.props.draftOptions
    )
    this.setState({
      saved: false,
    })
  }

  private onGotifyTokenChanged = (e: React.ChangeEvent<HTMLInputElement>) => {
    const token = e.target.value

    const type = this.props.draftType as NotificationTypeGotify

    this.props.setNotificationSettingsDraft(
      {
        ...type,
        token,
      },
      this.props.draftOptions
    )
    this.setState({
      saved: false,
    })
  }

  private onTriggersChanged = (options: NotificationTriggerOptions) => {
    this.props.setNotificationSettingsDraft(this.props.draftType, options)
    this.setState({
      saved: false,
    })
  }

  private onSaveClicked = async () => {
    const { draftType: type, draftOptions: options } = this.props
    if (type.provider === 'webhook') {
      const isValidURL = type.url.match(/^https?:\/\/[^\s/$.?#].[^\s]*$/i)
      if (!isValidURL) {
        this.setState({
          error: 'Webhook URL must be a valid URL',
        })

        return
      }
    }

    this.setState({
      error: '',
      saved: true,
    })

    try {
      await this.props.updateNotificationSettings(type, options)
    } catch {
      this.setState({
        saved: false,
        error: 'Could not update notification settings',
      })
    }
  }

  private onSnackbarCloseClicked = () => {
    this.setState({ error: '' })
  }

  render(): React.ReactNode {
    const { error, saved } = this.state
    const { draftType: type, draftOptions: options } = this.props

    const placeholderURL =
      type.provider === 'webhook'
        ? 'https://example.com/api/webhook/mmbd7gtpoxp'
        : 'https://mygotifyendpoint.com/'

    return (
      <Box sx={{ mt: 2 }}>
        <Typography level='h3'>Custom Notification</Typography>
        <Divider />

        <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            gap: 2,
            mt: 1,
          }}
        >
          <Box>
            <Typography>Choose a provider: </Typography>
            <Select
              value={type.provider}
              sx={{ maxWidth: '200px' }}
              onChange={this.onNotificationProviderChange}
            >
              <Option value={'none'}>None</Option>
              <Option value={'webhook'}>Webhook</Option>
              <Option value={'gotify'}>Gotify</Option>
            </Select>
          </Box>

          {type.provider === 'webhook' && (
            <Box>
              <Typography>Method: </Typography>
              <Select
                value={type.method}
                sx={{ maxWidth: '200px' }}
                onChange={this.onWebhookMethodChanged}
              >
                <Option value={'GET'}>GET</Option>
                <Option value={'POST'}>POST</Option>
              </Select>
            </Box>
          )}

          {(type.provider === 'webhook' || type.provider === 'gotify') && (
            <Box>
              <Typography>Webhook URL: </Typography>
              <Input
                placeholder={placeholderURL}
                value={type.url}
                onChange={this.onURLChanged}
              />
            </Box>
          )}

          {type.provider === 'gotify' && (
            <Box>
              <Typography>Token: </Typography>
              <Input
                type='password'
                placeholder='Your Gotify token'
                value={type.token}
                onChange={this.onGotifyTokenChanged}
              />
            </Box>
          )}

          {type.provider !== 'none' && (
            <NotificationOptions
              notification={options}
              onChange={this.onTriggersChanged}
            />
          )}

          {!saved && !error && (
            <Box
              sx={{
                display: 'flex',
                justifyContent: 'flex-start',
              }}
            >
              <Button onClick={this.onSaveClicked}>Save</Button>
            </Box>
          )}

          <Snackbar
            open={error !== ''}
            onClose={this.onSnackbarCloseClicked}
            autoHideDuration={3000}
          >
            {error}
          </Snackbar>
        </Box>
      </Box>
    )
  }
}

const mapStateToProps = (state: RootState) => {
  return {
    draftType: state.user.draftNotificationSettings.provider,
    draftOptions: state.user.draftNotificationSettings.triggers,
  }
}

const mapDispatchToProps = (dispatch: AppDispatch) => ({
  setNotificationSettingsDraft: (provider: NotificationType, triggers: NotificationTriggerOptions) =>
    dispatch(setNotificationSettingsDraft({ provider, triggers })),
  updateNotificationSettings: (type: NotificationType, options: NotificationTriggerOptions) =>
    dispatch(updateNotificationSettings({ type, options })),
})

export const NotificationSettings = connect(
  mapStateToProps,
  mapDispatchToProps,
)(NotificationSettingsImpl)
