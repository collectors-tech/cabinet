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

const boundedValue = (value) => {
  if (typeof value === 'string') return value.slice(0, 8192)
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'boolean' || value === null) return value
  if (Array.isArray(value) && value.length <= 100 && value.every((item) => typeof item === 'string')) {
    return value.map((item) => item.slice(0, 2048))
  }
  throw new Error('capture_field_value_invalid')
}

export const normaliseCapture = (module, observation, rawURL, profileID) => {
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
    profile_id: profileID,
    module_id: module.id,
    url: sourceURL(module, rawURL),
    payload_type: schema.payload_type,
    captured_at: new Date().toISOString(),
    passive: true,
    attempted_write: false,
    confidence_score: confidence,
    data,
  }
}
