export const dayOfMonthSuffix = (n: number): string => {
  if (n >= 11 && n <= 13) {
    return 'th'
  }

  switch (n % 10) {
    case 1:
      return 'st'
    case 2:
      return 'nd'
    case 3:
      return 'rd'
    default:
      return 'th'
  }
}
