import { useAuthStore } from '@/stores/auth-store'
import { activateProfile } from '@/lib/profile-activation'

const protectedExactPaths = new Set([
  '/api/chat/messages',
  '/api/chat/actions/preview',
  '/api/chat/actions/apply',
  '/api/chat/actions/cancel',
  '/api/chat/workflow-runs',
])

function isProtectedCabinetPath(pathname: string) {
  return (
    protectedExactPaths.has(pathname) ||
    pathname.startsWith('/api/chat/workflow-runs/') ||
    pathname.startsWith('/api/agent/')
  )
}

function resolveProtectedURL(input: string | URL) {
  const url = new URL(String(input), window.location.origin)
  if (url.origin !== window.location.origin || !isProtectedCabinetPath(url.pathname)) {
    throw new Error('cabinet_session_untrusted_request')
  }
  return url
}

export async function bootstrapLocalServerSession(profileID: string) {
  const normalizedProfileID = profileID.trim()
  if (!normalizedProfileID) {
    throw new Error('cabinet_session_profile_required')
  }

  const response = await fetch('/api/auth/local/session', {
    method: 'POST',
    credentials: 'same-origin',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ profile_id: normalizedProfileID }),
  })
  if (!response.ok) {
    throw new Error(`cabinet_session_bootstrap_${response.status}`)
  }
  const payload = (await response.json()) as { session_token?: string }
  const token = payload.session_token?.trim() || ''
  if (!token) {
    throw new Error('cabinet_session_missing_token')
  }

  useAuthStore.getState().auth.setLocalSession(normalizedProfileID, token)
}

export async function bootstrapLocalServerSessionForActiveProfile() {
  let response = await fetch('/api/profiles/active', {
    credentials: 'same-origin',
    headers: { Accept: 'application/json' },
  })
  if (response.status === 404) {
    const createResponse = await fetch('/api/profiles', {
      method: 'POST',
      credentials: 'same-origin',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ name: 'Default' }),
    })
    if (!createResponse.ok) {
      throw new Error(`cabinet_session_create_profile_${createResponse.status}`)
    }
    const created = (await createResponse.json()) as { id?: string }
    const createdProfileID = created.id?.trim() || ''
    if (!createdProfileID) {
      throw new Error('cabinet_session_created_profile_missing_id')
    }
    const activateResponse = await activateProfile(createdProfileID)
    if (!activateResponse.ok) {
      throw new Error(`cabinet_session_activate_profile_${activateResponse.status}`)
    }
    response = activateResponse
  }
  if (!response.ok) {
    throw new Error(`cabinet_session_active_profile_${response.status}`)
  }
  const payload = (await response.json()) as { id?: string }
  const profileID = payload.id?.trim() || ''
  await bootstrapLocalServerSession(profileID)
  return profileID
}

export async function cabinetProtectedFetch(
  input: string | URL,
  profileID: string,
  init: RequestInit = {}
) {
  const url = resolveProtectedURL(input)
  const session = useAuthStore.getState().auth
  const normalizedProfileID = profileID.trim()
  const hasBoundLocalSession =
    session.localSessionProfileID === normalizedProfileID &&
    Boolean(session.localSessionToken)
  if (!normalizedProfileID || (!hasBoundLocalSession && !session.remoteSession)) {
    throw new Error('cabinet_session_unavailable')
  }

  const headers = new Headers(init.headers)
  if (hasBoundLocalSession) {
    headers.set('X-Cabinet-Session', session.localSessionToken)
  }
  const response = await fetch(url.pathname + url.search, {
    ...init,
    credentials: 'same-origin',
    headers,
  })
  if (response.status === 401 || response.status === 423) {
    useAuthStore.getState().auth.clearLocalSession()
  }
  return response
}

export async function lockLocalServerSession() {
  const auth = useAuthStore.getState().auth
  const token = auth.localSessionToken
  try {
    if (token) {
      await fetch('/api/auth/session/lock', {
        method: 'POST',
        credentials: 'same-origin',
        headers: {
          Accept: 'application/json',
          'Content-Type': 'application/json',
          'X-Cabinet-Session': token,
        },
        body: '{}',
      })
    }
  } finally {
    useAuthStore.getState().auth.clearLocalSession()
  }
}
