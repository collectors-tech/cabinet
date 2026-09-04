import { useCallback, useEffect, useState } from 'react'
import { getRouteApi } from '@tanstack/react-router'
import { Users as UsersIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { UsersDialogs } from './components/users-dialogs'
import { UsersPrimaryButtons } from './components/users-primary-buttons'
import { UsersProvider } from './components/users-provider'
import { UsersTable } from './components/users-table'
import { type User } from './data/schema'

const route = getRouteApi('/_authenticated/users/')

export function Users() {
  const search = route.useSearch()
  const navigate = route.useNavigate()
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState<string | null>(null)

  const loadUsers = useCallback(async () => {
    setLoading(true)
    setLoadError(null)
    try {
      const response = await fetch('/api/users')
      if (!response.ok) {
        throw new Error(`users_fetch_failed_${response.status}`)
      }
      const payload = (await response.json()) as { users?: User[] }
      setUsers(payload.users ?? [])
    } catch (error) {
      const message =
        error instanceof Error ? error.message : 'users_fetch_failed'
      setLoadError(message)
      setUsers([])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadUsers()
  }, [loadUsers])

  return (
    <UsersProvider>
      <Header fixed data-testid='users-shell-header'>
        <Search />
        <HeaderTitle
          title='Users'
          description='Manage users and roles.'
          icon={UsersIcon}
          testId='users-header-title'
          iconTestId='users-page-icon'
        />
        <div
          className='ms-auto flex min-w-0 items-center gap-3'
          data-header-title-avoid='true'
        >
          <div
            className='flex min-w-0 flex-wrap items-center justify-end gap-2'
            data-testid='users-global-header-actions'
          >
            <UsersPrimaryButtons />
          </div>
          <Separator
            orientation='vertical'
            className='h-6'
            data-testid='users-header-action-separator'
          />
          <div className='flex items-center gap-4'>
            <LanguageSwitch />
            <ThemeSwitch />
            <ConfigDrawer />
            <ProfileDropdown />
          </div>
        </div>
      </Header>

      <Main fixed className='gap-3 sm:gap-4' data-testid='users-workspace'>
        {loadError ? (
          <div
            className='rounded-md border border-destructive/40 bg-destructive/10 p-4 text-sm'
            data-testid='users-load-error'
          >
            <p className='font-medium'>Users load failed</p>
            <p className='mt-1 text-muted-foreground'>{loadError}</p>
            <Button
              className='mt-3'
              variant='outline'
              size='sm'
              onClick={() => void loadUsers()}
            >
              Retry
            </Button>
          </div>
        ) : null}
        {!loadError && loading ? (
          <div className='rounded-md border p-6 text-sm text-muted-foreground'>
            Loading users...
          </div>
        ) : (
          <UsersTable
            data={users}
            search={search}
            navigate={navigate}
            onMutated={loadUsers}
          />
        )}
      </Main>

      <UsersDialogs onMutated={loadUsers} />
    </UsersProvider>
  )
}
