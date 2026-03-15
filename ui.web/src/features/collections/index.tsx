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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
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
  const [createPanelOpen, setCreatePanelOpen] = useState(false)
  const [createValidationMessage, setCreateValidationMessage] = useState('')

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
            <div className='flex flex-wrap items-center justify-between gap-2'>
              <CardTitle>Workspace Collections</CardTitle>
              <div className='flex items-center gap-2'>
                <Button
                  type='button'
                  data-testid='collections-new-action'
                  onClick={() => {
                    setCreateValidationMessage('')
                    setCreatePanelOpen(true)
                  }}
                >
                  New
                </Button>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      type='button'
                      variant='outline'
                      data-testid='collections-create-menu-trigger'
                    >
                      Create
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align='end'>
                    <DropdownMenuItem
                      data-testid='collections-create-menu-new'
                      onClick={() => {
                        setCreateValidationMessage('')
                        setCreatePanelOpen(true)
                      }}
                    >
                      New Collection
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      data-testid='collections-create-menu-starter'
                      onClick={() => {
                        addCollection('New Arrivals')
                        addCollection('Recently Graded')
                      }}
                    >
                      Add Starter Collections
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </div>
            <CardDescription>
              Create, list, and switch active collection context.
            </CardDescription>
          </CardHeader>
          <CardContent className='space-y-3'>
            {createPanelOpen ? (
              <div
                className='rounded-md border bg-muted/20 p-3'
                data-testid='collections-create-panel'
              >
                <div className='flex flex-wrap gap-2'>
                  <Input
                    data-testid='collections-new-input'
                    placeholder='Collection name'
                    aria-invalid={createValidationMessage ? 'true' : 'false'}
                    value={newCollectionName}
                    onChange={(event) => {
                      setNewCollectionName(event.target.value)
                      if (createValidationMessage) {
                        setCreateValidationMessage('')
                      }
                    }}
                  />
                  <Button
                    data-testid='collections-new-save'
                    onClick={() => {
                      const created = addCollection(newCollectionName)
                      if (!created) {
                        setCreateValidationMessage('Collection name is required.')
                        return
                      }
                      setCreateValidationMessage('')
                      setNewCollectionName('')
                      setCreatePanelOpen(false)
                    }}
                  >
                    Save
                  </Button>
                  <Button
                    variant='outline'
                    data-testid='collections-new-cancel'
                    onClick={() => {
                      setCreatePanelOpen(false)
                      setNewCollectionName('')
                      setCreateValidationMessage('')
                    }}
                  >
                    Cancel
                  </Button>
                </div>
                {createValidationMessage ? (
                  <p
                    className='mt-2 text-sm text-destructive'
                    data-testid='collections-new-validation'
                    role='alert'
                  >
                    {createValidationMessage}
                  </p>
                ) : null}
              </div>
            ) : null}
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
