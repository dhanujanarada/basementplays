export function formatPrice(cents: number, currency = 'LKR'): string {
  const amount = cents / 100
  const hasFraction = !Number.isInteger(amount)
  const formatted = new Intl.NumberFormat('en-LK', {
    maximumFractionDigits: hasFraction ? 2 : 0,
    minimumFractionDigits: hasFraction ? 2 : 0,
  }).format(amount)

  if (currency === 'LKR') {
    return `Rs. ${formatted}`
  }

  return `${currency} ${formatted}`
}

export function formatCountdown(totalSeconds: number): string {
  const safe = Math.max(0, totalSeconds)
  const minutes = Math.floor(safe / 60)
  const seconds = safe % 60

  return `${minutes}:${String(seconds).padStart(2, '0')}`
}

export function secondsUntil(iso: string): number {
  const target = new Date(iso).getTime()
  if (Number.isNaN(target)) {
    return 0
  }

  return Math.max(0, Math.ceil((target - Date.now()) / 1000))
}
