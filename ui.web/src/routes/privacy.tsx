import { createFileRoute } from '@tanstack/react-router'
import { PrivacyPolicy } from '@/features/auth/privacy-policy'

export const Route = createFileRoute('/privacy')({
  component: PrivacyPolicy,
})
