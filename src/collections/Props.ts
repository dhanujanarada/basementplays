import type { CollectionConfig } from 'payload'

import { isStaff } from './access'

export const Props: CollectionConfig = {
  slug: 'props',
  admin: {
    group: 'Backstage',
    useAsTitle: 'name',
    defaultColumns: ['name', 'status', 'production', 'assignedTo'],
  },
  access: {
    read: isStaff,
    create: isStaff,
    update: isStaff,
    delete: isStaff,
  },
  fields: [
    {
      name: 'production',
      type: 'relationship',
      relationTo: 'productions',
      required: true,
    },
    {
      name: 'name',
      type: 'text',
      required: true,
    },
    {
      name: 'scene',
      type: 'text',
    },
    {
      name: 'status',
      type: 'select',
      required: true,
      defaultValue: 'needed',
      options: [
        {
          label: 'Needed',
          value: 'needed',
        },
        {
          label: 'Sourced',
          value: 'sourced',
        },
        {
          label: 'Ready',
          value: 'ready',
        },
        {
          label: 'Returned',
          value: 'returned',
        },
      ],
    },
    {
      name: 'assignedTo',
      type: 'text',
    },
    {
      name: 'notes',
      type: 'textarea',
    },
  ],
}
