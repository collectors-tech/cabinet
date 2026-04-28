import { useEffect, useMemo, useState } from 'react'
import { Tags, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { ContentSection } from '../components/content-section'
import { ProfileContextBlocked } from '../components/profile-context-blocked'
import { useProfileSettings } from '../use-profile-settings'
import {
  defaultInventoryCategoryOptions,
  inventoryCategoryOptionsSettingsKey,
  normalizeCategoryName,
  normalizeCategoryOptions,
  parseCategoryOptions,
  serializeCategoryOptions,
} from '@/features/inventory/category-options'

export function SettingsCategories() {
  const {
    settings,
    loading,
    error,
    profileContextMissing,
    saving,
    reload,
    saveSettings,
  } = useProfileSettings()
  const [categories, setCategories] = useState<string[]>(
    defaultInventoryCategoryOptions
  )
  const [newCategory, setNewCategory] = useState('')
  const [status, setStatus] = useState<string | null>(null)

  useEffect(() => {
    setCategories(
      parseCategoryOptions(settings[inventoryCategoryOptionsSettingsKey])
    )
  }, [settings])

  const canSave = useMemo(
    () => !loading && !saving && !profileContextMissing,
    [loading, profileContextMissing, saving]
  )

  const addCategory = () => {
    const next = normalizeCategoryName(newCategory)
    if (next === '') {
      return
    }
    setCategories((current) => normalizeCategoryOptions([...current, next]))
    setNewCategory('')
    setStatus(null)
  }

  const removeCategory = (category: string) => {
    setCategories((current) =>
      normalizeCategoryOptions(current.filter((value) => value !== category))
    )
    setStatus(null)
  }

  const saveCategories = async () => {
    setStatus(null)
    const nextSettings = {
      ...settings,
      [inventoryCategoryOptionsSettingsKey]: serializeCategoryOptions(
        categories
      ),
    }
    await saveSettings(nextSettings)
    setStatus('Saved categories.')
  }

  return (
    <ContentSection
      title='Categories'
      desc='Manage reusable inventory category values used by item forms and filters.'
    >
      <div className='space-y-4' data-testid='settings-categories-page'>
        {profileContextMissing ? (
          <ProfileContextBlocked
            error={error ?? 'Active profile is unavailable.'}
            onRetry={reload}
          />
        ) : null}

        <div className='rounded-lg border bg-card p-4'>
          <div className='flex items-start gap-3'>
            <div className='rounded-md border bg-muted p-2'>
              <Tags className='size-4' aria-hidden='true' />
            </div>
            <div>
              <h4 className='font-medium'>Reusable category list</h4>
              <p className='text-sm text-muted-foreground'>
                These values appear in Inventory category multi-selects. Items
                can still keep existing free-text category values.
              </p>
            </div>
          </div>

          <div className='mt-4 flex gap-2'>
            <Input
              data-testid='settings-categories-new'
              placeholder='Add category'
              value={newCategory}
              disabled={!canSave}
              onChange={(event) => setNewCategory(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === 'Enter') {
                  event.preventDefault()
                  addCategory()
                }
              }}
            />
            <Button
              type='button'
              variant='outline'
              data-testid='settings-categories-add'
              disabled={!canSave || normalizeCategoryName(newCategory) === ''}
              onClick={addCategory}
            >
              Add
            </Button>
          </div>

          <div
            className='mt-4 flex flex-wrap gap-2'
            data-testid='settings-categories-list'
          >
            {categories.map((category) => (
              <Badge
                key={category}
                variant='outline'
                className='gap-2 px-3 py-1.5'
              >
                {category}
                <button
                  type='button'
                  className='rounded-sm text-muted-foreground hover:text-foreground'
                  data-testid={`settings-category-remove-${category}`}
                  aria-label={`Remove ${category}`}
                  disabled={!canSave}
                  onClick={() => removeCategory(category)}
                >
                  <Trash2 className='size-3.5' aria-hidden='true' />
                </button>
              </Badge>
            ))}
          </div>
        </div>

        {status ? (
          <p
            className='rounded-md border border-emerald-500/40 bg-emerald-500/10 p-3 text-sm'
            data-testid='settings-categories-status'
          >
            {status}
          </p>
        ) : null}

        <div className='flex justify-end'>
          <Button
            type='button'
            data-testid='settings-categories-save'
            disabled={!canSave}
            onClick={() => void saveCategories()}
          >
            Save categories
          </Button>
        </div>
      </div>
    </ContentSection>
  )
}
