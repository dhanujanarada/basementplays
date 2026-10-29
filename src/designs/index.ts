import type { ComponentType } from 'react'

import type { ProductionDesignProps } from '@/lib/types'

import { PlainDesign } from './PlainDesign'

// Register custom production designs here by slug as they are built.
export const designs: Partial<
  Record<string, ComponentType<ProductionDesignProps>>
> = {}

export { PlainDesign }
