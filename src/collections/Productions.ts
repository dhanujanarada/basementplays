import type { CollectionConfig } from 'payload'

import { anyone, isStaff } from './access'
import { slugify, uniqueSlug } from './hooks/slug'

export const Productions: CollectionConfig = {
  slug: 'productions',
  labels: {
    singular: 'Event',
    plural: 'Events',
  },
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
              name: 'presentedBy',
              type: 'text',
              defaultValue: 'A Basement production',
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
              name: 'synopsis',
              type: 'richText',
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
          label: 'Shows & tickets',
          fields: [
            {
              name: 'venue',
              type: 'text',
              admin: {
                description:
                  'Default venue for every show. A show can override it.',
              },
            },
            {
              name: 'shows',
              type: 'array',
              fields: [
                {
                  name: 'startsAt',
                  type: 'date',
                  required: true,
                  admin: {
                    date: {
                      pickerAppearance: 'dayAndTime',
                    },
                  },
                },
                {
                  name: 'venue',
                  type: 'text',
                  admin: {
                    description: 'Optional. Overrides the event venue.',
                  },
                },
                {
                  name: 'cancelled',
                  type: 'checkbox',
                  defaultValue: false,
                },
              ],
            },
            {
              type: 'group',
              name: 'ticketing',
              label: 'Ticketing',
              fields: [
                {
                  name: 'paymentMode',
                  type: 'select',
                  required: true,
                  defaultValue: 'bank_transfer',
                  options: [
                    {
                      label: 'Bank transfer',
                      value: 'bank_transfer',
                    },
                    {
                      label: 'Online (PayHere sandbox)',
                      value: 'online',
                    },
                    {
                      label: 'Off (WhatsApp fallback)',
                      value: 'off',
                    },
                  ],
                },
                {
                  name: 'selectionMode',
                  type: 'select',
                  required: true,
                  defaultValue: 'quantity',
                  options: [
                    {
                      label: 'Quantity (tier boxes)',
                      value: 'quantity',
                    },
                    {
                      label: 'Seats (seat map)',
                      value: 'seats',
                    },
                  ],
                },
                {
                  name: 'stagePosition',
                  type: 'select',
                  required: true,
                  defaultValue: 'front',
                  options: [
                    {
                      label: 'Front (above the tiers)',
                      value: 'front',
                    },
                    {
                      label: 'Back (below the tiers)',
                      value: 'back',
                    },
                    {
                      label: 'None',
                      value: 'none',
                    },
                  ],
                },
                {
                  name: 'numbering',
                  type: 'checkbox',
                  defaultValue: false,
                  admin: {
                    condition: (data) =>
                      data?.ticketing?.selectionMode === 'seats',
                    description: 'Number seats in seats mode.',
                  },
                },
                {
                  name: 'columns',
                  type: 'number',
                  defaultValue: 10,
                  min: 1,
                  admin: {
                    condition: (data) =>
                      data?.ticketing?.selectionMode === 'seats',
                    description: 'Seat grid width in seats mode.',
                  },
                },
                {
                  name: 'tiers',
                  type: 'array',
                  required: true,
                  minRows: 1,
                  admin: {
                    description:
                      'At least one tier. The order here is stage order: first is nearest the stage.',
                  },
                  fields: [
                    {
                      name: 'name',
                      type: 'text',
                      required: true,
                    },
                    {
                      name: 'price',
                      type: 'number',
                      required: true,
                      min: 0,
                      admin: {
                        description: 'Whole rupees, e.g. 2500.',
                      },
                    },
                    {
                      name: 'count',
                      type: 'number',
                      required: true,
                      min: 1,
                    },
                  ],
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
                {
                  name: 'category',
                  type: 'select',
                  defaultValue: 'on-stage',
                  options: [
                    {
                      label: 'On stage',
                      value: 'on-stage',
                    },
                    {
                      label: 'Behind the scenes',
                      value: 'behind-the-scenes',
                    },
                  ],
                },
              ],
            },
          ],
        },
      ],
    },
  ],
}
