import { StrictMode } from 'react'
import ReactDOM from 'react-dom/client'
import { AxiosError } from 'axios'
import {
  QueryCache,
  QueryClient,
  QueryClientProvider,
} from '@tanstack/react-query'
import { RouterProvider, createRouter } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useAuthStore } from '@/stores/auth-store'
import { normalizeAuthRedirectTarget } from '@/lib/auth-redirect'
import { handleServerError } from '@/lib/handle-server-error'
import { DirectionProvider } from './context/direction-provider'
import { FontProvider } from './context/font-provider'
import { ThemeProvider } from './context/theme-provider'
import './i18n'
// Generated Routes
import { routeTree } from './routeTree.gen'
// Styles
import './styles/index.css'

function queryClientToastHistory({
  title,
  summary,
  category = 'system',
}: {
  title: string
  summary: string
  category?: 'auth' | 'system'
}) {
  return {
    history: {
      title,
      summary,
      source_label: 'Global query client',
      category,
    },
  } as never
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error) => {
        // eslint-disable-next-line no-console
        if (import.meta.env.DEV) console.log({ failureCount, error })

        if (failureCount >= 0 && import.meta.env.DEV) return false
        if (failureCount > 3 && import.meta.env.PROD) return false

        return !(
          error instanceof AxiosError &&
          [401, 403].includes(error.response?.status ?? 0)
        )
      },
      refetchOnWindowFocus: import.meta.env.PROD,
      staleTime: 10 * 1000, // 10s
    },
    mutations: {
      onError: (error) => {
        handleServerError(error)

        if (error instanceof AxiosError) {
          if (error.response?.status === 304) {
            toast.error(
              'Content not modified!',
              queryClientToastHistory({
                title: 'Mutation content not modified',
                summary: 'Mutation request returned HTTP 304.',
              })
            )
          }
        }
      },
    },
  },
  queryCache: new QueryCache({
    onError: (error) => {
      if (error instanceof AxiosError) {
        if (error.response?.status === 401) {
          toast.error(
            'Session expired!',
            queryClientToastHistory({
              title: 'Session expired',
              summary: 'Query request returned HTTP 401 and reset auth state.',
              category: 'auth',
            })
          )
          useAuthStore.getState().auth.reset()
          const redirect = normalizeAuthRedirectTarget(
            `${router.history.location.href}`
          )
          router.navigate({
            to: '/sign-in',
            search: redirect ? { redirect } : undefined,
          })
        }
        if (error.response?.status === 500) {
          toast.error(
            'Internal Server Error!',
            queryClientToastHistory({
              title: 'Internal server error',
              summary: 'Query request returned HTTP 500.',
            })
          )
          // Only navigate to error page in production to avoid disrupting HMR in development
          if (import.meta.env.PROD) {
            router.navigate({ to: '/500' })
          }
        }
        if (error.response?.status === 403) {
          // router.navigate("/forbidden", { replace: true });
        }
      }
    },
  }),
})

// Create a new router instance
const router = createRouter({
  routeTree,
  context: { queryClient },
  defaultPreload: 'intent',
  defaultPreloadStaleTime: 0,
})

// Register the router instance for type safety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

// Render the app
const rootElement = document.getElementById('root')!
if (!rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement)
  root.render(
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider>
          <FontProvider>
            <DirectionProvider>
              <RouterProvider router={router} />
            </DirectionProvider>
          </FontProvider>
        </ThemeProvider>
      </QueryClientProvider>
    </StrictMode>
  )
}
