export const normaliseCabinetURL = (rawURL) => {
  const url = new URL(rawURL)
  if (url.protocol !== 'http:' || !['127.0.0.1', 'localhost', '[::1]'].includes(url.hostname) || url.username || url.password) {
    throw new Error('cabinet_url_must_be_loopback_http')
  }
  url.pathname = '/'
  url.search = ''
  url.hash = ''
  return url.href
}
