import { TriangleAlert } from 'lucide-react'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { Button } from '@/components/ui/button'

type RouteErrorState = {
  status: string
  title: string
  body: string
  showBack: boolean
  home: boolean
  learnMore?: boolean
}

type RouteErrorPageProps = {
  error: string
}

export function RouteErrorPage({ error }: RouteErrorPageProps) {
  const state = getErrorState(error)

  return (
    <>
      <Header fixed className='border-b'>
        <Search />
        <HeaderTitle
          title='Error'
          description='Review the current route error state.'
          icon={TriangleAlert}
          testId='error-header-title'
          iconTestId='error-page-icon'
        />
        <div
          className='ms-auto flex items-center space-x-4'
          data-header-title-avoid='true'
        >
          <LanguageSwitch />
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>
      <div className='flex-1'>
        <div className='m-auto flex min-h-[calc(100svh-4rem)] w-full flex-col items-center justify-center gap-2'>
          <h1 className='text-[7rem] leading-tight font-bold'>
            {state.status}
          </h1>
          <span className='font-medium'>{state.title}</span>
          <p className='text-center text-muted-foreground'>{state.body}</p>
          <div className='mt-6 flex gap-4'>
            {state.showBack && (
              <Button variant='outline' onClick={() => window.history.back()}>
                Go Back
              </Button>
            )}
            {state.home && (
              <Button onClick={() => window.location.assign('/')}>
                Back to Home
              </Button>
            )}
            {state.learnMore && <Button variant='outline'>Learn more</Button>}
          </div>
        </div>
      </div>
    </>
  )
}

function getErrorState(error: string): RouteErrorState {
  switch (error) {
    case 'unauthorized':
      return {
        status: '401',
        title: 'Unauthorized Access',
        body: 'Please log in with the appropriate credentials to access this resource.',
        showBack: true,
        home: true,
      }
    case 'forbidden':
      return {
        status: '403',
        title: 'Access Forbidden',
        body: "You don't have necessary permission to view this resource.",
        showBack: true,
        home: true,
      }
    case 'internal-server-error':
      return {
        status: '500',
        title: "Oops! Something went wrong :')",
        body: 'We apologize for the inconvenience. Please try again later.',
        showBack: true,
        home: true,
      }
    case 'maintenance-error':
      return {
        status: '503',
        title: 'Website is under maintenance!',
        body: "The site is not available at the moment. We'll be back online shortly.",
        showBack: false,
        home: false,
        learnMore: true,
      }
    default:
      return {
        status: '404',
        title: 'Oops! Page Not Found!',
        body: "It seems like the page you're looking for does not exist or might have been removed.",
        showBack: true,
        home: true,
      }
  }
}
