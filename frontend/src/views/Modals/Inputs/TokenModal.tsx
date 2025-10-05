import { ApiTokenScope } from '@/models/token'
import { moveFocusToJoyInput } from '@/utils/joy'
import { TokenScopes } from '@/views/Tokens/TokenScopes'
import { SelectValue } from '@mui/base'
import {
  Box,
  Button,
  FormControl,
  Input,
  Modal,
  ModalDialog,
  Option,
  Select,
  Typography,
} from '@mui/joy'
import React, { ChangeEvent } from 'react'

interface TokenModalProps {
  current?: string
  okText?: string
  cancelText?: string

  onClose: (newName: string | null, scopes: ApiTokenScope[], expiration: number) => void
}

interface TokenModalState {
  isOpen: boolean
  name: string
  errorName: boolean
  scopes: ApiTokenScope[]
  expiration: number
}

export class TokenModal extends React.Component<
  TokenModalProps,
  TokenModalState
> {
  private inputRef = React.createRef<HTMLDivElement>()

  constructor(props: TokenModalProps) {
    super(props)
    this.state = {
      isOpen: false,
      name: props.current ?? '',
      errorName: false,
      scopes: [],
      expiration: 7,
    }
  }

  public open = async (): Promise<void> => {
    await this.setState({
      isOpen: true,
    })

    moveFocusToJoyInput(this.inputRef)
  }

  private onSave = () => {
    const { name, scopes, errorName, expiration } = this.state

    if (errorName) {
      return
    }

    this.setState({
      isOpen: false,
      name: '',
      scopes: [],
      expiration: 7,
    })

    this.props.onClose(name, scopes, expiration)
  }

  private onCancel = () => {
    this.setState({
      isOpen: false,
      name: '',
      scopes: [],
    })

    this.props.onClose(null, [], -1)
  }

  private onNameChange = (e: ChangeEvent<HTMLInputElement>) => {
    this.setState({
      name: e.target.value,
      errorName: e.target.value.length === 0,
    })
  }

  private onScopesChange = (scopes: ApiTokenScope[]) => {
    this.setState({
      scopes,
    })
  }

  private onExpirationChange = (e: React.MouseEvent | React.KeyboardEvent | React.FocusEvent | null, option: SelectValue<number, false>) => {
    this.setState({
      expiration: option || 7,
    })
  }

  render(): React.ReactNode {
    const { okText, cancelText } = this.props
    const { name, isOpen, errorName, scopes, expiration } = this.state

    const validState = name.length > 0 && scopes.length > 0

    return (
      <Modal
        open={isOpen}
        onClose={this.onCancel}
      >
        <ModalDialog>
          <Typography
            level='h4'
            mb={1}
          >
            Generate a new token
          </Typography>

          <FormControl error>
            <Typography>Give a name for your new token:</Typography>
            <Input
              placeholder='Name your token'
              value={name}
              onChange={this.onNameChange}
              error={errorName}
              ref={this.inputRef}
              sx={{ minWidth: 300 }}
            />
          </FormControl>

          <FormControl>
            <Typography>Expiration:</Typography>
            <Select
              value={expiration}
              sx={{
                maxWidth: '200px',
              }}
              onChange={this.onExpirationChange}
            >
              <Option value={7}>7 days</Option>
              <Option value={30}>30 days</Option>
              <Option value={60}>60 days</Option>
              <Option value={90}>90 days</Option>
            </Select>
          </FormControl>

          <TokenScopes onChange={this.onScopesChange} />

          <Box
            display={'flex'}
            justifyContent={'space-around'}
            mt={1}
          >
            <Button
              onClick={this.onSave}
              fullWidth
              disabled={!validState}
              sx={{ mr: 1 }}
            >
              {okText ? okText : 'Save'}
            </Button>
            <Button
              onClick={this.onCancel}
              variant='outlined'
            >
              {cancelText ? cancelText : 'Cancel'}
            </Button>
          </Box>
        </ModalDialog>
      </Modal>
    )
  }
}
