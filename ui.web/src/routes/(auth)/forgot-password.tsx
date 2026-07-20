import { createFileRoute } from '@tanstack/react-router'
import { ForgotPassword } from '@/features/auth/forgot-password'

export const Route = createFileRoute('/(auth)/forgot-password')({
  beforeLoad: async () => {
    try {
      const response = await fetch('/api/auth/provider-options')
      const payload = response.ok
        ? ((await response.json()) as { identity_mode?: string })
        : null
      if (payload?.identity_mode === 'zitadel') {
        window.location.replace('/api/auth/zitadel/login?intent=recover')
      }
    } catch {
      // Render the local recovery-safe page when provider discovery is offline.
    }
  },
  component: ForgotPassword,
})
