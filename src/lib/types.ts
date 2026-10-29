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
