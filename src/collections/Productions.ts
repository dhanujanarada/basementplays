import type { CollectionConfig } from 'payload'

import { anyone, isStaff } from './access'
import { slugify, uniqueSlug } from './hooks/slug'

export const Productions: CollectionConfig = {
  slug: 'productions',
  admin: {
    group: 'Content',
    useAsTitle: 'title',
    defaultColumns: ['title', 'slug', 'updatedAt'],
  },
  access: {
    read: anyone,
    create: isStaff,
    update: isStaff,
    delete: isStaff,
  },
  hooks: {
    beforeValidate: [
      async ({ data, req, originalDoc }) => {
        if (!data) {
          return data
        }

        const values = data as {
          slug?: string
          title?: string
        }

        if (!values.slug && values.title) {
          values.slug = await uniqueSlug(
            slugify(values.title),
            req,
            originalDoc?.id,
          )
        } else if (values.slug) {
          values.slug = slugify(values.slug)
        }

        return data
      },
    ],
  },
  fields: [
    {
      type: 'tabs',
      tabs: [
        {
          label: 'Details',
          fields: [
            {
              name: 'title',
              type: 'text',
              required: true,
            },
            {
              name: 'slug',
              type: 'text',
              required: true,
              unique: true,
              index: true,
              admin: {
                description:
                  'Generated from the title. Edit only if you need a custom URL.',
              },
            },
            {
              name: 'tagline',
              type: 'text',
            },
            {
              name: 'synopsis',
              type: 'richText',
            },
            {
              name: 'presentedBy',
              type: 'text',
              admin: {
                description: 'e.g. A Basement production',
              },
            },
            {
              name: 'writtenBy',
              type: 'text',
            },
            {
              name: 'directedBy',
              type: 'text',
            },
            {
              name: 'poster',
              type: 'upload',
              relationTo: 'media',
              required: true,
            },
            {
              name: 'heroImage',
              type: 'upload',
              relationTo: 'media',
            },
          ],
        },
        {
          label: 'Partners',
          fields: [
            {
              name: 'partners',
              type: 'array',
              admin: {
                description:
                  'Optional. Logos may be omitted; the public design shows the partner name as text.',
              },
              fields: [
                {
                  name: 'name',
                  type: 'text',
                  required: true,
                },
                {
                  name: 'role',
                  type: 'text',
                  admin: {
                    description:
                      'e.g. Location Partner, Promotional Partner',
                  },
                },
                {
                  name: 'logo',
                  type: 'upload',
                  relationTo: 'media',
                },
              ],
            },
          ],
        },
        {
          label: 'People',
          fields: [
            {
              name: 'cast',
              type: 'array',
              fields: [
                {
                  name: 'name',
                  type: 'text',
                  required: true,
                },
                {
                  name: 'role',
                  type: 'text',
                },
                {
                  name: 'photo',
                  type: 'upload',
                  relationTo: 'media',
                },
              ],
            },
            {
              name: 'crew',
              type: 'array',
              fields: [
                {
                  name: 'name',
                  type: 'text',
                  required: true,
                },
                {
                  name: 'role',
                  type: 'text',
                },
              ],
            },
          ],
        },
        {
          label: 'Gallery',
          fields: [
            {
              name: 'gallery',
              type: 'array',
              admin: {
                description:
                  'Show photos. Add a caption for each image where useful.',
              },
              fields: [
                {
                  name: 'image',
                  type: 'upload',
                  relationTo: 'media',
                  required: true,
                },
                {
                  name: 'caption',
                  type: 'text',
                },
              ],
            },
          ],
        },
      ],
    },
  ],
}
