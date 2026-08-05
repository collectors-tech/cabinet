import assert from 'node:assert/strict'
import test from 'node:test'

import { DurableQueue } from '../runtime/durable-queue.mjs'
import { RetryGovernor } from '../runtime/retry-governor.mjs'

class MemoryStore {
  constructor(snapshot) { this.snapshot = snapshot }
  async get() { return this.snapshot }
  async set(_key, value) { this.snapshot = structuredClone(value) }
}

test('queue resumes idempotent jobs after service-worker or Cabinet restart', async () => {
  const store = new MemoryStore()
  const first = new DurableQueue({ storage: store, now: () => 1000 })
  await first.enqueue({ id: 'job-1', module_id: 'example', kind: 'capture', payload: { item: 1 } })
  await first.enqueue({ id: 'job-1', module_id: 'example', kind: 'capture', payload: { item: 2 } })

  const restarted = new DurableQueue({ storage: store, now: () => 2000 })
  assert.equal((await restarted.pending()).length, 1)
  const claimed = await restarted.claim()
  assert.equal(claimed.id, 'job-1')
  await restarted.fail('job-1', 'cabinet_offline', 5000)
  assert.equal((await restarted.pending())[0].attempts, 1)

  const recovered = new DurableQueue({ storage: store, now: () => 7000 })
  assert.equal((await recovered.claim()).id, 'job-1')
  await recovered.complete('job-1')
  assert.deepEqual(await recovered.pending(), [])
})

test('queue recovers a job when the MV3 worker stops after claiming it', async () => {
  const store = new MemoryStore()
  const first = new DurableQueue({ storage: store, now: () => 1000, leaseMs: 30_000 })
  await first.enqueue({ id: 'job-1', module_id: 'example', kind: 'capture', payload: {} })
  assert.equal((await first.claim()).state, 'running')

  const restarted = new DurableQueue({ storage: store, now: () => 31_001, leaseMs: 30_000 })
  const recovered = await restarted.claim()
  assert.equal(recovered.id, 'job-1')
  assert.equal(recovered.attempts, 1)
})

test('governor applies bounded exponential backoff and opens a circuit', () => {
  const governor = new RetryGovernor({ baseDelayMs: 1000, maxDelayMs: 8000, failureLimit: 3 })
  assert.deepEqual(governor.failure('example', 100), { delay_ms: 1000, circuit_open: false })
  assert.deepEqual(governor.failure('example', 200), { delay_ms: 2000, circuit_open: false })
  assert.deepEqual(governor.failure('example', 300), { delay_ms: 4000, circuit_open: true })
  assert.equal(governor.canRun('example', 4299), false)
  assert.equal(governor.canRun('example', 4300), true)
  governor.success('example')
  assert.equal(governor.canRun('example', 4300), true)
})

test('governor state can be restored after MV3 service-worker suspension', () => {
  const first = new RetryGovernor({ baseDelayMs: 1000, maxDelayMs: 8000, failureLimit: 2 })
  first.failure('example', 100)
  first.failure('example', 200)

  const restored = new RetryGovernor({
    baseDelayMs: 1000,
    maxDelayMs: 8000,
    failureLimit: 2,
    snapshot: first.snapshot(),
  })
  assert.equal(restored.canRun('example', 2199), false)
  assert.equal(restored.canRun('example', 2200), true)
})
