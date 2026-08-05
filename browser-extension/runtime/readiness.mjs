const intersects = (expected, observed) => expected.some((selector) => observed.has(selector))

export const classifyReadiness = (definition = {}, evidence = []) => {
  const observed = new Set(Array.isArray(evidence) ? evidence.slice(0, 60) : [])
  if (intersects(definition.challenge ?? [], observed)) {
    return { state: 'action_required', guidance: 'Complete the site check in the open browser tab.' }
  }
  if (intersects(definition.logged_out ?? [], observed)) {
    return { state: 'logged_out', guidance: 'Sign in on the provider site, then check again.' }
  }
  if (intersects(definition.ready ?? [], observed)) {
    return { state: 'ready', guidance: 'The current browser session is ready.' }
  }
  return { state: 'unsupported', guidance: 'No supported signed-in page state was detected.' }
}
