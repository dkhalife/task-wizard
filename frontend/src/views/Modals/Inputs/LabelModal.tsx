import { ColorOption, LABEL_COLORS } from '@/utils/colors'
import { ModalDialog } from '@mui/joy'
import {
  Modal,
  Typography,
  FormControl,
  Input,
  Select,
  Box,
  Button,
  Option,
} from '@mui/joy'
import React, { ChangeEvent } from 'react'
import { SelectValue } from '@mui/base/useSelect/useSelect.types'
import { moveFocusToJoyInput } from '@/utils/joy'
import { connect } from 'react-redux'
import { Label } from '@/models/label'
import { createLabel, exitEditMode, updateLabel } from '@/store/labelsSlice'
import { AppDispatch, RootState } from '@/store/store'

interface LabelModalProps {
  id: number | undefined
  name: string | undefined
  color: ColorOption | undefined
  isOpen: boolean
  isUnique: (id: number | undefined, labelName: string) => boolean

  createLabel: (label: Omit<Label, 'id'>) => Promise<any>
  updateLabel: (label: Label) => Promise<any>
  exitEditMode: () => void
}

interface LabelModalState {
  labelName: string
  color: ColorOption | undefined
  error: string
}

class LabelModalImpl extends React.Component<
  LabelModalProps,
  LabelModalState
> {
  private inputRef = React.createRef<HTMLInputElement>()

  constructor(props: LabelModalProps) {
    super(props)
    this.state = {
      labelName: props.name ?? '',
      color: props.color,
      error: '',
    }
  }

  componentDidUpdate(prevProps: Readonly<LabelModalProps>): void {
    if (
      this.props.name !== prevProps.name ||
      this.props.color !== prevProps.color ||
      this.props.id !== prevProps.id
    ) {
      this.setState({
        labelName: this.props.name ?? '',
        color: this.props.color,
        error: '',
      })
    }

    if (this.props.isOpen && !prevProps.isOpen && this.inputRef.current) {
      moveFocusToJoyInput(this.inputRef)
    }
  }

  private validateLabel = () => {
    const { labelName, color } = this.state

    if (!labelName.trim()) {
      this.setState({ error: 'Name cannot be empty' })
      return false
    }

    const { isUnique } = this.props
    if (!isUnique(this.props.id, labelName)) {
      this.setState({ error: 'Label with this name already exists' })
      return false
    }

    if (!color) {
      this.setState({ error: 'Please select a color' })
      return false
    }

    return true
  }

  private onSave = async () => {
    const { labelName, color } = this.state

    if (!this.validateLabel() || !color) {
      return
    }

    const { id } = this.props

    if (id) {
      try {
        await this.props.updateLabel({
          id: id,
          name: labelName,
          color: color.value,
        })
      } catch {
        this.setState({ error: 'Failed to save label. Please try again.' })
      }
    } else {
      try {
          await this.props.createLabel({
            name: labelName,
            color: color.value,
          })
      } catch {
        this.setState({ error: 'Failed to create label. Please try again.' })
      }
    }

    this.props.exitEditMode()
  }

  private onCancel = () => {
    this.props.exitEditMode()
  }

  private onLabelNameChange = (e: ChangeEvent<HTMLInputElement>) => {
    this.setState({ labelName: e.target.value })
  }

  private onColorChange = (
    e: React.MouseEvent | React.KeyboardEvent | React.FocusEvent | null,
    value: SelectValue<ColorOption, false>,
  ) => {
    if (value) {
      this.setState({ color: value })
    }
  }

  public render(): React.ReactNode {
    const { id, isOpen } = this.props
    const { labelName, color, error } = this.state

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
            {id ? 'Edit Label' : 'Add Label'}
          </Typography>

          <FormControl>
            <Typography
              gutterBottom
              level='body-sm'
              alignSelf='start'
            >
              Name
            </Typography>
            <Input
              fullWidth
              value={labelName}
              ref={this.inputRef}
              onChange={this.onLabelNameChange}
            />
          </FormControl>

          <FormControl>
            <Typography
              gutterBottom
              level='body-sm'
              alignSelf='start'
            >
              Color
            </Typography>
            <Select
              onChange={this.onColorChange}
              required={true}
              value={color}
              startDecorator={
                <span
                  style={{
                    borderRadius: 10,
                    display: 'inline-block',
                    background: color?.value,
                    width: '20px',
                    height: '20px',
                  }}
                />
              }
            >
              {LABEL_COLORS.map((opt: ColorOption) => (
                <Option
                  key={opt.value}
                  value={opt}
                >
                  <Box
                    style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                    }}
                  >
                    <Box
                      width={20}
                      height={20}
                      borderRadius={10}
                      sx={{ background: opt.value }}
                    />
                    <Typography sx={{ ml: 1 }}>{opt.name}</Typography>
                  </Box>
                </Option>
              ))}
            </Select>
          </FormControl>

          {error && (
            <Typography
              color='warning'
              level='body-sm'
            >
              {error}
            </Typography>
          )}

          <Box
            display='flex'
            justifyContent='space-around'
            mt={1}
          >
            <Button
              onClick={this.onSave}
              fullWidth
              sx={{ mr: 1 }}
            >
              {id ? 'Save Changes' : 'Add Label'}
            </Button>
            <Button
              onClick={this.onCancel}
              variant='outlined'
            >
              Cancel
            </Button>
          </Box>
        </ModalDialog>
      </Modal>
    )
  }
}

const mapStateToProps = (state: RootState) => ({
  isOpen: state.labels.isEditing,
})

const mapDispatchToProps = (dispatch: AppDispatch) => ({
  createLabel: (label: Omit<Label, 'id'>) => dispatch(createLabel(label)),
  updateLabel: (label: Label) => dispatch(updateLabel(label)),
  exitEditMode: () => dispatch(exitEditMode()),
})

export const LabelModal = connect(
  mapStateToProps,
  mapDispatchToProps,
)(LabelModalImpl)
