<template>
  <v-card
    class="panel-update mb-5"
    variant="flat"
  >
    <v-card-text>
      <div class="panel-update__header">
        <div class="panel-update__heading">
          <span class="panel-update__icon"><v-icon
            icon="mdi-cloud-sync-outline"
            size="24"
          /></span>
          <div>
            <h2>{{ $t('updater.title') }}</h2>
            <p>{{ $t('updater.current') }} · v{{ status?.currentVersion || '—' }}</p>
          </div>
        </div>
        <div class="panel-update__actions">
          <v-btn
            variant="tonal"
            color="primary"
            prepend-icon="mdi-refresh"
            :loading="checking"
            :disabled="active || starting"
            @click="check"
          >
            {{ $t('updater.check') }}
          </v-btn>
          <v-btn
            v-if="release?.available && status?.supported"
            color="primary"
            variant="flat"
            prepend-icon="mdi-download-outline"
            :disabled="disabled || active || starting"
            @click="confirmUpdate = true"
          >
            {{ $t('updater.install') }} v{{ release.latestVersion }}
          </v-btn>
        </div>
      </div>
      <p
        v-if="release"
        class="panel-update__release"
      >
        {{ release.available ? $t('updater.available', { version: release.latestVersion }) : $t('updater.latest') }}
        <a
          :href="release.url"
          target="_blank"
          rel="noopener noreferrer"
        >{{ $t('updater.notes') }} ↗</a>
      </p>
      <v-alert
        v-if="status && !status.supported"
        class="mt-4"
        type="info"
        variant="tonal"
        density="compact"
        :text="$t(`updater.unsupported.${status.reason || 'installation'}`)"
      />
      <v-alert
        v-if="disabled && release?.available && status?.supported"
        class="mt-4"
        type="info"
        variant="tonal"
        density="compact"
        :text="$t('updater.saveFirst')"
      />
      <v-alert
        v-if="phase !== 'idle'"
        class="mt-4"
        :type="alertType"
        variant="tonal"
        density="compact"
      >
        {{ disconnected ? $t('updater.reconnecting') : $t(`updater.phase.${phase}`) }}
        <span v-if="phase === 'failed' || phase === 'interrupted'"> {{ $t('updater.recovery') }}</span>
      </v-alert>
      <v-progress-linear
        v-if="active"
        indeterminate
        color="primary"
        rounded
        class="mt-4"
        :aria-label="$t('updater.title')"
      />
      <v-btn
        v-if="monitorExpired"
        variant="text"
        class="mt-2"
        @click="resumeMonitoring"
      >
        {{ $t('updater.reconnect') }}
      </v-btn>
      <p
        v-if="statusError"
        class="panel-update__error"
        role="alert"
      >
        {{ statusError }}
      </p>
    </v-card-text>
    <v-dialog
      v-model="confirmUpdate"
      max-width="460"
    >
      <v-card rounded="xl">
        <v-card-title>{{ $t('updater.install') }} v{{ release?.latestVersion }}</v-card-title>
        <v-card-text>{{ $t('updater.confirm') }}</v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="confirmUpdate = false">
            {{ $t('actions.close') }}
          </v-btn>
          <v-btn
            color="primary"
            variant="flat"
            :disabled="disabled || starting || active"
            @click="install"
          >
            {{ $t('updater.install') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { i18n } from '@/locales'
import api from '@/plugins/api'
import HttpUtils, { type Msg } from '@/plugins/httputil'

const props = defineProps<{ disabled?: boolean }>()

interface UpdateState {
  phase: string
  targetVersion?: string
  previousVersion?: string
}
interface UpdateStatus {
  currentVersion: string
  supported: boolean
  reason?: string
  state: UpdateState
}
interface Release extends UpdateStatus {
  latestVersion: string
  available: boolean
  url: string
}

const status = ref<UpdateStatus>()
const release = ref<Release>()
const checking = ref(false)
const starting = ref(false)
const confirmUpdate = ref(false)
const disconnected = ref(false)
const monitorExpired = ref(false)
const statusError = ref('')
const phase = computed(() => status.value?.state.phase || 'idle')
const active = computed(() => ['queued', 'downloading', 'verifying', 'backingUp', 'installing', 'restarting', 'rollingBack'].includes(phase.value))
const alertType = computed(() => ['failed', 'interrupted', 'rolledBack'].includes(phase.value) ? 'warning' : phase.value === 'succeeded' ? 'success' : 'info')
let timer: ReturnType<typeof setTimeout> | undefined
let deadline = 0
let disposed = false
let acceptedUpdate = false

const check = async () => {
  checking.value = true
  try {
    const msg = await HttpUtils.get<Release>('api/checkUpdate', {}, { timeout: 20000 })
    if (disposed) return
    if (msg.success) {
      release.value = msg.obj
      status.value = msg.obj
      statusError.value = ''
      if (active.value) resumeMonitoring()
    } else statusError.value = msg.msg
  } finally { checking.value = false }
}

const refreshStatus = async () => {
  try {
    // Expected disconnections during service restart should not create a toast
    // every three seconds or log the user out on a transient network error.
    const response = await api.get<Msg<UpdateStatus>>('api/updateStatus', { timeout: 4000 })
    if (disposed) return
    if (!response.data.success || !response.data.obj) throw new Error(response.data.msg)
    status.value = response.data.obj
    disconnected.value = false
    statusError.value = ''
    if (!active.value) {
      if (acceptedUpdate && phase.value === 'succeeded') {
        window.location.reload()
      }
      return
    }
  } catch (error: unknown) {
    if (disposed) return
    if ((error as { response?: { status?: number } }).response?.status === 401) {
      statusError.value = i18n.global.t('updater.sessionExpired')
      monitorExpired.value = true
      return
    }
    if (!deadline) {
      statusError.value = i18n.global.t('updater.statusError')
      return
    }
    disconnected.value = true
  }
  if (Date.now() < deadline) timer = setTimeout(refreshStatus, 3000)
  else monitorExpired.value = true
}

const resumeMonitoring = () => {
  clearTimeout(timer)
  monitorExpired.value = false
  deadline = Date.now() + 25 * 60 * 1000
  void refreshStatus()
}

const install = async () => {
  if (!release.value || props.disabled || active.value || starting.value) return
  confirmUpdate.value = false
  starting.value = true
  try {
    const msg = await HttpUtils.post<UpdateState>('api/update', { version: release.value.latestVersion }, { timeout: 20000 })
    if (disposed) return
    if (msg.success && status.value) {
      status.value.state = msg.obj
      acceptedUpdate = true
      resumeMonitoring()
    } else statusError.value = msg.msg
  } finally { starting.value = false }
}

onMounted(async () => {
  await refreshStatus()
  if (!disposed && active.value) resumeMonitoring()
})
onBeforeUnmount(() => { disposed = true; clearTimeout(timer) })
</script>

<style scoped>
.panel-update { border: 1px solid var(--panel-border); border-radius: var(--panel-radius); }
.panel-update > :deep(.v-card-text) { padding: 24px; }
.panel-update__header, .panel-update__heading, .panel-update__actions { display: flex; align-items: center; gap: 12px; }
.panel-update__header { justify-content: space-between; flex-wrap: wrap; gap: 20px; }
.panel-update__actions { flex-wrap: wrap; gap: 8px; }
.panel-update__icon { display: grid; place-items: center; flex: 0 0 44px; width: 44px; height: 44px; border-radius: 14px; background: var(--panel-gradient); color: #fff; }
.panel-update h2 { font-size: 1rem; font-weight: 600; }
.panel-update__heading p, .panel-update__release { font-size: .8125rem; color: var(--panel-muted); }
.panel-update__release { margin-top: 16px; }
.panel-update__release a { display: inline-block; margin-inline-start: 12px; color: rgb(var(--v-theme-primary)); }
.panel-update__error { margin-top: 12px; font-size: .8125rem; color: rgb(var(--v-theme-error)); overflow-wrap: anywhere; }
</style>
