import type { ComponentType, ReactNode } from 'react'

import type { Production, Run } from '@/payload-types'

export type Show = NonNullable<Run['shows']>[number]
export type ShowStatus = Show['status']

export type ShowEntry = {
  run: Run
  show: Show
}

export type ProductionDesignProps = {
  production: Production
  runs: Run[]
}

export type ProductionShellProps = ProductionDesignProps & {
  children: ReactNode
}

export type ProductionBookingProps = ProductionDesignProps & {
  run: Run
  show: Show
}

// A design owns its page content, its theme shell, and optionally a custom
// booking layout. The shell is reused by the production, booking, and ticket
// routes so they all inherit the same palette, fonts, and footer.
export type ProductionDesign = {
  Page: ComponentType<ProductionDesignProps>
  Shell: ComponentType<ProductionShellProps>
  Booking?: ComponentType<ProductionBookingProps>
}
