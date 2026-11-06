import type { ProductionDesign } from '@/lib/types'

import {
  DearDaughterPage,
  DearDaughterShell,
} from './dear-daughter/DearDaughterDesign'
import { PlainPage, PlainShell } from './PlainDesign'

// Register custom production designs here by slug as they are built. A design
// provides its page content, the shared theme shell, and optionally a custom
// booking layout.
export const designs: Partial<Record<string, ProductionDesign>> = {
  'dear-daughter': {
    Page: DearDaughterPage,
    Shell: DearDaughterShell,
  },
}

export const plainDesign: ProductionDesign = {
  Page: PlainPage,
  Shell: PlainShell,
}
