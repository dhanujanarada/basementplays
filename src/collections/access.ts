import type { Access } from 'payload'

type UserWithRole = {
  role?: 'admin' | 'editor'
} | null

export const anyone: Access = () => true

export const isStaff: Access = ({ req: { user } }) => Boolean(user)

export const isAdmin: Access = ({ req: { user } }) =>
  (user as UserWithRole)?.role === 'admin'

export const canCreateUser: Access = async ({ req }) => {
  if ((req.user as UserWithRole)?.role === 'admin') {
    return true
  }

  const { totalDocs } = await req.payload.count({
    collection: 'users',
  })

  return totalDocs === 0
}
