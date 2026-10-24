import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { mount } from '@vue/test-utils'
import Update, { type Release, type UpdateState, type UpdateStatus } from './update'

const requests = vi.hoisted(() => ({
  status: vi.fn(), check: vi.fn(), install: vi.fn(), legacyInstall: vi.fn(), clearAuthenticated: vi.fn(),
}))
vi.mock('@/plugins/api', () => ({ default: { get: requests.status, post: requests.install } }))
vi.mock('@/plugins/httputil', () => ({ default: { get: requests.check, post: requests.legacyInstall } }))
vi.mock('@/plugins/auth', () => ({ clearAuthenticated: requests.clearAuthenticated }))
vi.mock('@/locales', () => ({ i18n: { global: { t: (key: string) => key } } }))

const storageKey = 'm1ss-ui-update-job'
const job = 'job-current'
const targetVersion = '1.6.9'
const pending = { job, targetVersion }
const message = <T>(obj: T) => ({ success: true, msg: '', obj })
const response = (obj: UpdateStatus) => ({ data: message(obj) })
const status = (phase: string, currentJob = job): UpdateStatus => ({
  currentVersion: phase === 'succeeded' ? targetVersion : '1.6.8',
  supported: true,
  state: phase === 'idle' ? { phase } : { phase, job: currentJob, targetVersion, previousVersion: '1.6.8' },
})
const release = (): Release => ({
  ...status('idle'), latestVersion: targetVersion, available: true,
  url: `https://example.test/releases/v${targetVersion}`,
})

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason: unknown) => void
  const promise = new Promise<T>((done, fail) => { resolve = done; reject = fail })
  return { promise, resolve, reject }
}

let stores: ReturnType<typeof Update>[]
let reload: ReturnType<typeof vi.spyOn>
const useStore = () => {
  const store = Update()
  if (!stores.includes(store)) stores.push(store)
  return store
}
const flushRequests = () => vi.advanceTimersByTimeAsync(0)
const installResponse = (state: UpdateState) => ({ data: message(state) })
const readPending = () => JSON.parse(sessionStorage.getItem(storageKey)!)
const prepareInstall = async (initial = status('idle')) => {
  requests.status.mockResolvedValueOnce(response(initial))
  const store = useStore()
  await store.start()
  requests.check.mockResolvedValueOnce(message({ ...release(), ...initial }))
  await store.check()
  return store
}

beforeEach(() => {
  vi.useFakeTimers()
  vi.setSystemTime(new Date('2026-10-24T00:00:00Z'))
  sessionStorage.clear()
  setActivePinia(createPinia())
  stores = []
  requests.status.mockReset().mockResolvedValue(response(status('idle')))
  requests.check.mockReset().mockResolvedValue(message(release()))
  requests.install.mockReset().mockResolvedValue(installResponse(status('queued').state))
  requests.legacyInstall.mockReset()
  requests.clearAuthenticated.mockReset()
  reload = vi.spyOn(window.location, 'reload').mockImplementation(() => {})
})

afterEach(() => {
  for (const store of stores) store.stop()
  vi.clearAllTimers()
  vi.useRealTimers()
  vi.restoreAllMocks()
  sessionStorage.clear()
})

describe('update monitoring outside the settings page', () => {
  it('initializes once and leaves an idle panel without a polling timer', async () => {
    const initial = deferred<ReturnType<typeof response>>()
    requests.status.mockReturnValueOnce(initial.promise)
    const store = useStore()
    const starts = [store.start(), store.start(), store.start()]
    expect(requests.status).toHaveBeenCalledTimes(1)
    initial.resolve(response(status('idle')))
    await Promise.all(starts)
    await store.start()
    await vi.advanceTimersByTimeAsync(12000)
    expect(requests.status).toHaveBeenCalledTimes(1)
    expect(requests.status).toHaveBeenCalledWith('api/updateStatus', { timeout: 4000 })
    expect(vi.getTimerCount()).toBe(0)
    expect(store.active).toBe(false)
    expect(sessionStorage.getItem(storageKey)).toBeNull()
  })

  it('blocks check and install while the initial status request is in flight', async () => {
    const initial = deferred<ReturnType<typeof response>>()
    requests.status.mockReturnValueOnce(initial.promise)
    const store = useStore()
    // An available release must not bypass the status-request guard.
    store.release = release()
    const start = store.start()
    expect(store.refreshing).toBe(true)
    await store.check()
    await store.install()
    expect(requests.check).not.toHaveBeenCalled()
    expect(requests.install).not.toHaveBeenCalled()
    initial.resolve(response(status('idle')))
    await start
    expect(store.refreshing).toBe(false)
    await store.check()
    expect(requests.check).toHaveBeenCalledTimes(1)
  })

  it('keeps an accepted update monitored after its settings consumer unmounts', async () => {
    const store = useStore()
    await store.start()
    const SettingsConsumer = defineComponent({
      setup() {
        const update = Update()
        return () => h('p', update.phase)
      },
    })
    const settings = mount(SettingsConsumer)
    await store.check()
    requests.status.mockResolvedValueOnce(response(status('installing')))
    await store.install()
    await flushRequests()
    expect(Update()).toBe(store)
    expect(store.phase).toBe('installing')
    settings.unmount()

    requests.status.mockResolvedValue(response(status('succeeded')))
    await vi.advanceTimersByTimeAsync(3000)
    expect(store.phase).toBe('succeeded')
    expect(reload).toHaveBeenCalledTimes(1)
    expect(sessionStorage.getItem(storageKey)).toBeNull()
    expect(vi.getTimerCount()).toBe(0)
    await store.start()
    await vi.advanceTimersByTimeAsync(12000)
    expect(reload).toHaveBeenCalledTimes(1)
  })

  it('does not create a second polling chain when start is called during an update', async () => {
    requests.status.mockResolvedValue(response(status('downloading')))
    const store = useStore()
    await store.start()
    await Promise.all([store.start(), store.start()])
    expect(requests.status).toHaveBeenCalledTimes(1)
    expect(vi.getTimerCount()).toBe(1)
    await vi.advanceTimersByTimeAsync(3000)
    expect(requests.status).toHaveBeenCalledTimes(2)
    expect(vi.getTimerCount()).toBe(1)
  })

  it('persists a job first observed as active and refreshes when that job succeeds', async () => {
    requests.status.mockResolvedValueOnce(response(status('downloading')))
      .mockResolvedValue(response(status('succeeded')))
    const store = useStore()
    await store.start()
    expect(JSON.parse(sessionStorage.getItem(storageKey)!)).toEqual(pending)
    await vi.advanceTimersByTimeAsync(3000)
    expect(reload).toHaveBeenCalledTimes(1)
    expect(sessionStorage.getItem(storageKey)).toBeNull()
  })

  it('restores a persisted job in a fresh store and reloads only once', async () => {
    sessionStorage.setItem(storageKey, JSON.stringify(pending))
    requests.status.mockResolvedValueOnce(response(status('restarting')))
      .mockResolvedValue(response(status('succeeded')))
    const store = useStore()
    await store.start()
    expect(store.active).toBe(true)
    await vi.advanceTimersByTimeAsync(3000)
    expect(reload).toHaveBeenCalledTimes(1)
    expect(sessionStorage.getItem(storageKey)).toBeNull()

    // A browser reload recreates Pinia. The completed backend record remains.
    vi.clearAllTimers()
    setActivePinia(createPinia())
    await useStore().start()
    await vi.advanceTimersByTimeAsync(9000)
    expect(reload).toHaveBeenCalledTimes(1)
    expect(vi.getTimerCount()).toBe(0)
  })

  it('keeps reconnecting after starting offline with a persisted job', async () => {
    sessionStorage.setItem(storageKey, JSON.stringify(pending))
    requests.status.mockRejectedValueOnce(new Error('offline'))
      .mockResolvedValueOnce(response(status('restarting')))
      .mockResolvedValue(response(status('succeeded')))
    const store = useStore()
    await store.start()
    expect(store.disconnected).toBe(true)
    expect(store.pauseDataPolling).toBe(true)
    expect(sessionStorage.getItem(storageKey)).not.toBeNull()
    await vi.advanceTimersByTimeAsync(3000)
    expect(store.disconnected).toBe(false)
    expect(store.phase).toBe('restarting')
    await vi.advanceTimersByTimeAsync(3000)
    expect(reload).toHaveBeenCalledTimes(1)
  })

  it('does not reload or keep polling for an untracked historical success', async () => {
    requests.status.mockResolvedValue(response(status('succeeded', 'job-historical')))
    const store = useStore()
    await store.start()
    await vi.advanceTimersByTimeAsync(60000)
    expect(store.phase).toBe('succeeded')
    expect(reload).not.toHaveBeenCalled()
    expect(requests.status).toHaveBeenCalledTimes(1)
    expect(vi.getTimerCount()).toBe(0)
  })

  it('does not reload for another completed job when a different job was persisted', async () => {
    sessionStorage.setItem(storageKey, JSON.stringify(pending))
    requests.status.mockResolvedValue(response(status('succeeded', 'job-other')))
    await useStore().start()
    expect(reload).not.toHaveBeenCalled()
    expect(sessionStorage.getItem(storageKey)).toBeNull()
    expect(vi.getTimerCount()).toBe(0)
  })

  it('requires the expected version to be running before reloading a completed job', async () => {
    sessionStorage.setItem(storageKey, JSON.stringify(pending))
    requests.status.mockResolvedValue(response({ ...status('succeeded'), currentVersion: '1.6.8' }))
    await useStore().start()
    expect(reload).not.toHaveBeenCalled()
    expect(sessionStorage.getItem(storageKey)).toBeNull()
    expect(vi.getTimerCount()).toBe(0)
  })

  it.each(['failed', 'rolledBack', 'interrupted'])('stops and clears the pending job after %s', async (phase) => {
    requests.status.mockResolvedValueOnce(response(status('installing')))
      .mockResolvedValue(response(status(phase)))
    const store = useStore()
    await store.start()
    expect(store.pauseDataPolling).toBe(true)
    await vi.advanceTimersByTimeAsync(3000)
    expect(store.phase).toBe(phase)
    expect(store.active).toBe(false)
    expect(store.pauseDataPolling).toBe(false)
    expect(store.alertType).toBe('warning')
    expect(sessionStorage.getItem(storageKey)).toBeNull()
    expect(vi.getTimerCount()).toBe(0)
    await vi.advanceTimersByTimeAsync(9000)
    expect(requests.status).toHaveBeenCalledTimes(2)
    expect(reload).not.toHaveBeenCalled()
  })

  it('uses HttpUtils for checks and persists an intent before posting the installation', async () => {
    const store = useStore()
    await store.start()
    await store.check()
    expect(requests.check).toHaveBeenCalledWith('api/checkUpdate', {}, { timeout: 20000 })
    expect(store.release).toEqual(release())
    const accepted = deferred<ReturnType<typeof installResponse>>()
    requests.install.mockImplementationOnce(() => {
      expect(readPending()).toEqual({ targetVersion })
      return accepted.promise
    })
    requests.status.mockResolvedValue(response(status('queued')))
    const install = store.install()
    await store.install()
    expect(store.starting).toBe(true)
    expect(requests.install).toHaveBeenCalledTimes(1)
    expect(requests.install).toHaveBeenCalledWith('api/update', { version: targetVersion }, { timeout: 20000 })
    expect(store.confirming).toBe(true)
    expect(store.busy).toBe(true)
    expect(store.pauseDataPolling).toBe(true)
    expect(readPending()).toEqual({ targetVersion })
    expect(requests.legacyInstall).not.toHaveBeenCalled()
    accepted.resolve(installResponse(status('queued').state))
    await install
    await flushRequests()
    expect(store.starting).toBe(false)
    expect(JSON.parse(sessionStorage.getItem(storageKey)!)).toEqual(pending)
    expect(vi.getTimerCount()).toBe(1)
  })

  it('clears update state and polling when the authenticated layout stops', async () => {
    requests.status.mockResolvedValue(response(status('installing')))
    const store = useStore()
    await store.start()
    expect(vi.getTimerCount()).toBe(1)
    store.stop()
    expect(store.phase).toBe('idle')
    expect(store.status).toBeUndefined()
    expect(store.release).toBeUndefined()
    expect(store.pauseDataPolling).toBe(false)
    expect(sessionStorage.getItem(storageKey)).toBeNull()
    expect(vi.getTimerCount()).toBe(0)
    await vi.advanceTimersByTimeAsync(9000)
    expect(requests.status).toHaveBeenCalledTimes(1)
  })

  it('ignores a previous session status response even after a fresh start', async () => {
    const late = deferred<ReturnType<typeof response>>()
    requests.status.mockReturnValueOnce(late.promise)
    const store = useStore()
    const initial = store.start()
    store.stop()
    await store.start()
    expect(store.phase).toBe('idle')
    late.resolve(response(status('installing')))
    await initial
    expect(store.phase).toBe('idle')
    expect(store.pauseDataPolling).toBe(false)
    expect(sessionStorage.getItem(storageKey)).toBeNull()
    expect(vi.getTimerCount()).toBe(0)
  })

  it.each(['check', 'install'] as const)('ignores a late %s response after stop', async (action) => {
    const store = useStore()
    await store.start()
    const late = deferred<ReturnType<typeof message<Release>> | ReturnType<typeof installResponse>>()
    if (action === 'install') {
      await store.check()
      requests.install.mockReturnValueOnce(late.promise)
    } else requests.check.mockReturnValueOnce(late.promise)
    const request = store[action]()
    store.stop()
    late.resolve(action === 'check' ? message(release()) : installResponse(status('queued').state))
    await request
    await flushRequests()
    expect(store.status).toBeUndefined()
    expect(store.release).toBeUndefined()
    expect(store.checking).toBe(false)
    expect(store.starting).toBe(false)
    expect(sessionStorage.getItem(storageKey)).toBeNull()
    expect(vi.getTimerCount()).toBe(0)
    expect(reload).not.toHaveBeenCalled()
  })

  it('pauses after 25 minutes and allows monitoring to resume without reinstalling', async () => {
    sessionStorage.setItem(storageKey, JSON.stringify(pending))
    requests.status.mockRejectedValue(new Error('offline'))
    const store = useStore()
    await store.start()
    await vi.advanceTimersByTimeAsync(25 * 60 * 1000)
    expect(store.monitorExpired).toBe(true)
    expect(vi.getTimerCount()).toBe(0)
    const count = requests.status.mock.calls.length
    await vi.advanceTimersByTimeAsync(30000)
    expect(requests.status).toHaveBeenCalledTimes(count)
    expect(sessionStorage.getItem(storageKey)).not.toBeNull()
    store.resumeMonitoring()
    await flushRequests()
    expect(store.monitorExpired).toBe(false)
    expect(requests.status).toHaveBeenCalledTimes(count + 1)
    expect(vi.getTimerCount()).toBe(1)
    expect(requests.install).not.toHaveBeenCalled()
  })

  it('preserves a known job on 401 and restores monitoring after login reload', async () => {
    sessionStorage.setItem(storageKey, JSON.stringify(pending))
    requests.status.mockRejectedValueOnce({ response: { status: 401 } })
    const store = useStore()
    await store.start()
    expect(store.statusError).toBe('updater.sessionExpired')
    expect(requests.clearAuthenticated).toHaveBeenCalledTimes(1)
    expect(reload).toHaveBeenCalledTimes(1)
    expect(readPending()).toEqual(pending)
    expect(vi.getTimerCount()).toBe(0)
    await vi.advanceTimersByTimeAsync(30000)
    expect(requests.status).toHaveBeenCalledTimes(1)
    setActivePinia(createPinia())
    requests.status.mockResolvedValueOnce(response(status('restarting')))
      .mockResolvedValue(response(status('rolledBack')))
    const restored = useStore()
    await restored.start()
    await vi.advanceTimersByTimeAsync(3000)
    expect(restored.phase).toBe('rolledBack')
    expect(sessionStorage.getItem(storageKey)).toBeNull()
    expect(vi.getTimerCount()).toBe(0)
    expect(requests.install).not.toHaveBeenCalled()
    expect(reload).toHaveBeenCalledTimes(1)
  })

  it.each([
    ['queued', false], ['downloading', false], ['verifying', false],
    ['backingUp', true], ['installing', true], ['restarting', true], ['rollingBack', true],
  ])('pauses ordinary data polling for %s: %s', async (phase, pause) => {
    requests.status.mockResolvedValue(response(status(phase as string)))
    const store = useStore()
    await store.start()
    expect(store.pauseDataPolling).toBe(pause)
  })
})

describe('recovering an installation whose POST response was lost', () => {
  it.each([
    ['network timeout', new Error('timeout')],
    ['HTTP 502', { response: { status: 502 } }],
  ])('adopts the active backend job after %s without posting again', async (_name, error) => {
    const store = await prepareInstall()
    requests.install.mockRejectedValueOnce(error)
    requests.status.mockResolvedValueOnce(response(status('downloading')))
      .mockResolvedValue(response(status('succeeded')))
    await store.install()
    await flushRequests()
    expect(store.phase).toBe('downloading')
    expect(store.confirming).toBe(false)
    expect(store.busy).toBe(true)
    expect(readPending()).toEqual(pending)
    await store.install()
    await store.check()
    expect(requests.install).toHaveBeenCalledTimes(1)
    expect(requests.check).toHaveBeenCalledTimes(1)
    expect(requests.legacyInstall).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(3000)
    expect(store.phase).toBe('succeeded')
    expect(reload).toHaveBeenCalledTimes(1)
    expect(store.busy).toBe(false)
    expect(sessionStorage.getItem(storageKey)).toBeNull()
    expect(vi.getTimerCount()).toBe(0)
  })

  it.each(['succeeded', 'rolledBack', 'failed'])('recognizes a new job already in %s after a lost response', async (phase) => {
    const store = await prepareInstall(status('failed', 'job-previous'))
    requests.install.mockRejectedValueOnce(new Error('response lost'))
    requests.status.mockResolvedValue(response(status(phase)))
    await store.install()
    await flushRequests()
    expect(store.phase).toBe(phase)
    expect(store.confirming).toBe(false)
    expect(store.busy).toBe(false)
    expect(sessionStorage.getItem(storageKey)).toBeNull()
    expect(vi.getTimerCount()).toBe(0)
    expect(reload).toHaveBeenCalledTimes(phase === 'succeeded' ? 1 : 0)
    expect(requests.install).toHaveBeenCalledTimes(1)
  })

  it.each(['idle', 'succeeded', 'rolledBack', 'failed'])('does not mistake an old %s record for this submission', async (phase) => {
    const store = await prepareInstall(status('failed', 'job-previous'))
    requests.install.mockImplementationOnce(() => {
      expect(readPending()).toEqual({ targetVersion, previousJob: 'job-previous' })
      return Promise.reject(new Error('response lost'))
    })
    requests.status.mockResolvedValue(response(status(phase, 'job-previous')))
    await store.install()
    await flushRequests()
    expect(store.phase).toBe(phase)
    expect(store.statusError).toBe('updater.notAccepted')
    expect(store.busy).toBe(false)
    expect(store.confirming).toBe(false)
    expect(sessionStorage.getItem(storageKey)).toBeNull()
    expect(vi.getTimerCount()).toBe(0)
    expect(reload).not.toHaveBeenCalled()
  })

  it('requires a matching target before adopting a new terminal record', async () => {
    const store = await prepareInstall()
    requests.install.mockRejectedValueOnce(new Error('response lost'))
    requests.status.mockResolvedValue(response({
      ...status('succeeded'), currentVersion: '1.7.0',
      state: { ...status('succeeded').state, targetVersion: '1.7.0' },
    }))
    await store.install()
    await flushRequests()
    expect(store.statusError).toBe('updater.notAccepted')
    expect(store.busy).toBe(false)
    expect(sessionStorage.getItem(storageKey)).toBeNull()
    expect(reload).not.toHaveBeenCalled()
  })

  it('tracks an actually active job even when it was already present', async () => {
    const store = await prepareInstall(status('failed', 'job-previous'))
    requests.install.mockRejectedValueOnce(new Error('response lost'))
    requests.status.mockResolvedValue(response({
      ...status('downloading', 'job-previous'),
      state: { ...status('downloading', 'job-previous').state, targetVersion: '1.7.0' },
    }))
    await store.install()
    await flushRequests()
    expect(store.active).toBe(true)
    expect(store.busy).toBe(true)
    expect(store.confirming).toBe(false)
    expect(readPending()).toEqual({ job: 'job-previous', targetVersion: '1.7.0' })
    expect(vi.getTimerCount()).toBe(1)
  })

  it('keeps an offline intent busy and confirms it through GET when connectivity returns', async () => {
    const store = await prepareInstall()
    requests.install.mockRejectedValueOnce(new Error('response lost'))
    requests.status.mockRejectedValue(new Error('offline'))
    await store.install()
    await flushRequests()
    expect(store.confirming).toBe(true)
    expect(store.busy).toBe(true)
    expect(store.pauseDataPolling).toBe(true)
    expect(store.disconnected).toBe(true)
    expect(readPending()).toEqual({ targetVersion })
    const calls = requests.status.mock.calls.length
    await store.install()
    await store.check()
    await vi.advanceTimersByTimeAsync(3000)
    expect(requests.status).toHaveBeenCalledTimes(calls + 1)
    expect(requests.install).toHaveBeenCalledTimes(1)
    expect(requests.check).toHaveBeenCalledTimes(1)
    requests.status.mockResolvedValueOnce(response(status('restarting')))
      .mockResolvedValue(response(status('succeeded')))
    await vi.advanceTimersByTimeAsync(3000)
    expect(store.confirming).toBe(false)
    expect(readPending()).toEqual(pending)
    await vi.advanceTimersByTimeAsync(3000)
    expect(reload).toHaveBeenCalledTimes(1)
    expect(store.busy).toBe(false)
  })

  it('keeps an unconfirmed intent after the deadline and resumes only status requests', async () => {
    const store = await prepareInstall()
    requests.install.mockRejectedValueOnce(new Error('response lost'))
    requests.status.mockRejectedValue(new Error('offline'))
    await store.install()
    await flushRequests()
    await vi.advanceTimersByTimeAsync(25 * 60 * 1000)
    expect(store.monitorExpired).toBe(true)
    expect(store.confirming).toBe(true)
    expect(store.busy).toBe(true)
    expect(readPending()).toEqual({ targetVersion })
    expect(vi.getTimerCount()).toBe(0)
    const calls = requests.status.mock.calls.length
    await vi.advanceTimersByTimeAsync(30000)
    await store.check()
    await store.install()
    expect(requests.status).toHaveBeenCalledTimes(calls)
    expect(requests.install).toHaveBeenCalledTimes(1)
    expect(requests.check).toHaveBeenCalledTimes(1)
    store.resumeMonitoring()
    await flushRequests()
    expect(requests.status).toHaveBeenCalledTimes(calls + 1)
    expect(requests.install).toHaveBeenCalledTimes(1)
    expect(store.monitorExpired).toBe(false)
    expect(vi.getTimerCount()).toBe(1)
  })

  it('restores an unconfirmed submission after a full page reload without resubmitting', async () => {
    const store = await prepareInstall(status('failed', 'job-previous'))
    requests.install.mockRejectedValueOnce(new Error('response lost'))
    requests.status.mockRejectedValue(new Error('offline'))
    await store.install()
    await flushRequests()
    expect(readPending()).toEqual({ targetVersion, previousJob: 'job-previous' })
    // A page reload discards JS timers without calling the explicit logout stop.
    vi.clearAllTimers()
    setActivePinia(createPinia())
    const restored = useStore()
    await restored.start()
    expect(restored.confirming).toBe(true)
    expect(restored.pauseDataPolling).toBe(true)
    requests.status.mockResolvedValue(response(status('succeeded')))
    await vi.advanceTimersByTimeAsync(3000)
    expect(restored.phase).toBe('succeeded')
    expect(reload).toHaveBeenCalledTimes(1)
    expect(requests.install).toHaveBeenCalledTimes(1)
    expect(sessionStorage.getItem(storageKey)).toBeNull()
  })

  it.each(['post', 'confirmation status'])('keeps the intent and reloads for login after a 401 from %s', async (source) => {
    const store = await prepareInstall()
    if (source === 'post') requests.install.mockRejectedValueOnce({ response: { status: 401 } })
    else {
      requests.install.mockRejectedValueOnce(new Error('response lost'))
      requests.status.mockRejectedValueOnce({ response: { status: 401 } })
    }
    await store.install()
    await flushRequests()
    expect(readPending()).toEqual({ targetVersion })
    expect(requests.clearAuthenticated).toHaveBeenCalledTimes(1)
    expect(reload).toHaveBeenCalledTimes(1)
    expect(vi.getTimerCount()).toBe(0)
    const calls = requests.status.mock.calls.length
    await vi.advanceTimersByTimeAsync(30000)
    expect(requests.status).toHaveBeenCalledTimes(calls)
    expect(requests.install).toHaveBeenCalledTimes(1)

    setActivePinia(createPinia())
    requests.status.mockResolvedValue(response(status('rolledBack')))
    await useStore().start()
    expect(sessionStorage.getItem(storageKey)).toBeNull()
    expect(requests.install).toHaveBeenCalledTimes(1)
    expect(reload).toHaveBeenCalledTimes(1)
  })

  it('treats an explicit backend rejection as final without inventing a failed job', async () => {
    const store = await prepareInstall()
    const calls = requests.status.mock.calls.length
    requests.install.mockResolvedValueOnce({ data: { success: false, msg: 'update unavailable', obj: null } })
    await store.install()
    expect(store.phase).toBe('idle')
    expect(store.statusError).toBe('update unavailable')
    expect(store.busy).toBe(false)
    expect(store.confirming).toBe(false)
    expect(sessionStorage.getItem(storageKey)).toBeNull()
    expect(vi.getTimerCount()).toBe(0)
    expect(requests.status).toHaveBeenCalledTimes(calls)
    expect(requests.legacyInstall).not.toHaveBeenCalled()
    expect(reload).not.toHaveBeenCalled()
  })

  it('ignores a late failed POST after logout instead of recreating an intent or polling', async () => {
    const store = await prepareInstall()
    const late = deferred<ReturnType<typeof installResponse>>()
    requests.install.mockReturnValueOnce(late.promise)
    const install = store.install()
    expect(readPending()).toEqual({ targetVersion })
    store.stop()
    late.reject(new Error('late timeout'))
    await install
    await flushRequests()
    expect(store.phase).toBe('idle')
    expect(store.busy).toBe(false)
    expect(store.confirming).toBe(false)
    expect(sessionStorage.getItem(storageKey)).toBeNull()
    expect(vi.getTimerCount()).toBe(0)
    expect(requests.status).toHaveBeenCalledTimes(1)
    expect(reload).not.toHaveBeenCalled()
  })

  it('pauses ordinary data polling when a known downloading job disconnects', async () => {
    requests.status.mockResolvedValueOnce(response(status('downloading')))
      .mockRejectedValue(new Error('offline'))
    const store = useStore()
    await store.start()
    expect(store.pauseDataPolling).toBe(false)
    await vi.advanceTimersByTimeAsync(3000)
    expect(store.disconnected).toBe(true)
    expect(store.pauseDataPolling).toBe(true)
    expect(store.busy).toBe(true)
    expect(store.confirming).toBe(false)
  })
})
