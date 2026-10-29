'use client'

import Image from 'next/image'
import { useState } from 'react'
import Lightbox from 'yet-another-react-lightbox'
import 'yet-another-react-lightbox/styles.css'

import type { MediaImage } from '@/lib/media'

type GalleryProps = {
  images: MediaImage[]
}

export function Gallery({ images }: GalleryProps) {
  const [activeIndex, setActiveIndex] = useState(-1)

  if (images.length === 0) {
    return null
  }

  return (
    <>
      <div className="columns-2 gap-4 md:columns-3">
        {images.map((image, index) => (
          <button
            aria-label={`Open image ${index + 1}: ${image.alt}`}
            className="mb-4 block w-full cursor-zoom-in overflow-hidden rounded-sm border border-[var(--border)] bg-[var(--surface-raised)] focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--accent)]"
            key={`${image.src}-${index}`}
            onClick={() => setActiveIndex(index)}
            type="button"
          >
            <Image
              alt={image.alt}
              className="h-auto w-full transition duration-700 hover:scale-[1.02] motion-reduce:transition-none motion-reduce:hover:scale-100"
              height={image.height}
              sizes="(min-width: 768px) 33vw, 50vw"
              src={image.src}
              width={image.width}
            />
          </button>
        ))}
      </div>

      <Lightbox
        close={() => setActiveIndex(-1)}
        index={activeIndex}
        open={activeIndex >= 0}
        slides={images.map((image) => ({
          alt: image.alt,
          height: image.height,
          src: image.src,
          width: image.width,
        }))}
      />
    </>
  )
}
