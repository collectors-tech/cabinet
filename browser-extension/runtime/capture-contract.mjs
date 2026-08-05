const forbiddenField = /(secret|token|password|cookie|authorization|credential|raw[_-]?(html|page|dom))/i

const sourceURL = (module, rawURL) => {
  const url = new URL(rawURL)
  if (url.protocol !== 'https:' || url.username || url.password || !module.browser.url_patterns.some((pattern) => url.href.startsWith(pattern.slice(0, -1)))) {
    throw new Error('capture_url_outside_module')
  }
  url.search = ''
  url.hash = ''
  return url.href
}

const boundedValue = (value, depth = 0, counter = { count: 0 }) => {
	if (depth > 8 || ++counter.count > 1000) throw new Error('capture_field_value_invalid')
  if (typeof value === 'string') return value.slice(0, 8192)
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'boolean' || value === null) return value
  if (Array.isArray(value) && value.length <= 200) {
    return value.map((item) => boundedValue(item, depth + 1, counter))
  }
  if (value && typeof value === 'object' && !Array.isArray(value) && Object.keys(value).length <= 100) {
    const result = {}
    for (const key of Object.keys(value).sort()) {
      if (forbiddenField.test(key)) throw new Error(`capture_field_rejected:${key}`)
      result[key] = boundedValue(value[key], depth + 1, counter)
    }
    return result
  }
  throw new Error('capture_field_value_invalid')
}

const canonicalJSON = (value) => {
  if (Array.isArray(value)) return `[${value.map(canonicalJSON).join(',')}]`
  if (value && typeof value === 'object') {
    return `{${Object.keys(value).sort().map((key) => `${JSON.stringify(key)}:${canonicalJSON(value[key])}`).join(',')}}`
  }
  return JSON.stringify(value)
}

const sha256 = async (value) => {
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(value))
  return `sha256:${[...new Uint8Array(digest)].map((byte) => byte.toString(16).padStart(2, '0')).join('')}`
}

export const normaliseCapture = async (module, observation, rawURL, profileID, idempotencyKey = crypto.randomUUID()) => {
  const schema = module.capture_schemas.find((item) => item.payload_type === observation?.payload_type)
  if (!schema) throw new Error('capture_payload_type_unsupported')
  if (!observation.data || typeof observation.data !== 'object' || Array.isArray(observation.data)) {
    throw new Error('capture_data_required')
  }
  const allowed = new Set([...schema.fields, ...schema.media_fields])
  const data = {}
  for (const [key, value] of Object.entries(observation.data)) {
    if (forbiddenField.test(key) || !allowed.has(key)) throw new Error(`capture_field_rejected:${key}`)
    data[key] = boundedValue(value)
  }
  const encoded = JSON.stringify(data)
  if (encoded.length > 256 * 1024) throw new Error('capture_payload_too_large')
  const confidence = Number(observation.confidence_score)
  if (!Number.isFinite(confidence) || confidence < 0 || confidence > 1) throw new Error('capture_confidence_invalid')
  return {
	protocol_version: '1',
    profile_id: profileID,
    module_id: module.id,
	module_version: module.module_version,
	schema_version: module.fixture_version,
	integration_instance_id: module.integration_instance_id,
	provider_id: module.provider_id,
    url: sourceURL(module, rawURL),
    payload_type: schema.payload_type,
    captured_at: new Date().toISOString(),
	page_complete: observation.page_complete === true,
    passive: true,
    attempted_write: false,
    confidence_score: confidence,
	redaction_summary: [...module.redaction_rules],
	payload_hash: await sha256(canonicalJSON(data)),
	idempotency_key: idempotencyKey,
    data,
  }
}
