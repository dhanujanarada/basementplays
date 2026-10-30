'use client'

import { usePathname } from 'next/navigation'
import type { ReactNode } from 'react'

type FooterGateProps = {
  children: ReactNode
}

/**
 * Production pages render their own footer inside the design root so the
 * design's CSS variables can theme it. Everywhere else the global footer is
 * shown by the frontend layout.
 */
export function FooterGate({ children }: FooterGateProps) {
  const pathname = usePathname()
  const isProductionPage = /^\/events\/[^/]+$/.test(pathname)

  if (isProductionPage) {
    return null
  }

  return <>{children}</>
}
