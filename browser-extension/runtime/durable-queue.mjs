const queueKey = 'cabinet.companion.queue.v1'

export class DurableQueue {
  constructor({ storage, now = Date.now, leaseMs = 30_000 }) {
    this.storage = storage
    this.now = now
    this.leaseMs = leaseMs
  }

  async #load() {
    const jobs = (await this.storage.get(queueKey)) ?? []
    let recovered = false
    for (const job of jobs) {
      if (job.state === 'running' && Number(job.lease_until) <= this.now()) {
        job.state = 'pending'
        job.attempts += 1
        job.last_error = 'worker_interrupted'
        job.available_at = this.now()
        delete job.lease_until
        recovered = true
      }
    }
    if (recovered) await this.#save(jobs)
    return jobs
  }
  async #save(jobs) { await this.storage.set(queueKey, jobs) }

  async enqueue(input) {
    if (!input?.id || !input?.module_id || !input?.kind) throw new Error('queue_job_invalid')
    const jobs = await this.#load()
    if (jobs.some((job) => job.id === input.id)) return false
    jobs.push({ ...input, attempts: 0, available_at: this.now(), state: 'pending' })
    await this.#save(jobs)
    return true
  }

  async pending() { return this.#load() }

  async claim() {
    const jobs = await this.#load()
    const job = jobs.find((candidate) => candidate.state === 'pending' && candidate.available_at <= this.now())
    if (!job) return undefined
    job.state = 'running'
    job.lease_until = this.now() + this.leaseMs
    await this.#save(jobs)
    return structuredClone(job)
  }

  async fail(id, errorCode, retryDelayMs) {
    const jobs = await this.#load()
    const job = jobs.find((candidate) => candidate.id === id)
    if (!job) throw new Error('queue_job_not_found')
    job.state = 'pending'
    job.attempts += 1
    job.last_error = String(errorCode).slice(0, 128)
    job.available_at = this.now() + Math.max(0, Math.min(retryDelayMs, 24 * 60 * 60 * 1000))
    delete job.lease_until
    await this.#save(jobs)
  }

  async complete(id) {
    const jobs = await this.#load()
    await this.#save(jobs.filter((candidate) => candidate.id !== id))
  }
}

export const durableQueueStorageKey = queueKey
