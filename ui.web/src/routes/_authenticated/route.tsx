import { createFileRoute, redirect } from '@tanstack/react-router'
import { AuthenticatedLayout } from '@/components/layout/authenticated-layout'
import { normalizeAuthRedirectTarget } from '@/lib/auth-redirect'
import { useAuthStore } from '@/stores/auth-store'

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
