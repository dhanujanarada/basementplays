import type {
  PublicOrder,
  PublicTicket,
  TicketingClient,
} from './types'
import { TicketingError } from './types'

export const usingMockTicketing =
  process.env.NEXT_PUBLIC_TICKET_API_MODE !== 'http'

let clientPromise: Promise<TicketingClient> | undefined

async function createClient(): Promise<TicketingClient> {
  if (usingMockTicketing) {
    const { MockTicketingClient } = await import('./mock')
    return new MockTicketingClient()
  }

  const { HttpTicketingClient } = await import('./http')
  return new HttpTicketingClient()
}

function loadClient(): Promise<TicketingClient> {
  clientPromise ??= createClient()
  return clientPromise
}

function ticketHref(_order: PublicOrder, ticket: PublicTicket): string {
  if (ticket.qr_url) {
    return ticket.qr_url
  }
  if (ticket.qr_token) {
    return `/t/${ticket.qr_token}`
  }

  throw new TicketingError(
    'ticket_not_found',
    'This ticket link is not available yet.',
    404,
  )
}

const client: TicketingClient = {
  getShowSummary: async (showId) =>
    (await loadClient()).getShowSummary(showId),
  getSeatMap: async (showId) => (await loadClient()).getSeatMap(showId),
  createHold: async (showId, request, options) =>
    (await loadClient()).createHold(showId, request, options),
  releaseOrder: async (orderID, orderToken) =>
    (await loadClient()).releaseOrder(orderID, orderToken),
  checkout: async (orderID, orderToken, input, options) =>
    (await loadClient()).checkout(orderID, orderToken, input, options),
  getOrder: async (orderID, orderToken) =>
    (await loadClient()).getOrder(orderID, orderToken),
  getTicket: async (token) => (await loadClient()).getTicket(token),
  devMarkPaid: async (orderID, orderToken) => {
    const active = await loadClient()

    if (!active.devMarkPaid) {
      throw new TicketingError(
        'not_supported',
        'Only the mock client can mark an order paid.',
      )
    }

    return active.devMarkPaid(orderID, orderToken)
  },
  ticketHref,
}

export function getTicketingClient(): TicketingClient {
  return client
}
