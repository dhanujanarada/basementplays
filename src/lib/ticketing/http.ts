import type {
  CheckoutInput,
  Hold,
  HoldRequest,
  PublicOrder,
  PublicTicket,
  ShowSummary,
  SeatMap,
  Ticket,
  TicketingClient,
} from './types'
import { TicketingError } from './types'

const apiBase = (
  process.env.NEXT_PUBLIC_TICKET_API_BASE ?? '/ticket-api'
).replace(/\/$/, '')

type ErrorBody = {
  code?: string
  message?: string
  error?: {
    code?: string
    message?: string
  }
}

function orderHeaders(orderToken: string): HeadersInit {
  return {
    'X-Order-Token': orderToken,
  }
}

async function request<T>(
  path: string,
  init: RequestInit = {},
  fallbackMessage = 'The booking system returned an error.',
): Promise<T> {
  const response = await fetch(`${apiBase}${path}`, {
    ...init,
    headers: {
      Accept: 'application/json',
      ...(init.body ? { 'Content-Type': 'application/json' } : {}),
      ...init.headers,
    },
  })

  if (!response.ok) {
    let body: ErrorBody | null = null

    try {
      body = (await response.json()) as ErrorBody
    } catch {
      // Some upstream errors have no JSON body.
    }

    throw new TicketingError(
      body?.error?.code ?? body?.code ?? 'request_failed',
      body?.error?.message ?? body?.message ?? fallbackMessage,
      response.status,
    )
  }

  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}

export class HttpTicketingClient implements TicketingClient {
  getShowSummary(showId: string): Promise<ShowSummary> {
    return request<ShowSummary>(
      `/v1/public/shows/${encodeURIComponent(showId)}`,
      { cache: 'no-store' },
      'We could not load this performance.',
    )
  }

  getSeatMap(showId: string): Promise<SeatMap> {
    return request<SeatMap>(
      `/v1/public/shows/${encodeURIComponent(showId)}/seats`,
      { cache: 'no-store' },
      'We could not load the seats.',
    )
  }

  createHold(
    showId: string,
    hold: HoldRequest,
    options?: { idempotencyKey?: string },
  ): Promise<Hold> {
    return request<Hold>(
      `/v1/public/shows/${encodeURIComponent(showId)}/holds`,
      {
        body: JSON.stringify(hold),
        headers: options?.idempotencyKey
          ? { 'Idempotency-Key': options.idempotencyKey }
          : undefined,
        method: 'POST',
      },
      'We could not hold those tickets.',
    )
  }

  releaseOrder(orderID: string, orderToken: string): Promise<void> {
    return request<void>(
      `/v1/public/orders/${encodeURIComponent(orderID)}`,
      {
        headers: orderHeaders(orderToken),
        method: 'DELETE',
      },
      'We could not release those seats.',
    )
  }

  checkout(
    orderID: string,
    orderToken: string,
    input: CheckoutInput,
    options?: { idempotencyKey?: string },
  ): Promise<PublicOrder> {
    return request<PublicOrder>(
      `/v1/public/orders/${encodeURIComponent(orderID)}/checkout`,
      {
        body: JSON.stringify(input),
        headers: {
          ...orderHeaders(orderToken),
          ...(options?.idempotencyKey
            ? { 'Idempotency-Key': options.idempotencyKey }
            : {}),
        },
        method: 'POST',
      },
      'We could not confirm this booking.',
    )
  }

  getOrder(orderID: string, orderToken: string): Promise<PublicOrder> {
    return request<PublicOrder>(
      `/v1/public/orders/${encodeURIComponent(orderID)}`,
      {
        cache: 'no-store',
        headers: orderHeaders(orderToken),
      },
      'We could not load this order.',
    )
  }

  getTicket(token: string): Promise<Ticket> {
    return request<Ticket>(
      `/v1/public/tickets/${encodeURIComponent(token)}`,
      { cache: 'no-store' },
      'Ticket not found.',
    )
  }

  ticketHref(_order: PublicOrder, ticket: PublicTicket): string {
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
}
