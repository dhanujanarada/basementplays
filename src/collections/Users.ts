import type { CollectionConfig } from 'payload'

import { canCreateUser, isAdmin } from './access'

export const Users: CollectionConfig = {
  slug: 'users',
  admin: {
    group: 'System',
    useAsTitle: 'email',
  },
  auth: true,
  access: {
    create: canCreateUser,
    read: isAdmin,
    update: isAdmin,
    delete: isAdmin,
    unlock: isAdmin,
  },
  fields: [
    {
      name: 'role',
      type: 'select',
      required: true,
      defaultValue: 'editor',
      options: [
        {
          label: 'Admin',
          value: 'admin',
        },
        {
          label: 'Editor',
          value: 'editor',
        },
      ],
      saveToJWT: true,
    },
  ],
}
