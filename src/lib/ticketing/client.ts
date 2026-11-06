import { MockTicketingClient } from './mock'
import type { TicketingClient } from './types'

let client: TicketingClient | undefined

// The frontend-first track runs against the in-browser mock. Real wiring is a
// separate step: implement an HTTP client, add the `/ticket-api` proxy route,
// and switch NEXT_PUBLIC_TICKET_API_MODE to `http`.
export function getTicketingClient(): TicketingClient {
  if (process.env.NEXT_PUBLIC_TICKET_API_MODE === 'http') {
    throw new Error(
      'The HTTP ticketing client is not wired yet. Use the mock API mode.',
    )
  }

  client ??= new MockTicketingClient()
  return client
}

export const usingMockTicketing =
  process.env.NEXT_PUBLIC_TICKET_API_MODE !== 'http'
