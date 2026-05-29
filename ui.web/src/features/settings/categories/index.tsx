import { useEffect, useMemo, useState } from 'react'
import { Tags, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
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
import {
  defaultInventoryItemTypeConditionScales,
  inventoryItemTypeConditionScalesSettingsKey,
  normalizeDisplayOption,
  normalizeDisplayOptions,
  normalizeItemTypeConditionScales,
  parseItemTypeConditionScales,
  serializeItemTypeConditionScales,
  type InventoryItemTypeConditionScale,
} from '@/features/inventory/item-type-condition-scales'
import {
  defaultInventoryPackagingGrades,
  inventoryPackagingGradesSettingsKey,
  normalizePackagingGradeName,
  parsePackagingGradeOptions,
  serializePackagingGradeOptions,
} from '@/features/inventory/packaging-grade-options'

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
  const [itemTypeScales, setItemTypeScales] = useState<
    InventoryItemTypeConditionScale[]
  >(defaultInventoryItemTypeConditionScales)
  const [packagingGrades, setPackagingGrades] = useState<string[]>(
    defaultInventoryPackagingGrades
  )
  const [newCategory, setNewCategory] = useState('')
  const [newItemType, setNewItemType] = useState('')
  const [newPackagingGrade, setNewPackagingGrade] = useState('')
  const [status, setStatus] = useState<string | null>(null)

  useEffect(() => {
    setCategories(
      parseCategoryOptions(settings[inventoryCategoryOptionsSettingsKey])
    )
    setItemTypeScales(
      parseItemTypeConditionScales(
        settings[inventoryItemTypeConditionScalesSettingsKey]
      )
    )
    setPackagingGrades(
      parsePackagingGradeOptions(settings[inventoryPackagingGradesSettingsKey])
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
      [inventoryItemTypeConditionScalesSettingsKey]:
        serializeItemTypeConditionScales(itemTypeScales),
      [inventoryPackagingGradesSettingsKey]:
        serializePackagingGradeOptions(packagingGrades),
    }
    await saveSettings(nextSettings)
    setStatus('Saved categories, packaging grades, and item type condition scales.')
  }

  const addItemType = () => {
    const itemType = normalizeDisplayOption(newItemType)
    if (itemType === '') {
      return
    }
    setItemTypeScales((current) =>
      normalizeItemTypeConditionScales([
        ...current,
        { item_type: itemType, conditions: ['New', 'Used'] },
      ])
    )
    setNewItemType('')
    setStatus(null)
  }

  const updateItemTypeConditions = (itemType: string, value: string) => {
    const conditions = normalizeDisplayOptions(value.split(/\r?\n/))
    setItemTypeScales((current) =>
      normalizeItemTypeConditionScales(
        current.map((scale) =>
          scale.item_type === itemType ? { ...scale, conditions } : scale
        )
      )
    )
    setStatus(null)
  }

  const removeItemType = (itemType: string) => {
    setItemTypeScales((current) =>
      normalizeItemTypeConditionScales(
        current.filter((scale) => scale.item_type !== itemType)
      )
    )
    setStatus(null)
  }

  const addPackagingGrade = () => {
    const packagingGrade = normalizePackagingGradeName(newPackagingGrade)
    if (packagingGrade === '') {
      return
    }
    setPackagingGrades((current) =>
      normalizeDisplayOptions([...current, packagingGrade])
    )
    setNewPackagingGrade('')
    setStatus(null)
  }

  const removePackagingGrade = (packagingGrade: string) => {
    setPackagingGrades((current) =>
      normalizeDisplayOptions(
        current.filter((value) => value !== packagingGrade)
      )
    )
    setStatus(null)
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

        <div className='rounded-lg border bg-card p-4'>
          <div className='flex items-start gap-3'>
            <div className='rounded-md border bg-muted p-2'>
              <Tags className='size-4' aria-hidden='true' />
            </div>
            <div>
              <h4 className='font-medium'>Packaging grades</h4>
              <p className='text-sm text-muted-foreground'>
                Packaging grades are reusable grading values for boxed,
                opened, complete, and loose item states.
              </p>
            </div>
          </div>

          <div className='mt-4 flex gap-2'>
            <Input
              data-testid='settings-packaging-grade-new'
              placeholder='Add packaging grade'
              value={newPackagingGrade}
              disabled={!canSave}
              onChange={(event) => setNewPackagingGrade(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === 'Enter') {
                  event.preventDefault()
                  addPackagingGrade()
                }
              }}
            />
            <Button
              type='button'
              variant='outline'
              data-testid='settings-packaging-grade-add'
              disabled={
                !canSave ||
                normalizePackagingGradeName(newPackagingGrade) === ''
              }
              onClick={addPackagingGrade}
            >
              Add
            </Button>
          </div>

          <div
            className='mt-4 flex flex-wrap gap-2'
            data-testid='settings-packaging-grades-list'
          >
            {packagingGrades.map((packagingGrade) => (
              <Badge
                key={packagingGrade}
                variant='outline'
                className='gap-2 px-3 py-1.5'
              >
                {packagingGrade}
                <button
                  type='button'
                  className='rounded-sm text-muted-foreground hover:text-foreground'
                  data-testid={`settings-packaging-grade-remove-${packagingGrade}`}
                  aria-label={`Remove ${packagingGrade}`}
                  disabled={!canSave}
                  onClick={() => removePackagingGrade(packagingGrade)}
                >
                  <Trash2 className='size-3.5' aria-hidden='true' />
                </button>
              </Badge>
            ))}
          </div>
        </div>

        <div className='rounded-lg border bg-card p-4'>
          <div className='flex items-start gap-3'>
            <div className='rounded-md border bg-muted p-2'>
              <Tags className='size-4' aria-hidden='true' />
            </div>
            <div>
              <h4 className='font-medium'>Item type condition scales</h4>
              <p className='text-sm text-muted-foreground'>
                Item Type is single-select and controls the condition choices
                shown in Inventory item forms.
              </p>
            </div>
          </div>

          <div className='mt-4 flex gap-2'>
            <Input
              data-testid='settings-item-type-new'
              placeholder='Add item type'
              value={newItemType}
              disabled={!canSave}
              onChange={(event) => setNewItemType(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === 'Enter') {
                  event.preventDefault()
                  addItemType()
                }
              }}
            />
            <Button
              type='button'
              variant='outline'
              data-testid='settings-item-type-add'
              disabled={!canSave || normalizeDisplayOption(newItemType) === ''}
              onClick={addItemType}
            >
              Add
            </Button>
          </div>

          <div
            className='mt-4 grid gap-3 lg:grid-cols-2'
            data-testid='settings-item-type-scales-list'
          >
            {itemTypeScales.map((scale) => (
              <div key={scale.item_type} className='rounded-md border p-3'>
                <div className='mb-2 flex items-center justify-between gap-2'>
                  <h5 className='font-medium'>{scale.item_type}</h5>
                  <Button
                    type='button'
                    size='icon'
                    variant='ghost'
                    data-testid={`settings-item-type-remove-${scale.item_type}`}
                    aria-label={`Remove ${scale.item_type}`}
                    disabled={!canSave}
                    onClick={() => removeItemType(scale.item_type)}
                  >
                    <Trash2 className='size-4' aria-hidden='true' />
                  </Button>
                </div>
                <Textarea
                  data-testid={`settings-item-type-conditions-${scale.item_type}`}
                  value={scale.conditions.join('\n')}
                  disabled={!canSave}
                  rows={Math.min(8, Math.max(3, scale.conditions.length))}
                  onChange={(event) =>
                    updateItemTypeConditions(
                      scale.item_type,
                      event.target.value
                    )
                  }
                />
              </div>
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
            Save taxonomy settings
          </Button>
        </div>
      </div>
    </ContentSection>
  )
}
