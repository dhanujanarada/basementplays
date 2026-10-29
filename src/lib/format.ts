import type { ShowStatus } from './types'

const siteTimeZone = 'Asia/Colombo'

const dateFormatter = new Intl.DateTimeFormat('en-GB', {
  day: 'numeric',
  month: 'short',
  timeZone: siteTimeZone,
  weekday: 'short',
  year: 'numeric',
})

const longDateFormatter = new Intl.DateTimeFormat('en-GB', {
  day: 'numeric',
  month: 'long',
  timeZone: siteTimeZone,
  weekday: 'long',
  year: 'numeric',
})

const timeFormatter = new Intl.DateTimeFormat('en-GB', {
  hour: 'numeric',
  hour12: true,
  minute: '2-digit',
  timeZone: siteTimeZone,
})

export function formatShowDate(value: string): string {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return 'Date to be announced'
  }

  return `${dateFormatter.format(date)}, ${timeFormatter.format(date)}`
}

export function formatLongShowDate(value: string): string {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return 'Date to be announced'
  }

  return `${longDateFormatter.format(date)} at ${timeFormatter.format(date)}`
}

export function formatShowTime(value: string): string {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return 'Time to be announced'
  }

  return timeFormatter.format(date)
}

export function formatStatus(status: ShowStatus): string {
  switch (status) {
    case 'few-left':
      return 'Few left'
    case 'sold-out':
      return 'Sold out'
    case 'cancelled':
      return 'Cancelled'
    default:
      return 'Open'
  }
}

export function statusTone(
  status: ShowStatus,
): 'accent' | 'default' | 'muted' {
  switch (status) {
    case 'few-left':
      return 'accent'
    case 'sold-out':
    case 'cancelled':
      return 'muted'
    default:
      return 'default'
  }
}
