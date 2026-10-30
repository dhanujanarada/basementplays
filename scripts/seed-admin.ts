import 'dotenv/config'

import fs from 'node:fs'

import type { Payload } from 'payload'
import { getPayload } from 'payload'

import config from '../src/payload.config'

type Role = 'admin' | 'member'

const adminEmail = process.env.SEED_ADMIN_EMAIL
const adminPassword = process.env.SEED_ADMIN_PASSWORD
const memberEmail = process.env.SEED_MEMBER_EMAIL
const memberPassword = process.env.SEED_MEMBER_PASSWORD
const databaseUrl = process.env.DATABASE_URL || 'file:./payload.db'

if (!adminEmail || !adminPassword) {
  throw new Error(
    'Set SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD before running this script.',
  )
}

if (databaseUrl.startsWith('file:')) {
  const databasePath = databaseUrl.slice('file:'.length)

  if (!fs.existsSync(databasePath)) {
    throw new Error(
      'payload.db does not exist yet. Run `npm run dev` once to create the schema, then run `npm run seed:admin`.',
    )
  }
}

const payload = await getPayload({ config })

await ensureUser(payload, adminEmail, adminPassword, 'admin')

if (memberEmail && memberPassword) {
  await ensureUser(payload, memberEmail, memberPassword, 'member')
} else {
  console.log(
    'Skipping member user seed. Set SEED_MEMBER_EMAIL and SEED_MEMBER_PASSWORD to create one.',
  )
}

process.exit(0)

async function ensureUser(
  payload: Payload,
  email: string,
  password: string,
  role: Role,
): Promise<void> {
  const existing = await payload.find({
    collection: 'users',
    depth: 0,
    limit: 1,
    overrideAccess: true,
    where: {
      email: {
        equals: email,
      },
    },
  })

  if (existing.docs.length > 0) {
    console.log(`${role} user already exists: ${email}`)
    return
  }

  await payload.create({
    collection: 'users',
    data: {
      email,
      password,
      role,
    },
    overrideAccess: true,
  })

  console.log(`Created ${role} user: ${email}`)
}
