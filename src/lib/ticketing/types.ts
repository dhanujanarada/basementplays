// Public ticketing API types. Field names are snake_case because they mirror
// the Go service's JSON contract exactly; keep them in sync when the real HTTP
// client is wired up.

export type SeatStatus = 'available' | 'held' | 'sold' | 'blocked'
export type PaymentMode = 'bank_transfer' | 'online' | 'off'
export type SelectionMode = 'quantity' | 'seats'
export type StagePosition = 'front' | 'back' | 'none'
type PaymentMethod = 'payhere' | 'bank_transfer' | 'cash' | 'whatsapp'

export const MAX_SEATS_PER_ORDER = 6

export type TierSummary = {
  tier_id: string
  name: string
  price_cents: number
  seats_total: number
  seats_remaining: number
}

export type BankTransferDetails = {
  bank_name: string
  account_name: string
  account_number: string
  branch: string
  instructions: string
}

export type ShowSummary = {
  id: string
  production_slug: string
  title: string
  starts_at: string
  venue: string
  status: string
  payment_mode: PaymentMode
  selection_mode: SelectionMode
  stage_position: StagePosition
  numbering: boolean
  currency: string
  columns: number
  stage_label: string
  total_seats: number
  seats_remaining: number
  price_from: number
  tiers: TierSummary[]
}

export type Seat = {
  code: string
  label: string | null
  tier_id: string
  tier_name: string
  price_cents: number
  status: SeatStatus
  sort_index: number
}

export type SeatMap = {
  show_id: string
  status: string
  payment_mode: PaymentMode
  selection_mode: SelectionMode
  stage_position: StagePosition
  numbering: boolean
  currency: string
  columns: number
  stage_label: string
  total_seats: number
  seats_remaining: number
  tiers: TierSummary[]
  seats: Seat[]
}

export type OrderSeat = {
  code: string
  label: string | null
  tier_id: string
  tier_name: string
  price_cents: number
}

export type PublicTicket = {
  ticket_id: string
  seat_code: string
  status: 'issued' | 'used' | 'void'
  issued_at: string
  used_at: string | null
  qr_token?: string | null
  qr_url?: string | null
}

export type PublicOrder = {
  order_id: string
  show_id: string
  status: 'pending' | 'paid' | 'expired' | 'cancelled' | 'needs_review'
  source: 'online' | 'manual'
  payment_method: PaymentMethod
  total_cents: number
  currency: string
  expires_at: string
  paid_at: string | null
  created_at: string
  bank_transfer: BankTransferDetails | null
  seats: OrderSeat[]
  tickets: PublicTicket[]
}

export type Hold = {
  order_id: string
  show_id: string
  status: string
  source: string
  total_cents: number
  currency: string
  expires_at: string
  seats: OrderSeat[]
  order_token?: string
}

export type HoldRequest =
  | {
      tiers: { tier_id: string; quantity: number }[]
    }
  | {
      seats: string[]
    }

export type CheckoutInput = {
  buyer_name: string
  buyer_phone: string
  buyer_email: string
  payment_method: 'payhere' | 'bank_transfer'
}

export type Ticket = {
  ticket_id: string
  status: 'issued' | 'used' | 'void'
  production_slug: string
  show_id: string
  show_title: string
  starts_at: string
  venue: string
  seat_code: string
  seat_label: string | null
  tier_id: string
  tier_name: string
  description: string
  price_cents: number
  order_reference: string
  qr_token: string
  qr_url: string
}

export type TicketingClient = {
  getShowSummary(showId: string): Promise<ShowSummary>
  getSeatMap(showId: string): Promise<SeatMap>
  createHold(
    showId: string,
    request: HoldRequest,
    options?: { idempotencyKey?: string },
  ): Promise<Hold>
  releaseOrder(orderID: string, orderToken: string): Promise<void>
  checkout(
    orderID: string,
    orderToken: string,
    input: CheckoutInput,
    options?: { idempotencyKey?: string },
  ): Promise<PublicOrder>
  getOrder(orderID: string, orderToken: string): Promise<PublicOrder>
  getTicket(token: string): Promise<Ticket>
  ticketHref(order: PublicOrder, ticket: PublicTicket): string
  /** Dev-only mock helper; the real service marks paid from Backstage. */
  devMarkPaid?(orderID: string, orderToken: string): Promise<PublicOrder>
}

export class TicketingError extends Error {
  code: string
  status: number

  constructor(code: string, message: string, status = 400) {
    super(message)
    this.name = 'TicketingError'
    this.code = code
    this.status = status
  }
}
