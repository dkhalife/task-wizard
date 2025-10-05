import {
  DayOfTheWeek,
  Month,
  RepeatCustom,
  RepeatDayOfTheMonths,
  RepeatDaysOfTheWeek,
  RepeatInterval,
  UniqueDaysOfWeek,
  UniqueMonths,
} from '@/models/task'
import { dayOfMonthSuffix } from '@/utils/date'
import { INTERVAL_UNITS, IntervalUnit } from '@/utils/recurrence'
import { SelectValue } from '@mui/base/useSelect/useSelect.types'
import {
  Grid,
  Typography,
  Input,
  Select,
  Card,
  List,
  ListItem,
  Checkbox,
  Box,
  Option,
} from '@mui/joy'
import { format, getDate, getDay, getMonth, setDay, setMonth } from 'date-fns'
import React from 'react'

interface RepeatOnProps {
  nextDueDate: Date | null
  frequency: RepeatCustom
  onUpdate: (options: RepeatCustom) => void
}

type RepeatOnState =
  | {
      every: number
      unit: IntervalUnit
    }
  | {
      days: UniqueDaysOfWeek
    }
  | {
      months: UniqueMonths
    }

export class RepeatOn extends React.Component<RepeatOnProps, RepeatOnState> {
  constructor(props: RepeatOnProps) {
    super(props)

    this.state = this.initWith(props.frequency as any)
  }

  private initWith(frequency: RepeatInterval): RepeatOnState
  private initWith(frequency: RepeatDaysOfTheWeek): RepeatOnState
  private initWith(frequency: RepeatDayOfTheMonths): RepeatOnState
  private initWith(frequency: any): RepeatOnState {
    if (frequency.type != 'custom') {
      throw new Error('Invalid frequency')
    }

    switch (frequency.on) {
      case 'interval':
        return {
          every: frequency.interval,
          unit: frequency.unit,
        }

      case 'days_of_the_week':
        return {
          days: frequency.days,
        }

      case 'day_of_the_months':
        return {
          months: frequency.months,
        }

      default:
        throw new Error('Invalid frequency')
    }
  }

  public componentDidUpdate(prevProps: Readonly<RepeatOnProps>): void {
    if (prevProps.frequency.on != this.props.frequency.on) {
      this.setState(this.initWith(this.props.frequency as any))
    }
  }

  private onIntervalChange = (evt: React.ChangeEvent<HTMLInputElement>) => {
    if (
      this.props.frequency.type != 'custom' &&
      this.props.frequency.on != 'interval'
    ) {
      throw new Error('Invalid frequency')
    }

    const currentState = this.state as RepeatInterval
    const newState = {
      every: parseInt(evt.target.value) || 1,
    }

    this.setState(newState)
    this.props.onUpdate({
      type: 'custom',
      on: 'interval',
      unit: currentState.unit,
      ...newState,
    })
  }

  private onIntervalUnitChange = (
    evt: React.MouseEvent | React.KeyboardEvent | React.FocusEvent | null,
    value: SelectValue<IntervalUnit, false>,
  ) => {
    if (
      this.props.frequency.type != 'custom' &&
      this.props.frequency.on != 'interval'
    ) {
      throw new Error('Invalid frequency')
    }

    const currentState = this.state as RepeatInterval
    const newState = {
      unit: value as IntervalUnit,
    }

    this.setState(newState)
    this.props.onUpdate({
      type: 'custom',
      on: 'interval',
      every: currentState.every,
      ...newState,
    })
  }

  private onDayOfTheWeekChange = (evt: React.ChangeEvent<HTMLInputElement>) => {
    const { frequency, nextDueDate } = this.props
    if (frequency.type != 'custom' && frequency.on != 'days_of_the_week') {
      throw new Error('Invalid frequency')
    }

    const currentState = this.state as RepeatDaysOfTheWeek
    const days = currentState.days

    const day = parseInt(evt.target.value, 10) as DayOfTheWeek
    const newDays = days.includes(day)
      ? days.filter(d => d !== day)
      : [...days, day].sort()
    if (newDays.length == 0) {
      const defaultDay = getDay(nextDueDate ?? new Date()) as DayOfTheWeek
      newDays.push(defaultDay)
    }

    const newState = {
      days: newDays as UniqueDaysOfWeek,
    }

    this.setState(newState)
    this.props.onUpdate({
      type: 'custom',
      on: 'days_of_the_week',
      ...newState,
    })
  }

  private onMonthChange = (evt: React.ChangeEvent<HTMLInputElement>) => {
    const { frequency, nextDueDate } = this.props
    if (frequency.type != 'custom' && frequency.on != 'day_of_the_months') {
      throw new Error('Invalid frequency')
    }

    const currentState = this.state as RepeatDayOfTheMonths
    const months = currentState.months

    const month = parseInt(evt.target.value, 10) as Month
    const newMonths = months.includes(month)
      ? months.filter(m => m !== month)
      : [...months, month].sort()
    if (newMonths.length == 0) {
      const defaultMonth = getMonth(nextDueDate ?? new Date()) as Month
      newMonths.push(defaultMonth)
    }

    const newState = {
      months: newMonths as UniqueMonths,
    }

    this.setState(newState)
    this.props.onUpdate({
      type: 'custom',
      on: 'day_of_the_months',
      ...newState,
    })
  }

  render(): React.ReactNode {
    const { frequency, nextDueDate } = this.props
    const day = nextDueDate ? getDate(nextDueDate) : 1

    switch (frequency.on) {
      case 'interval':
        return (
          <>
            <Grid
              sm={12}
              sx={{ display: 'flex', alignItems: 'center' }}
            >
              <Typography>Every :&nbsp;</Typography>
              <Input
                slotProps={{
                  input: {
                    min: 1,
                    max: 365,
                  },
                }}
                type='number'
                sx={{
                  maxWidth: '75px',
                  mr: 1,
                }}
                value={frequency.every}
                onChange={this.onIntervalChange}
              />
              <Select
                placeholder='Unit'
                value={frequency.unit}
                onChange={this.onIntervalUnitChange}
              >
                {INTERVAL_UNITS.map(item => (
                  <Option
                    key={item}
                    value={item}
                  >
                    {item.charAt(0).toUpperCase() + item.slice(1)}
                  </Option>
                ))}
              </Select>
            </Grid>
          </>
        )
      case 'days_of_the_week':
        return (
          <>
            <Grid
              sm={12}
              sx={{ display: 'flex', alignItems: 'center' }}
            >
              <Card>
                <List
                  orientation='horizontal'
                  wrap
                  sx={{
                    '--List-gap': '8px',
                    '--ListItem-radius': '20px',
                  }}
                >
                  {[...Array(7).keys()].map(item => (
                    <ListItem key={item}>
                      <Checkbox
                        checked={
                          frequency.days.includes(item as DayOfTheWeek) || false
                        }
                        value={item as DayOfTheWeek}
                        onChange={this.onDayOfTheWeekChange}
                        overlay
                        disableIcon
                        variant='soft'
                        label={format(setDay(new Date(), item), 'EEE')}
                      />
                    </ListItem>
                  ))}
                </List>
              </Card>
            </Grid>
          </>
        )
      case 'day_of_the_months':
        return (
          <Grid
            sm={12}
            sx={{
              display: 'flex',
              alignItems: 'center',
              flexDirection: 'column',
              justifyContent: 'space-between',
            }}
          >
            <Box
              sx={{
                display: 'flex',
                textAlign: 'left',
                mb: 1.5,
              }}
            >
              <Typography>
                on the {day}
                {dayOfMonthSuffix(day)} of the following month(s)
              </Typography>
            </Box>
            <Card>
              <List
                orientation='horizontal'
                wrap
                sx={{
                  '--List-gap': '8px',
                  '--ListItem-radius': '20px',
                }}
              >
                {[...Array(12).keys()].map(item => (
                  <ListItem key={item}>
                    <Checkbox
                      checked={frequency.months.includes(item as Month)}
                      value={item as Month}
                      onChange={this.onMonthChange}
                      overlay
                      disableIcon
                      variant='soft'
                      label={format(setMonth(new Date(), item), 'MMM')}
                    />
                  </ListItem>
                ))}
              </List>
            </Card>
          </Grid>
        )

      default:
        return null
    }
  }
}
