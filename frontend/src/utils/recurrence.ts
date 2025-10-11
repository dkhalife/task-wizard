export type IntervalUnit = 'hours' | 'days' | 'weeks' | 'months' | 'years'

export const INTERVAL_UNITS: IntervalUnit[] = [
  'hours',
  'days',
  'weeks',
  'months',
  'years',
]

export type FrequencyType =
  | 'once'
  | 'daily'
  | 'weekly'
  | 'monthly'
  | 'yearly'
  | 'custom'

export const FREQUENCY_TYPES: FrequencyType[] = [
  'once',
  'daily',
  'weekly',
  'monthly',
  'yearly',
  'custom',
]

export type RepeatOnType = 'interval' | 'days_of_the_week' | 'day_of_the_months'

export const REPEAT_ON_TYPES: RepeatOnType[] = [
  'interval',
  'days_of_the_week',
  'day_of_the_months',
]
