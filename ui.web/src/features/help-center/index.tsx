import { BookOpenText } from 'lucide-react'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { LanguageSwitch } from '@/components/language-switch'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export function HelpCenter() {
  return (
    <>
      <Header fixed>
        <Search />
        <div className='ms-auto flex items-center space-x-4'>
          <LanguageSwitch />
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>

      <Main className='space-y-6'>
        <div className='space-y-2'>
          <h1 className='text-2xl font-bold tracking-tight'>Help Center</h1>
          <p className='text-muted-foreground'>Find guides, diagnostics references, and support workflows.</p>
        </div>

        <Card data-testid='help-center-placeholder'>
          <CardHeader>
            <CardTitle className='flex items-center gap-2'>
              <BookOpenText className='h-5 w-5' />
              Documentation is being organized
            </CardTitle>
            <CardDescription>
              This workspace will host in-app guides and troubleshooting playbooks.
            </CardDescription>
          </CardHeader>
          <CardContent className='text-sm text-muted-foreground'>
            Start with Diagnostics and API docs while Help Center content is completed.
          </CardContent>
        </Card>
      </Main>
    </>
  )
}
