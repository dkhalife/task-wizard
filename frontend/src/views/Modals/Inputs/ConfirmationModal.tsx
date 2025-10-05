import {
  Box,
  Button,
  ColorPaletteProp,
  Modal,
  ModalDialog,
  Typography,
} from '@mui/joy'
import React from 'react'

type ConfirmedHandler = () => void

export interface ConfirmationModalProps {
  title: string
  message: string
  confirmText: string
  cancelText: string
  color?: ColorPaletteProp
}

type ConfirmationModalClosedState = {
  isOpen: false
  onClose: null
}

type ConfirmationModalOpenedState = {
  isOpen: true
  onClose: ConfirmedHandler
}

type ConfirmationModalState = ConfirmationModalClosedState | ConfirmationModalOpenedState

export class ConfirmationModal extends React.Component<
  ConfirmationModalProps,
  ConfirmationModalState
> {
  private confirmButtonRef: React.RefObject<HTMLButtonElement> = React.createRef()

  constructor(props: ConfirmationModalProps) {
    super(props)

    this.state = {
      isOpen: false,
      onClose: null,
    }
  }

  public async open(onClose: ConfirmedHandler): Promise<void> {
    await this.setState({
      isOpen: true,
      onClose,
    })

    this.confirmButtonRef.current?.focus()
  }

  private onConfirm = () => {
    const { onClose } = this.state as ConfirmationModalOpenedState

    this.setState({
      isOpen: false,
      onClose: null,
    })

    onClose()
  }

  private onCancel = () => {
    this.setState({
      isOpen: false,
      onClose: null,
    })
  }

  public render(): React.ReactNode {
    const { title, message, confirmText, cancelText, color } = this.props
    const { isOpen } = this.state

    return (
      <Modal
        open={isOpen}
        onClose={this.onCancel}
      >
        <ModalDialog>
          <Typography
            level='h4'
          >
            {title}
          </Typography>

          <Typography
            level='body-md'
            gutterBottom
          >
            {message}
          </Typography>

          <Box
            display={'flex'}
            justifyContent={'space-around'}
            mt={1}
          >
            <Button
              onClick={this.onConfirm}
              fullWidth
              sx={{ mr: 1 }}
              color={color ?? 'primary'}
              ref={this.confirmButtonRef}
            >
              {confirmText}
            </Button>
            <Button
              onClick={this.onCancel}
              variant='outlined'
            >
              {cancelText}
            </Button>
          </Box>
        </ModalDialog>
      </Modal>
    )
  }
}
