import type { ComponentType } from 'react'

import type { ProductionDesignProps } from '@/lib/types'

import { DearDaughterDesign } from './dear-daughter/DearDaughterDesign'
import { PlainDesign } from './PlainDesign'

// Register custom production designs here by slug as they are built.
export const designs: Partial<
  Record<string, ComponentType<ProductionDesignProps>>
> = {
  'dear-daughter': DearDaughterDesign,
}

export { PlainDesign }
