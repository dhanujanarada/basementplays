import type { PayloadRequest } from 'payload'

export function slugify(value: string): string {
  const slug = value
    .toLowerCase()
    .trim()
    .replace(/['’]/g, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')

  return slug || 'production'
}

export async function uniqueSlug(
  base: string,
  req: PayloadRequest,
  currentId?: number | string,
): Promise<string> {
  let candidate = base
  let suffix = 2

  while (true) {
    const { docs } = await req.payload.find({
      collection: 'productions',
      depth: 0,
      limit: 1,
      where: {
        slug: {
          equals: candidate,
        },
      },
    })

    const existing = docs[0]

    if (!existing || String(existing.id) === String(currentId)) {
      return candidate
    }

    candidate = `${base}-${suffix}`
    suffix += 1
  }
}
