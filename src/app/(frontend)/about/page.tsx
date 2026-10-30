import { RichText } from '@payloadcms/richtext-lexical/react'
import Image from 'next/image'

import { Container } from '@/components/ui/Container'
import { mediaImage } from '@/lib/media'
import { getSiteSettings } from '@/lib/payload'

export const dynamic = 'force-dynamic'

export const metadata = {
  title: 'About',
}

export default async function AboutPage() {
  const settings = await getSiteSettings()
  const aboutImage = mediaImage(settings.aboutImage, 'full')
  const groupName = settings.groupName || 'Basement Plays'

  return (
    <Container className="py-16 sm:py-24">
      <p className="text-xs font-medium uppercase tracking-[0.22em] text-[var(--accent)]">
        About
      </p>
      <h1 className="font-display mt-4 text-5xl tracking-tight text-[var(--text)] sm:text-7xl">
        {groupName}
      </h1>
      {settings.tagline ? (
        <p className="mt-5 max-w-2xl text-base leading-7 text-[var(--muted)] sm:text-lg">
          {settings.tagline}
        </p>
      ) : null}

      <div className="mt-12 grid gap-10 lg:grid-cols-[0.8fr_1.2fr] lg:items-start">
        {aboutImage ? (
          <div className="relative overflow-hidden rounded-sm border border-[var(--border)] bg-[var(--surface-raised)]">
            <Image
              alt={aboutImage.alt}
              className="h-auto w-full object-cover"
              height={aboutImage.height}
              sizes="(min-width: 1024px) 40vw, 100vw"
              src={aboutImage.src}
              width={aboutImage.width}
            />
          </div>
        ) : null}

        <div className={aboutImage ? '' : 'lg:col-span-2'}>
          {settings.aboutText ? (
            <RichText
              className="max-w-2xl text-base leading-8 text-[var(--muted)] [&_a]:text-[var(--accent)] [&_a]:underline [&_blockquote]:border-l-2 [&_blockquote]:border-[var(--accent)] [&_blockquote]:pl-5 [&_h2]:font-display [&_h2]:text-3xl [&_h2]:text-[var(--text)] [&_h3]:font-display [&_h3]:text-2xl [&_h3]:text-[var(--text)] [&_li]:mb-2 [&_ol]:mb-5 [&_ol]:list-decimal [&_ol]:pl-5 [&_p]:mb-5 [&_p:last-child]:mb-0 [&_ul]:mb-5 [&_ul]:list-disc [&_ul]:pl-5"
              data={settings.aboutText}
            />
          ) : (
            <p className="max-w-2xl text-base leading-8 text-[var(--muted)]">
              The about story will appear here once it is added in Backstage.
            </p>
          )}
        </div>
      </div>
    </Container>
  )
}
