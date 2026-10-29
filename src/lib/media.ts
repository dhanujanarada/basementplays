import type { Media } from '@/payload-types'

type MediaSize = 'card' | 'full' | 'thumbnail'

export type MediaImage = {
  alt: string
  height: number
  src: string
  width: number
}

export function isMedia(
  value: Media | number | null | undefined,
): value is Media {
  return typeof value === 'object' && value !== null
}

export function mediaImage(
  value: Media | number | null | undefined,
  preferredSize: MediaSize = 'card',
): MediaImage | null {
  if (!isMedia(value)) {
    return null
  }

  const sized = value.sizes?.[preferredSize]
  const src = sized?.url ?? value.url

  if (!src) {
    return null
  }

  return {
    alt: value.alt,
    height: sized?.height ?? value.height ?? 1200,
    src,
    width: sized?.width ?? value.width ?? 1600,
  }
}
