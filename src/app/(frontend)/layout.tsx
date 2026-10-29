import type { Metadata } from 'next'
import { Cormorant, Inter, Playfair_Display } from 'next/font/google'
import type { ReactNode } from 'react'

import { SiteFooter } from '@/components/SiteFooter'
import { SiteHeader } from '@/components/SiteHeader'

import './globals.css'

const playfair = Playfair_Display({
  subsets: ['latin'],
  variable: '--font-playfair',
  display: 'swap',
})

const cormorant = Cormorant({
  subsets: ['latin'],
  variable: '--font-cormorant',
  display: 'swap',
  weight: ['400', '500', '600', '700'],
})

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-inter',
  display: 'swap',
})

export const metadata: Metadata = {
  description: 'Independent theatre from Basement Plays.',
  title: {
    default: 'Basement Plays',
    template: '%s | Basement Plays',
  },
}

export default function FrontendLayout({ children }: { children: ReactNode }) {
  return (
    <html
      className={`${playfair.variable} ${cormorant.variable} ${inter.variable}`}
      lang="en"
    >
      <body className="min-h-screen bg-[var(--bg)] text-[var(--text)] antialiased">
        <SiteHeader />
        <main id="main">{children}</main>
        <SiteFooter />
      </body>
    </html>
  )
}
