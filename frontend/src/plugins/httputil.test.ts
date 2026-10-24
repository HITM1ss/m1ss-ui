import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises } from '@vue/test-utils'
import Data, { type LoadedData } from '@/store/modules/data'

const mocks = vi.hoisted(() => ({
  get: vi.fn(), post: vi.fn(), error: vi.fn(), success: vi.fn(),
  navigate: vi.fn(), clearAuthenticated: vi.fn(),
}))
vi.mock('@/plugins/api', () => ({ default: { get: mocks.get, post: mocks.post } }))
vi.mock('@/router', () => ({ default: { push: mocks.navigate } }))
vi.mock('@/plugins/auth', () => ({ clearAuthenticated: mocks.clearAuthenticated }))
vi.mock('notivue', () => ({ push: { error: mocks.error, success: mocks.success } }))
vi.mock('@/locales', () => ({ i18n: { global: { t: (key: string) => key } } }))

const payload = (tag: string): LoadedData => ({
  config: {}, inbounds: [{ id: 1, type: 'direct', tag }],
  onlines: { inbound: [tag] }, maintenance: true, os: 'linux',
} as LoadedData)
const response = (obj: LoadedData, msg = '') => ({ data: { success: true, msg, obj } })

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason: unknown) => void
  const promise = new Promise<T>((done, fail) => { resolve = done; reject = fail })
  return { promise, resolve, reject }
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.resetAllMocks()
  mocks.get.mockResolvedValue({ data: { success: true, msg: '', obj: null } })
  Data().setNewData(payload('original'))
})

afterEach(() => { vi.restoreAllMocks() })

// Keep Data and HttpUtils real: a store-only mock would miss the toast and
// logout that happen inside HttpUtils before Data receives its response.
describe('late data responses during an update or layout teardown', () => {
  it.each([
    ['network failure', new Error('offline')],
    ['401', { response: { status: 401 } }],
  ])('ignores a late %s without toast, logout or state changes', async (_name, error) => {
    const late = deferred<ReturnType<typeof response>>()
    mocks.get.mockReturnValueOnce(late.promise)
    const store = Data()
    let shouldHandle = true
    const load = store.loadData(() => shouldHandle)
    expect(mocks.get).toHaveBeenCalledTimes(1)
    shouldHandle = false
    late.reject(error)
    await load
    await flushPromises()
    expect(mocks.error).not.toHaveBeenCalled()
    expect(mocks.success).not.toHaveBeenCalled()
    expect(mocks.clearAuthenticated).not.toHaveBeenCalled()
    expect(mocks.navigate).not.toHaveBeenCalled()
    expect(mocks.get).toHaveBeenCalledTimes(1)
    expect(store.inbounds[0].tag).toBe('original')

    shouldHandle = true
    mocks.get.mockResolvedValueOnce(response(payload('resumed')))
    await store.loadData(() => shouldHandle)
    expect(store.inbounds[0].tag).toBe('resumed')
    expect(store.onlines.inbound).toEqual(['resumed'])
  })

  it('ignores late successful data and its messages, then accepts a fresh load after resume', async () => {
    const late = deferred<ReturnType<typeof response>>()
    mocks.get.mockReturnValueOnce(late.promise)
    const store = Data()
    const original = JSON.stringify(store.$state)
    let shouldHandle = true
    const load = store.loadData(() => shouldHandle)
    shouldHandle = false
    late.resolve(response({ ...payload('stale'), lastLog: 'stale core error' }, 'save'))
    await load
    expect(JSON.stringify(store.$state)).toBe(original)
    expect(mocks.error).not.toHaveBeenCalled()
    expect(mocks.success).not.toHaveBeenCalled()
    expect(mocks.clearAuthenticated).not.toHaveBeenCalled()
    expect(mocks.navigate).not.toHaveBeenCalled()

    shouldHandle = true
    mocks.get.mockResolvedValueOnce(response(payload('fresh')))
    await store.loadData(() => shouldHandle)
    expect(store.inbounds[0].tag).toBe('fresh')
    expect(store.onlines.inbound).toEqual(['fresh'])
  })

  it('still handles an actual 401 after ordinary polling resumes', async () => {
    let shouldHandle = true
    const late = deferred<ReturnType<typeof response>>()
    mocks.get.mockReturnValueOnce(late.promise)
    const load = Data().loadData(() => shouldHandle)
    shouldHandle = false
    late.reject({ response: { status: 401 } })
    await load
    expect(mocks.clearAuthenticated).not.toHaveBeenCalled()

    shouldHandle = true
    mocks.get.mockRejectedValueOnce({ response: { status: 401 } })
    await Data().loadData(() => shouldHandle)
    await flushPromises()
    expect(mocks.error).toHaveBeenCalledWith({ title: 'invalidLogin' })
    expect(mocks.get.mock.calls.some(([url]) => url === 'api/logout')).toBe(true)
    expect(mocks.clearAuthenticated).toHaveBeenCalledTimes(1)
    expect(mocks.navigate).toHaveBeenCalledWith('/login')
    expect(Data().inbounds).toEqual([])
  })
})
