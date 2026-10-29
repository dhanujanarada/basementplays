import config from '@payload-config'
import { getPayload } from 'payload'

import type { Production, Run } from '@/payload-types'

export async function getAllProductions(): Promise<Production[]> {
  const payload = await getPayload({ config })

  const result = await payload.find({
    collection: 'productions',
    depth: 2,
    limit: 100,
    overrideAccess: false,
    sort: 'title',
  })

  return result.docs
}

export async function getProduction(slug: string): Promise<Production | null> {
  const payload = await getPayload({ config })

  const result = await payload.find({
    collection: 'productions',
    depth: 2,
    limit: 1,
    overrideAccess: false,
    where: {
      slug: {
        equals: slug,
      },
    },
  })

  return result.docs[0] ?? null
}

export async function getRunsForProduction(
  productionId: number,
): Promise<Run[]> {
  const payload = await getPayload({ config })

  const result = await payload.find({
    collection: 'runs',
    depth: 0,
    limit: 100,
    overrideAccess: false,
    sort: '-createdAt',
    where: {
      production: {
        equals: productionId,
      },
    },
  })

  return result.docs
}
