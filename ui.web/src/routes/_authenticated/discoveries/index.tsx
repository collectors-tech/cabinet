import { createFileRoute } from '@tanstack/react-router'
import { Discover } from '@/features/discover'

export const Route = createFileRoute('/_authenticated/discoveries/')({
  component: Discover,
})
