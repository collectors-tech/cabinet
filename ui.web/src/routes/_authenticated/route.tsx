import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { normalizeAuthRedirectTarget } from '@/lib/auth-redirect'
import { AuthenticatedLayout } from '@/components/layout/authenticated-layout'

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: async ({ location }) => {
    const current = useAuthStore.getState().auth
    if (current.accessToken) {
      return
    }

    try {
      let response = await fetch('/api/auth/zitadel/session', {
        credentials: 'same-origin',
        headers: { Accept: 'application/json' },
      })
      if (response.status === 401) {
        const refreshed = await fetch('/api/auth/zitadel/refresh', {
          method: 'POST',
          credentials: 'same-origin',
          headers: { Accept: 'application/json' },
        })
        if (refreshed.ok) {
          response = await fetch('/api/auth/zitadel/session', {
            credentials: 'same-origin',
            headers: { Accept: 'application/json' },
          })
        }
      }
      if (response.ok) {
        const payload = (await response.json()) as {
          expires_at?: string
          user?: {
            subject?: string
            email?: string
            roles?: string[]
          }
        }
        const subject = payload.user?.subject?.trim()
        if (subject) {
          useAuthStore.getState().auth.setRemoteUser({
            accountNo: subject,
            email: payload.user?.email?.trim() || 'account@cabinet.invalid',
            role: Array.isArray(payload.user?.roles) ? payload.user.roles : [],
            exp: payload.expires_at
              ? new Date(payload.expires_at).getTime()
              : Date.now() + 60 * 60 * 1000,
          })
          return
        }
      }
    } catch {
      // The redirect below is the fail-closed path for an unavailable session.
    }

    const redirectTarget = normalizeAuthRedirectTarget(location.href)
    throw redirect({
      to: '/sign-in',
      search: redirectTarget ? { redirect: redirectTarget } : undefined,
    })
  },
  component: AuthenticatedLayout,
})
