import { useCallback, useEffect, useState } from 'react'

type ActiveProfileResponse = {
  id?: string
}

type ProfileSettingsResponse = {
  settings?: Record<string, string>
}

export function useProfileSettings() {
  const [activeProfileId, setActiveProfileId] = useState('')
  const [settings, setSettings] = useState<Record<string, string>>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const activeResp = await fetch('/api/profiles/active')
      if (!activeResp.ok) {
        throw new Error(`active_profile_${activeResp.status}`)
      }
      const active = (await activeResp.json()) as ActiveProfileResponse
      const profileID = active.id?.trim() ?? ''
      if (!profileID) {
        throw new Error('active_profile_missing')
      }
      setActiveProfileId(profileID)

      const settingsResp = await fetch(`/api/profiles/${profileID}/settings`)
      if (!settingsResp.ok) {
        throw new Error(`profile_settings_${settingsResp.status}`)
      }
      const payload = (await settingsResp.json()) as ProfileSettingsResponse
      setSettings(payload.settings ?? {})
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : 'Failed to load profile settings.'
      )
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const saveSettings = useCallback(
    async (next: Record<string, string>) => {
      if (!activeProfileId) {
        throw new Error('active_profile_missing')
      }
      setSaving(true)
      try {
        const response = await fetch(`/api/profiles/${activeProfileId}/settings`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ settings: next }),
        })
        if (!response.ok) {
          throw new Error(`profile_settings_save_${response.status}`)
        }
        const payload = (await response.json()) as ProfileSettingsResponse
        setSettings(payload.settings ?? {})
        return payload.settings ?? {}
      } finally {
        setSaving(false)
      }
    },
    [activeProfileId]
  )

  return {
    activeProfileId,
    settings,
    loading,
    error,
    profileContextMissing:
      error === 'active_profile_missing' || error?.startsWith('active_profile_'),
    saving,
    reload: load,
    saveSettings,
  }
}
