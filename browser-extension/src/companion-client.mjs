const credentialKey = 'cabinet.companion.credential.v1'
const pendingPairingKey = 'cabinet.companion.pending-pairing.v1'

export class CompanionProtocolError extends Error {
  constructor(status, code) {
    super(code)
    this.name = 'CompanionProtocolError'
    this.status = status
    this.code = code
  }
}

export class CompanionClient {
  constructor({ baseURL, deviceID, fetchImpl = fetch, storage }) {
    this.baseURL = new URL(baseURL)
    this.deviceID = deviceID
    this.fetchImpl = fetchImpl
    this.storage = storage
  }

  async startPairing(deviceName, capabilities) {
    const receipt = await this.#request('/api/companion/pairing/requests', {
      method: 'POST',
      body: {
        device_id: this.deviceID,
        device_name: deviceName,
        protocol_version: '1',
        capabilities,
      },
    })
    await this.storage.set(pendingPairingKey, receipt)
    return receipt
  }

  async exchangePairing() {
    const pending = await this.storage.get(pendingPairingKey)
    if (!pending) {
      throw new CompanionProtocolError(400, 'companion_pairing_not_started')
    }
    const response = await this.#request('/api/companion/pairing/exchanges', {
      method: 'POST',
      body: {
        request_id: pending.request_id,
        exchange_secret: pending.exchange_secret,
        device_id: this.deviceID,
        protocol_version: '1',
      },
    })
    await this.storage.set(credentialKey, response.credential)
    await this.storage.delete(pendingPairingKey)
    return response.session
  }

  async reconnect() {
    return this.#request('/api/companion/session', {
      method: 'GET',
      authenticated: true,
    })
  }

  async rotate() {
    const response = await this.#request('/api/companion/session/rotate', {
      method: 'POST',
      authenticated: true,
    })
    await this.storage.set(credentialKey, response.credential)
    return response.session
  }

  async revoke() {
    await this.#request('/api/companion/session', {
      method: 'DELETE',
      authenticated: true,
    })
    await this.storage.delete(credentialKey)
  }

  async modules() {
    return this.#request('/api/companion/modules', {
      method: 'GET',
      authenticated: true,
    })
  }

  async captureInbox() {
    return this.#request('/api/companion/payloads?limit=200', {
      method: 'GET',
      authenticated: true,
    })
  }

  async submitCapture(payload) {
    return this.#request('/api/companion/payloads', {
      method: 'POST',
      authenticated: true,
      body: payload,
	  headers: { 'X-Cabinet-Idempotency-Key': payload.idempotency_key },
    })
  }

  async submitMedia({ bytes, profileID, captureID, fieldName, filename, mimeType, sha256, idempotencyKey }) {
    return this.#request('/api/companion/media-submissions', {
      method: 'POST',
      authenticated: true,
      rawBody: bytes,
      headers: {
        'Content-Type': mimeType,
        'X-Cabinet-Profile': profileID,
        'X-Cabinet-Idempotency-Key': idempotencyKey,
        'X-Cabinet-Capture-ID': captureID,
        'X-Cabinet-Media-Field': fieldName,
        'X-Cabinet-Media-Filename': filename,
        'X-Cabinet-Media-SHA256': sha256,
      },
    })
  }

  async #request(path, { method, body, rawBody, authenticated = false, headers: additionalHeaders = {} }) {
    const headers = { 'X-Cabinet-Companion-Device': this.deviceID, ...additionalHeaders }
    if (body !== undefined) {
      headers['Content-Type'] = 'application/json'
    }
    if (authenticated) {
      const credential = await this.storage.get(credentialKey)
      if (!credential) {
        throw new CompanionProtocolError(401, 'companion_auth_required')
      }
      headers.Authorization = `Bearer ${credential}`
    }
    const response = await this.fetchImpl(new URL(path, this.baseURL), {
      method,
      headers,
      body: rawBody ?? (body === undefined ? undefined : JSON.stringify(body)),
    })
    const payload = await response.json()
    if (!response.ok) {
      throw new CompanionProtocolError(
        response.status,
        payload.error ?? 'companion_request_failed'
      )
    }
    return payload
  }
}

export const companionStorageKeys = Object.freeze({
  credential: credentialKey,
  pendingPairing: pendingPairingKey,
})
