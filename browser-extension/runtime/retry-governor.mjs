export class RetryGovernor {
  constructor({ baseDelayMs = 1000, maxDelayMs = 60_000, failureLimit = 5, snapshot = {} } = {}) {
    this.baseDelayMs = baseDelayMs
    this.maxDelayMs = maxDelayMs
    this.failureLimit = failureLimit
    this.providers = new Map(Object.entries(snapshot).filter(([, value]) =>
      Number.isInteger(value?.failures) && Number.isFinite(value?.blocked_until)
    ))
  }

  failure(providerID, now = Date.now()) {
    const failures = (this.providers.get(providerID)?.failures ?? 0) + 1
    const delay = Math.min(this.maxDelayMs, this.baseDelayMs * (2 ** (failures - 1)))
    const circuitOpen = failures >= this.failureLimit
    this.providers.set(providerID, { failures, blocked_until: circuitOpen ? now + delay : now })
    return { delay_ms: delay, circuit_open: circuitOpen }
  }

  success(providerID) { this.providers.delete(providerID) }

  canRun(providerID, now = Date.now()) {
    return now >= (this.providers.get(providerID)?.blocked_until ?? 0)
  }

  snapshot() { return Object.fromEntries(this.providers) }
}
