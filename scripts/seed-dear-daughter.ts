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
type ShowStatus = 'open' | 'few-left' | 'sold-out' | 'cancelled'

const productionSlug = 'dear-daughter'
const productionTitle = 'Dear Daughter,'
const runLabel = '2026 Season'
const venue = 'The Black Bean Cafe, Kandy'
const placeholderStartsAt = '2026-11-14T19:00:00+05:30'
const placeholderDateLabel = '14 November 2026 at 7:00 pm (Asia/Colombo)'
const ticketUrl = `https://wa.me/94764508448?text=${encodeURIComponent(
  `Hi! I would like to reserve seats for ${productionTitle}`,
)}`

const databaseUrl = process.env.DATABASE_URL || 'file:./payload.db'

if (databaseUrl.startsWith('file:')) {
  const databasePath = databaseUrl.slice('file:'.length)

  if (!fs.existsSync(databasePath)) {
    throw new Error(
      'payload.db does not exist yet. Run `npm run dev` once to create the schema, then run `npm run seed:dear-daughter`.',
    )
  }
}

const palette = {
  accent: '#9a4b38',
  background: '#4f7a56',
  deep: '#3d6247',
  foreground: '#f1ead9',
  wood: '#a8794a',
}

const poster: PlaceholderImage = {
  accent: palette.accent,
  alt: 'Placeholder poster for Dear Daughter, with a deep green background and terracotta lettering',
  background: palette.deep,
  eyebrow: 'A Basement production',
  foreground: palette.foreground,
  key: 'dear-daughter-poster',
  kind: 'poster',
  motif: 'stage',
  subtitle: 'A new play by Dhisal Kapuge',
  title: 'DEAR|DAUGHTER',
}

const hero: PlaceholderImage = {
  accent: palette.accent,
  alt: 'Placeholder hero image for Dear Daughter, with a green stage-light gradient',
  background: palette.background,
  eyebrow: 'A Basement production',
  foreground: palette.foreground,
  key: 'dear-daughter-hero',
  kind: 'hero',
  motif: 'stage',
  subtitle: 'A new play by Dhisal Kapuge',
  title: 'DEAR|DAUGHTER',
}

const galleryImages: (PlaceholderImage & { caption: string })[] = [
  {
    accent: palette.accent,
    alt: 'Placeholder production still: the family gathers around a kitchen table',
    background: palette.deep,
    caption: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit.',
    eyebrow: 'Production still',
    foreground: palette.foreground,
    key: 'dear-daughter-gallery-1',
    kind: 'gallery',
    motif: 'stage',
    subtitle: 'Act I, scene i',
    title: 'THE KITCHEN|TABLE',
  },
  {
    accent: palette.wood,
    alt: 'Placeholder production still: a letter is read aloud',
    background: palette.background,
    caption:
      'Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.',
    eyebrow: 'Production still',
    foreground: palette.foreground,
    key: 'dear-daughter-gallery-2',
    kind: 'gallery',
    motif: 'stage',
    subtitle: 'Act I, scene iii',
    title: 'THE|LETTER',
  },
  {
    accent: palette.accent,
    alt: 'Placeholder production still: two figures talk in the garden',
    background: palette.deep,
    caption:
      'Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.',
    eyebrow: 'Production still',
    foreground: palette.foreground,
    key: 'dear-daughter-gallery-3',
    kind: 'gallery',
    motif: 'stage',
    subtitle: 'Act II, scene i',
    title: 'THE|GARDEN',
  },
  {
    accent: palette.wood,
    alt: 'Placeholder production still: two chairs face each other under a lamp',
    background: palette.background,
    caption:
      'Duis aute irure dolor in reprehenderit in voluptate velit esse cillum.',
    eyebrow: 'Production still',
    foreground: palette.foreground,
    key: 'dear-daughter-gallery-4',
    kind: 'gallery',
    motif: 'stage',
    subtitle: 'Act II, scene ii',
    title: 'TWO|CHAIRS',
  },
  {
    accent: palette.accent,
    alt: 'Placeholder production still: a figure looks out of a window',
    background: palette.deep,
    caption:
      'Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia.',
    eyebrow: 'Production still',
    foreground: palette.foreground,
    key: 'dear-daughter-gallery-5',
    kind: 'gallery',
    motif: 'stage',
    subtitle: 'Act II, scene iv',
    title: 'THE|WINDOW',
  },
  {
    accent: palette.wood,
    alt: 'Placeholder production still: the final scene in low light',
    background: palette.background,
    caption:
      'Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit.',
    eyebrow: 'Production still',
    foreground: palette.foreground,
    key: 'dear-daughter-gallery-6',
    kind: 'gallery',
    motif: 'stage',
    subtitle: 'Act III, scene ii',
    title: 'LAST|LIGHT',
  },
  {
    accent: palette.accent,
    alt: 'Placeholder production still: a table set for a family meal',
    background: palette.deep,
    caption:
      'Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet.',
    eyebrow: 'Production still',
    foreground: palette.foreground,
    key: 'dear-daughter-gallery-7',
    kind: 'gallery',
    motif: 'stage',
    subtitle: 'Act III, scene iii',
    title: 'THE|MEAL',
  },
  {
    accent: palette.wood,
    alt: 'Placeholder production still: the cast takes a final bow',
    background: palette.background,
    caption:
      'Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit.',
    eyebrow: 'Production still',
    foreground: palette.foreground,
    key: 'dear-daughter-gallery-8',
    kind: 'gallery',
    motif: 'stage',
    subtitle: 'Curtain call',
    title: 'FINAL|BOW',
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
      equals: productionSlug,
    },
  },
})

let production = existing.docs[0]

if (production) {
  console.log(
    `Production already exists: ${productionTitle} (id ${production.id}). Skipping content creation.`,
  )
} else {
  console.log(`Creating production: ${productionTitle}`)

  const posterMedia = await uploadPlaceholder(payload, poster)
  const heroMedia = await uploadPlaceholder(payload, hero)
  const gallery: { caption?: string; image: number }[] = []

  for (const image of galleryImages) {
    const media = await uploadPlaceholder(payload, image)
    gallery.push({
      caption: image.caption,
      image: media.id,
    })
  }

  production = await payload.create({
    collection: 'productions',
    data: {
      cast: [
        { name: 'Cast Member 1', role: 'Role 1' },
        { name: 'Cast Member 2', role: 'Role 2' },
        { name: 'Cast Member 3', role: 'Role 3' },
        { name: 'Cast Member 4', role: 'Role 4' },
      ],
      crew: [
        { name: 'Crew Member 1', role: 'Crew Role 1' },
        { name: 'Crew Member 2', role: 'Crew Role 2' },
        { name: 'Crew Member 3', role: 'Crew Role 3' },
      ],
      directedBy: 'Dhisal Kapuge',
      gallery,
      heroImage: heroMedia.id,
      partners: [
        { name: 'The Black Bean Cafe', role: 'Location Partner' },
        {
          name: 'Promotional Partner TBC',
          role: 'Promotional Partner',
        },
      ],
      poster: posterMedia.id,
      presentedBy: 'A Basement production',
      slug: productionSlug,
      synopsis: richText([
        'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.',
        'Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.',
        'Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.',
      ]),
      tagline: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit.',
      title: productionTitle,
      writtenBy: 'Dhisal Kapuge',
    },
    overrideAccess: true,
  })

  console.log(`Created production: ${productionTitle} (id ${production.id})`)
}

const runId = await seedRun(payload, production.id)

console.log('')
console.log('!'.repeat(78))
console.log('!!  PLACEHOLDER SHOW DATE — SET THE REAL DATE IN BACKSTAGE             !!')
console.log(`!!  Seeded date: ${placeholderDateLabel}`)
console.log('!!  The poster cuts off the day of the month, so this date is NOT final.')
console.log(
  `!!  Backstage: http://localhost:3000/backstage/collections/runs/${runId}`,
)
console.log('!'.repeat(78))
console.log('')

process.exit(0)

async function seedRun(
  payload: Payload,
  productionId: number,
): Promise<number> {
  const existingRun = await payload.find({
    collection: 'runs',
    depth: 0,
    limit: 1,
    overrideAccess: true,
    where: {
      and: [
        {
          production: {
            equals: productionId,
          },
        },
        {
          label: {
            equals: runLabel,
          },
        },
      ],
    },
  })

  if (existingRun.docs.length > 0) {
    console.log(`Run already exists: ${runLabel}. Skipping.`)
    return existingRun.docs[0].id
  }

  const run = await payload.create({
    collection: 'runs',
    data: {
      label: runLabel,
      production: productionId,
      shows: [
        {
          startsAt: new Date(placeholderStartsAt).toISOString(),
          status: 'few-left' satisfies ShowStatus,
          price: 'Rs. 2500',
          seatsNote: 'Less than 20 seats left',
          ticketUrl,
        },
      ],
      venue,
    },
    overrideAccess: true,
  })

  console.log(`Created run: ${runLabel} at ${venue}`)
  return run.id
}

async function uploadPlaceholder(
  payload: Payload,
  image: PlaceholderImage,
) {
  const filename = `${image.key}.jpg`
  const existingMedia = await payload.find({
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

  if (existingMedia.docs[0]) {
    console.log(`Media already exists: ${filename}`)
    return existingMedia.docs[0]
  }

  const filePath = await renderPlaceholder(image)

  const media = await payload.create({
    collection: 'media',
    data: {
      alt: image.alt,
    },
    filePath,
    overrideAccess: true,
  })

  console.log(`Created media: ${filename}`)
  return media
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
