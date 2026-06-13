import { createFileRoute } from '@tanstack/react-router'
import { RouteErrorPage } from '@/features/errors/route-error-page'

export const Route = createFileRoute('/_authenticated/errors/$error')({
  component: RouteComponent,
})

function RouteComponent() {
  const { error } = Route.useParams()
  return <RouteErrorPage error={error} />
}
