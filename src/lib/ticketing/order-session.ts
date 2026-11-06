'use client'

const PREFIX = 'bp.ticketing.order-token.'

export function saveOrderToken(orderID: string, token: string): void {
  try {
    window.sessionStorage.setItem(`${PREFIX}${orderID}`, token)
  } catch {
    // Private browsing can disable storage. The confirmation page will explain
    // that the order cannot be recovered in this tab.
  }
}

export function readOrderToken(orderID: string): string | null {
  try {
    return window.sessionStorage.getItem(`${PREFIX}${orderID}`)
  } catch {
    return null
  }
}

export function clearOrderToken(orderID: string): void {
  try {
    window.sessionStorage.removeItem(`${PREFIX}${orderID}`)
  } catch {
    // Nothing to clean up.
  }
}
