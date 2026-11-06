import 'dotenv/config'

import fs from 'node:fs'

import type { Payload } from 'payload'
import { getPayload } from 'payload'

import config from '../src/payload.config'
import type { Production } from '../src/payload-types'
import {
  renderPlaceholder,
  type PlaceholderImage,
} from './seed-images'

type Synopsis = NonNullable<Production['synopsis']>

const databaseUrl = process.env.DATABASE_URL || 'file:./payload.db'

if (databaseUrl.startsWith('file:')) {
  const databasePath = databaseUrl.slice('file:'.length)

  if (!fs.existsSync(databasePath)) {
    throw new Error(
      'payload.db does not exist yet. Run `npm run dev` once to create the schema, then run `npm run seed:demo`.',
    )
  }
}

const poster: PlaceholderImage = {
  accent: '#b5453c',
  alt: 'The Hollow Crown poster with a crown motif and dark red lettering',
  background: '#08070b',
  eyebrow: 'A Basement Plays production',
  foreground: '#e9dfd2',
  key: 'hollow-crown-poster',
  kind: 'poster',
  motif: 'crown',
  subtitle: 'A gothic tragedy of inheritance and rot',
  title: 'THE HOLLOW|CROWN',
}

const hero: PlaceholderImage = {
  accent: '#b5453c',
  alt: 'The Hollow Crown hero image with a crown and stage light',
  background: '#08070b',
  eyebrow: 'Basement Plays presents',
  foreground: '#e9dfd2',
  key: 'hollow-crown-hero',
  kind: 'hero',
  motif: 'crown',
  subtitle: 'A gothic tragedy of inheritance and rot',
  title: 'THE HOLLOW|CROWN',
}

const galleryImages: PlaceholderImage[] = [
  {
    accent: '#b5453c',
    alt: 'The cast of The Hollow Crown gathered around a long banquet table',
    background: '#0d0a0e',
    eyebrow: 'Production still',
    foreground: '#efe6d8',
    key: 'hollow-crown-gallery-1',
    kind: 'gallery',
    motif: 'stage',
    subtitle: 'Act I, scene ii',
    title: 'THE BANQUET|SCENE',
  },
  {
    accent: '#b5453c',
    alt: 'The Heir confronts the Regent under a cold spotlight',
    background: '#100c11',
    eyebrow: 'Production still',
    foreground: '#efe6d8',
    key: 'hollow-crown-gallery-2',
    kind: 'gallery',
    motif: 'stage',
    subtitle: 'Act II, scene i',
    title: 'THE HEIR|RETURNS',
  },
]

const payload = await getPayload({ config })

const existing = await payload.find({
  collection: 'productions',
  depth: 0,
  limit: 1,
  overrideAccess: true,
  where: {
    slug: {
      equals: 'the-hollow-crown',
    },
  },
})

if (existing.docs[0]) {
  console.log('Dev-only PlainDesign event already exists. Skipping.')
} else {
  console.log('Creating dev-only PlainDesign event: The Hollow Crown')

  const posterMedia = await upload(payload, poster)
  const heroMedia = await upload(payload, hero)
  const gallery: {
    caption?: string
    category: 'on-stage' | 'behind-the-scenes'
    image: number
  }[] = []

  for (const [index, image] of galleryImages.entries()) {
    const media = await upload(payload, image)
    gallery.push({
      caption: image.subtitle,
      category: index === 0 ? 'on-stage' : 'behind-the-scenes',
      image: media.id,
    })
  }

  const production = await payload.create({
    collection: 'productions',
    data: {
      cast: [
        { name: 'Anushka Perera', role: 'Queen Isolde' },
        { name: 'Ravi Fernando', role: 'The Heir' },
        { name: 'Maya Jayasinghe', role: 'The Fool' },
      ],
      crew: [
        { name: 'Nadeesha Wickramasinghe', role: 'Director' },
        { name: 'Kasun Mendis', role: 'Set & Lighting' },
      ],
      directedBy: 'Nadeesha Wickramasinghe',
      gallery,
      heroImage: heroMedia.id,
      partners: [
        { name: 'The Basement Theatre', role: 'Venue Partner' },
      ],
      poster: posterMedia.id,
      presentedBy: 'A Basement production',
      shows: [
        { startsAt: new Date('2026-12-04T19:30:00+05:30').toISOString() },
        {
          cancelled: true,
          startsAt: new Date('2026-12-05T19:30:00+05:30').toISOString(),
        },
      ],
      slug: 'the-hollow-crown',
      synopsis: richText([
        'A dev-only placeholder production for checking the fallback design. The old king is dead, and the crown will not stay quiet.',
        'When a fractured family gathers to divide a rotting estate, a stranger arrives with a claim no one can prove.',
      ]),
      tagline: 'A gothic tragedy of inheritance and rot.',
      ticketing: {
        columns: 10,
        paymentMode: 'off',
        selectionMode: 'quantity',
        stagePosition: 'front',
        tiers: [{ count: 40, name: 'General', price: 1500 }],
      },
      title: 'The Hollow Crown',
      venue: 'The Basement Theatre, Colombo',
      writtenBy: 'Dev Seed',
    },
    overrideAccess: true,
  })

  console.log(`Created dev-only event (id ${production.id}).`)
}

await payload.destroy()
process.exit(0)

async function upload(payload: Payload, image: PlaceholderImage) {
  const filename = `${image.key}.jpg`
  const existing = await payload.find({
    collection: 'media',
    depth: 0,
    limit: 1,
    overrideAccess: true,
    where: {
      filename: {
        equals: filename,
      },
    },
  })

  if (existing.docs[0]) {
    return existing.docs[0]
  }

  const filePath = await renderPlaceholder(image)

  return payload.create({
    collection: 'media',
    data: {
      alt: image.alt,
    },
    filePath,
    overrideAccess: true,
  })
}

function richText(paragraphs: string[]): Synopsis {
  return {
    root: {
      children: paragraphs.map((text) => ({
        children: [
          {
            detail: 0,
            format: 0,
            mode: 'normal',
            style: '',
            text,
            type: 'text',
            version: 1,
          },
        ],
        direction: 'ltr',
        format: '',
        indent: 0,
        type: 'paragraph',
        version: 1,
      })),
      direction: 'ltr',
      format: '',
      indent: 0,
      type: 'root',
      version: 1,
    },
  }
}
