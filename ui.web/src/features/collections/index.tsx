import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { LanguageSwitch } from '@/components/language-switch'
import { ConfigDrawer } from '@/components/config-drawer'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { collectionKey, useWorkspaceCollections } from '@/features/collections/use-workspace-collections'
import { useState } from 'react'

export function Collections() {
  const {
    workspaceCollections,
    activeWorkspaceCollection,
    setActiveWorkspaceCollection,
    addCollection,
  } = useWorkspaceCollections()
  const [newCollectionName, setNewCollectionName] = useState('')

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

      <Main className='space-y-4'>
        <div className='space-y-1'>
          <h1 className='text-2xl font-bold tracking-tight'>Collections</h1>
          <p className='text-muted-foreground'>
            Manage workspace collections outside of sidebar chrome.
          </p>
        </div>
        <Card data-testid='collections-section'>
          <CardHeader>
            <CardTitle>Workspace Collections</CardTitle>
            <CardDescription>
              Create, list, and switch active collection context.
            </CardDescription>
          </CardHeader>
          <CardContent className='space-y-3'>
            <div className='flex gap-2'>
              <Input
                data-testid='collections-new-input'
                placeholder='Collection name'
                value={newCollectionName}
                onChange={(event) => setNewCollectionName(event.target.value)}
              />
              <Button
                data-testid='collections-new-save'
                onClick={() => {
                  addCollection(newCollectionName)
                  setNewCollectionName('')
                }}
              >
                New Collection
              </Button>
            </div>
            <div className='grid grid-cols-1 gap-2 md:grid-cols-2 lg:grid-cols-3'>
              {workspaceCollections.map((name) => {
                const key = collectionKey(name)
                const isActive = activeWorkspaceCollection === name
                return (
                  <button
                    key={name}
                    type='button'
                    className='rounded-md border px-3 py-2 text-left hover:bg-muted'
                    data-testid={`collections-item-${key}`}
                    data-state={isActive ? 'active' : 'inactive'}
                    onClick={() => setActiveWorkspaceCollection(name)}
                  >
                    {name}
                  </button>
                )
              })}
            </div>
          </CardContent>
        </Card>
      </Main>
    </>
  )
}
