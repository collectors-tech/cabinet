import { createFileRoute } from '@tanstack/react-router'
import { SignUp } from '@/features/auth/sign-up'

export const Route = createFileRoute('/(auth)/sign-up')({
  beforeLoad: async () => {
    try {
      const response = await fetch('/api/auth/provider-options')
      const payload = response.ok
        ? ((await response.json()) as { identity_mode?: string })
        : null
      if (payload?.identity_mode === 'zitadel') {
        window.location.replace('/api/auth/zitadel/login?intent=register')
      }
    } catch {
      // Render the local recovery-safe page when provider discovery is offline.
    }
  },
  component: SignUp,
})
