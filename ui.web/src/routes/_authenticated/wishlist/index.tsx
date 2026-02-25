import { createFileRoute } from '@tanstack/react-router'
import { Wishlist } from '@/features/wishlist'

export const Route = createFileRoute('/_authenticated/wishlist/')({
  component: Wishlist,
})
