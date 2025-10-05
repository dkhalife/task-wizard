import { Frequency, RepeatCustom } from '@/models/task'
import {
  FrequencyType,
  FREQUENCY_TYPES,
  REPEAT_ON_TYPES,
  RepeatOnType,
} from '@/utils/recurrence'
import {
  Box,
  Typography,
  FormControl,
  Checkbox,
  FormHelperText,
  Card,
  List,
  ListItem,
} from '@mui/joy'
import React, { ChangeEvent } from 'react'
import { RepeatOn } from './RepeatOn'

interface RepeatOptionProps {
  nextDueDate: Date | null
  frequency: Frequency
  onFrequencyUpdate: (metadata: Frequency) => void
  frequencyError: string
}

export class RepeatOption extends React.Component<RepeatOptionProps> {
  private onRepeatToggle = (evt: ChangeEvent<HTMLInputElement>): void => {
    if (evt.target.checked) {
      this.props.onFrequencyUpdate(this.defaultFrequencyForType('daily'))
    } else {
      this.props.onFrequencyUpdate(this.defaultFrequencyForType('once'))
    }
  }

  private defaultFrequencyForType = (type: FrequencyType): Frequency => {
    if (type == 'custom') {
      return {
        type: 'custom',
        on: 'interval',
        every: 1,
        unit: 'weeks',
      }
    } else {
      return {
        type,
      }
    }
  }

  private defaultRepeatOnForType = (on: RepeatOnType): RepeatCustom => {
    switch (on) {
      default:
      case 'interval':
        return {
          type: 'custom',
          on: 'interval',
          every: 1,
          unit: 'weeks',
        }
      case 'days_of_the_week':
        return {
          type: 'custom',
          on: 'days_of_the_week',
          days: [0],
        }
      case 'day_of_the_months':
        return {
          type: 'custom',
          on: 'day_of_the_months',
          months: [0],
        }
    }
  }

  private onFrequencyTypeChange = (
    evt: React.ChangeEvent<HTMLInputElement>,
    type: FrequencyType,
  ): void => {
    if (evt.target.checked) {
      this.props.onFrequencyUpdate(this.defaultFrequencyForType(type))
    } else {
      this.props.onFrequencyUpdate(this.defaultFrequencyForType('once'))
    }
  }

  private onRepeatOnTypeUpdate = (
    evt: React.ChangeEvent<HTMLInputElement>,
    type: RepeatOnType,
  ): void => {
    if (evt.target.checked) {
      this.props.onFrequencyUpdate(this.defaultRepeatOnForType(type))
    } else {
      this.props.onFrequencyUpdate(this.defaultRepeatOnForType('interval'))
    }
  }

  render(): React.ReactNode {
    const { frequency, frequencyError, nextDueDate } = this.props
    const isRepeating = frequency.type !== 'once'

    return (
      <Box mt={2}>
        <Typography level='h4'>Repeat :</Typography>
        <FormControl
          sx={{
            mt: 1,
          }}
        >
          <Checkbox
            onChange={this.onRepeatToggle}
            checked={isRepeating}
            overlay
            label='Repeat this task'
          />
        </FormControl>
        {isRepeating && (
          <Card
            sx={{
              mt: 1,
              alignItems: 'center',
            }}
          >
            <Typography>How often should it be repeated?</Typography>

            <List
              orientation='horizontal'
              wrap
              sx={{
                '--List-gap': '8px',
                '--ListItem-radius': '20px',
              }}
            >
              {FREQUENCY_TYPES.map((item: FrequencyType) => {
                if (item === 'once') {
                  return null
                }

                return (
                  <ListItem key={item}>
                    <Checkbox
                      checked={item === frequency.type}
                      onChange={evt => this.onFrequencyTypeChange(evt, item)}
                      overlay
                      disableIcon
                      variant='soft'
                      label={item.charAt(0).toUpperCase() + item.slice(1)}
                    />
                  </ListItem>
                )
              })}
            </List>
            {frequency.type === 'custom' && (
              <>
                <List
                  orientation='horizontal'
                  wrap
                  sx={{
                    '--List-gap': '8px',
                    '--ListItem-radius': '20px',
                  }}
                >
                  {REPEAT_ON_TYPES.map((item: RepeatOnType) => {
                    return (
                      <ListItem key={item}>
                        <Checkbox
                          checked={item === frequency.on}
                          onChange={evt => this.onRepeatOnTypeUpdate(evt, item)}
                          overlay
                          disableIcon
                          variant='soft'
                          label={
                            item.charAt(0).toUpperCase() +
                            item.slice(1).replace(/_/g, ' ')
                          }
                        />
                      </ListItem>
                    )
                  })}
                </List>
                <RepeatOn
                  frequency={frequency}
                  nextDueDate={nextDueDate}
                  onUpdate={this.props.onFrequencyUpdate}
                />
              </>
            )}
            <FormControl error={Boolean(frequencyError)}>
              <FormHelperText>{frequencyError}</FormHelperText>
            </FormControl>
          </Card>
        )}
      </Box>
    )
  }
}
