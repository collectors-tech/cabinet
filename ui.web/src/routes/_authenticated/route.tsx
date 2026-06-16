import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { normalizeAuthRedirectTarget } from '@/lib/auth-redirect'
import { AuthenticatedLayout } from '@/components/layout/authenticated-layout'

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: ({ location }) => {
    const token = useAuthStore.getState().auth.accessToken
    if (!token) {
      const redirectTarget = normalizeAuthRedirectTarget(location.href)
      throw redirect({
        to: '/sign-in',
        search: redirectTarget ? { redirect: redirectTarget } : undefined,
      })
    }
  },
  component: AuthenticatedLayout,
})
