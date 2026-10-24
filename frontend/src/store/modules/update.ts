import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { i18n } from '@/locales'
import api from '@/plugins/api'
import HttpUtils, { type Msg } from '@/plugins/httputil'
import { clearAuthenticated } from '@/plugins/auth'

export interface UpdateState {
  phase: string
  job?: string
  targetVersion?: string
  previousVersion?: string
}

export interface UpdateStatus {
  currentVersion: string
  supported: boolean
  reason?: string
  state: UpdateState
}

export interface Release extends UpdateStatus {
  latestVersion: string
  available: boolean
  url: string
}

interface PendingJob {
  job?: string
  targetVersion: string
  previousJob?: string
}

const storageKey = 'm1ss-ui-update-job'
const monitoringDuration = 25 * 60 * 1000

const Update = defineStore('Update', () => {
  const status = ref<UpdateStatus>()
  const release = ref<Release>()
  const checking = ref(false)
  const refreshing = ref(false)
  const starting = ref(false)
  const disconnected = ref(false)
  const monitorExpired = ref(false)
  const statusError = ref('')
  const pendingJob = ref<PendingJob>()
  const phase = computed(() => status.value?.state.phase || 'idle')
  const active = computed(() => ['queued', 'downloading', 'verifying', 'backingUp', 'installing', 'restarting', 'rollingBack'].includes(phase.value))
  const confirming = computed(() => !!pendingJob.value && !pendingJob.value.job)
  const busy = computed(() => active.value || starting.value || !!pendingJob.value)
  const alertType = computed(() => confirming.value ? 'info' : ['failed', 'interrupted', 'rolledBack'].includes(phase.value) ? 'warning' : phase.value === 'succeeded' ? 'success' : 'info')
  const pauseDataPolling = computed(() => ['backingUp', 'installing', 'restarting', 'rollingBack'].includes(phase.value)
    || confirming.value || ((!status.value || disconnected.value) && !!pendingJob.value))

  let timer: ReturnType<typeof setTimeout> | undefined
  let deadline = 0
  let running = false
  let generation = 0

  const clearPendingJob = () => {
    pendingJob.value = undefined
    sessionStorage.removeItem(storageKey)
  }

  const rememberJob = (state: UpdateState) => {
    if (!state.job || !state.targetVersion) return
    pendingJob.value = { job: state.job, targetVersion: state.targetVersion }
    sessionStorage.setItem(storageKey, JSON.stringify(pendingJob.value))
  }

  const applyStatus = (next: UpdateStatus) => {
    const wasConfirming = confirming.value
    const matchesSubmission = wasConfirming && !!next.state.job
      && next.state.job !== pendingJob.value?.previousJob
      && next.state.targetVersion === pendingJob.value?.targetVersion
    if (matchesSubmission) rememberJob(next.state)
    status.value = next
    disconnected.value = false
    statusError.value = wasConfirming && !matchesSubmission && !active.value
      ? i18n.global.t('updater.notAccepted') : ''
    if (active.value) {
      rememberJob(next.state)
      if (!deadline) deadline = Date.now() + monitoringDuration
      return
    }

    const shouldReload = phase.value === 'succeeded'
      && pendingJob.value?.job === next.state.job
      && pendingJob.value?.targetVersion === next.currentVersion
    // Clear before reloading so the historical result cannot cause a loop.
    clearPendingJob()
    clearTimeout(timer)
    deadline = 0
    monitorExpired.value = false
    if (shouldReload) window.location.reload()
  }

  const expireSession = () => {
    clearTimeout(timer)
    deadline = 0
    statusError.value = i18n.global.t('updater.sessionExpired')
    monitorExpired.value = true
    clearAuthenticated()
    // A full navigation preserves the pending job for the next login.
    window.location.reload()
  }

  const refreshStatus = async () => {
    if (!running || refreshing.value) return
    clearTimeout(timer)
    refreshing.value = true
    const currentGeneration = generation
    try {
      // Service restart is expected to interrupt these requests.
      const response = await api.get<Msg<UpdateStatus>>('api/updateStatus', { timeout: 4000 })
      if (!running || generation !== currentGeneration) return
      if (!response.data.success || !response.data.obj) throw new Error(response.data.msg)
      applyStatus(response.data.obj)
    } catch (error: unknown) {
      if (!running || generation !== currentGeneration) return
      if ((error as { response?: { status?: number } }).response?.status === 401) {
        expireSession()
        return
      }
      if (!deadline) {
        statusError.value = i18n.global.t('updater.statusError')
        return
      }
      disconnected.value = true
    } finally {
      if (generation === currentGeneration) refreshing.value = false
    }
    if (!running || generation !== currentGeneration || !deadline) return
    if (Date.now() < deadline) timer = setTimeout(refreshStatus, 3000)
    else monitorExpired.value = true
  }

  const resumeMonitoring = () => {
    if (!running) return
    clearTimeout(timer)
    monitorExpired.value = false
    deadline = Date.now() + monitoringDuration
    void refreshStatus()
  }

  const start = async () => {
    if (running) return
    running = true
    const saved = sessionStorage.getItem(storageKey)
    pendingJob.value = saved ? JSON.parse(saved) as PendingJob : undefined
    if (pendingJob.value) deadline = Date.now() + monitoringDuration
    await refreshStatus()
  }

  const stop = () => {
    running = false
    generation++
    clearTimeout(timer)
    deadline = 0
    refreshing.value = false
    clearPendingJob()
    status.value = undefined
    release.value = undefined
    checking.value = false
    starting.value = false
    disconnected.value = false
    monitorExpired.value = false
    statusError.value = ''
  }

  const check = async () => {
    if (!running || refreshing.value || checking.value || busy.value) return
    checking.value = true
    const currentGeneration = generation
    try {
      const msg = await HttpUtils.get<Release>('api/checkUpdate', {}, { timeout: 20000 })
      if (!running || generation !== currentGeneration) return
      if (msg.success) {
        release.value = msg.obj
        applyStatus(msg.obj)
        if (active.value) resumeMonitoring()
      } else statusError.value = msg.msg
    } finally {
      if (generation === currentGeneration) checking.value = false
    }
  }

  const install = async () => {
    if (!running || refreshing.value || checking.value || !release.value || busy.value) return
    starting.value = true
    const currentGeneration = generation
    pendingJob.value = { targetVersion: release.value.latestVersion, previousJob: status.value?.state.job }
    sessionStorage.setItem(storageKey, JSON.stringify(pendingJob.value))
    statusError.value = ''
    try {
      const response = await api.post<Msg<UpdateState>>('api/update', { version: release.value.latestVersion }, { timeout: 20000 })
      if (!running || generation !== currentGeneration) return
      const msg = response.data
      if (msg.success && status.value) {
        status.value.state = msg.obj
        rememberJob(msg.obj)
        resumeMonitoring()
      } else {
        clearPendingJob()
        statusError.value = msg.msg
      }
    } catch (error: unknown) {
      if (!running || generation !== currentGeneration) return
      if ((error as { response?: { status?: number } }).response?.status === 401) {
        expireSession()
        return
      }
      // The worker may have started even though its acknowledgement was lost.
      resumeMonitoring()
    } finally {
      if (generation === currentGeneration) starting.value = false
    }
  }

  return {
    status, release, checking, refreshing, starting, disconnected, monitorExpired, statusError,
    phase, active, confirming, busy, alertType, pauseDataPolling,
    start, stop, check, install, resumeMonitoring,
  }
})

export default Update
