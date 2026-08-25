import { useEffect } from 'react'
import { Outlet, useLocation } from '@tanstack/react-router'
import { getCookie } from '@/lib/cookies'
import { getDocumentTitle } from '@/lib/document-title'
import { cn } from '@/lib/utils'
import { LayoutProvider } from '@/context/layout-provider'
import { SearchProvider } from '@/context/search-provider'
import { ShellWorkspaceProvider } from '@/context/shell-workspace-provider'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { GuidanceOverlay } from '@/components/guidance/guidance-overlay'
import { AppSidebar } from '@/components/layout/app-sidebar'
import { SkipToMain } from '@/components/skip-to-main'

type AuthenticatedLayoutProps = {
  children?: React.ReactNode
}

export function AuthenticatedLayout({ children }: AuthenticatedLayoutProps) {
  const defaultOpen = getCookie('sidebar_state') === 'true'
  const location = useLocation()

  useEffect(() => {
    document.title = getDocumentTitle(location.pathname)
  }, [location.pathname])

  return (
    <SearchProvider>
      <LayoutProvider>
        <ShellWorkspaceProvider>
          <SidebarProvider defaultOpen={defaultOpen}>
            <SkipToMain />
            <AppSidebar />
            <SidebarInset
              className={cn(
                // Set content container, so we can use container queries
                '@container/content min-w-0 overflow-x-clip',

                // If layout is fixed, set the height
                // to 100svh to prevent overflow
                'has-data-[layout=fixed]:h-svh',

                // If layout is fixed and sidebar is inset,
                // set the height to 100svh - spacing (total margins) to prevent overflow
                'peer-data-[variant=inset]:has-data-[layout=fixed]:h-[calc(100svh-(var(--spacing)*4))]'
              )}
            >
              {children ?? <Outlet />}
            </SidebarInset>
            <GuidanceOverlay />
          </SidebarProvider>
        </ShellWorkspaceProvider>
      </LayoutProvider>
    </SearchProvider>
  )
}
