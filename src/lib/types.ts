import type { ComponentType, ReactNode } from 'react'

import type { EventShow, EventView } from './event-view'

export type EventDesignProps = {
  event: EventView
}

export type EventShellProps = EventDesignProps & {
  children: ReactNode
}

export type EventBookingProps = EventDesignProps & {
  show: EventShow
}

// A design owns its page content, its theme shell, and optionally a custom
// booking layout. The shell is reused by the event, booking, and ticket routes
// so they all inherit the same palette, fonts, and footer.
export type ProductionDesign = {
  Page: ComponentType<EventDesignProps>
  Shell: ComponentType<EventShellProps>
  Booking?: ComponentType<EventBookingProps>
}
