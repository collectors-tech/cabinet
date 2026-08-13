type ActivationFetch = (
  input: RequestInfo | URL,
  init?: RequestInit
) => Promise<Response>

type ActivationWait = (milliseconds: number) => Promise<void>

const maxActivationAttempts = 2
const defaultRetryAfterMs = 1000
const maximumRetryAfterMs = 2000

function waitFor(milliseconds: number) {
  return new Promise<void>((resolve) => window.setTimeout(resolve, milliseconds))
}

async function isRetryableActivationFailure(response: Response) {
  if (response.status !== 503) return false
  try {
    const payload = (await response.clone().json()) as {
      error?: string
      retryable?: boolean
    }
    return (
      payload.error === 'profile_activation_unavailable' &&
      payload.retryable === true
    )
  } catch {
    return false
  }
}

function activationRetryDelay(response: Response) {
  const seconds = Number(response.headers.get('Retry-After'))
  if (!Number.isFinite(seconds) || seconds <= 0) return defaultRetryAfterMs
  return Math.min(seconds * 1000, maximumRetryAfterMs)
}

export async function activateProfile(
  profileID: string,
  request: ActivationFetch = fetch,
  wait: ActivationWait = waitFor
) {
  let response: Response | undefined
  for (let attempt = 0; attempt < maxActivationAttempts; attempt += 1) {
    response = await request('/api/profiles/active', {
      method: 'PUT',
      credentials: 'same-origin',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ profile_id: profileID }),
    })
    if (response.ok) return response
    if (
      attempt + 1 >= maxActivationAttempts ||
      !(await isRetryableActivationFailure(response))
    ) {
      return response
    }
    const retryAfterMs = activationRetryDelay(response)
    await wait(retryAfterMs)
  }
  return response as Response
}
