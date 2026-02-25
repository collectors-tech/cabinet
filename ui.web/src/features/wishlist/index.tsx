import { Tasks } from '@/features/tasks'

export function Wishlist() {
  return (
    <Tasks
      title='Wishlist'
      description='Track discovered items, target entries, and scheduling priority.'
      routePath='/_authenticated/wishlist/'
    />
  )
}
