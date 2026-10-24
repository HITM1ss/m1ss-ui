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
            :loading="checking || refreshing"
            :disabled="busy"
            @click="check"
          >
            {{ $t('updater.check') }}
          </v-btn>
          <v-btn
            v-if="release?.available && status?.supported"
            color="primary"
            variant="flat"
            prepend-icon="mdi-download-outline"
            :disabled="disabled || busy || checking || refreshing"
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
        v-if="phase !== 'idle' || confirming || monitorExpired"
        class="mt-4"
        :type="alertType"
        variant="tonal"
        density="compact"
      >
        {{ monitorExpired ? $t('updater.monitorExpired') : confirming ? $t('updater.confirming') : disconnected ? $t('updater.reconnecting') : $t(`updater.phase.${phase}`) }}
        <span v-if="!confirming && (phase === 'failed' || phase === 'interrupted')"> {{ $t('updater.recovery') }}</span>
      </v-alert>
      <v-progress-linear
        v-if="(active || confirming) && !monitorExpired"
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
            :disabled="disabled || busy || checking || refreshing"
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
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import Update from '@/store/modules/update'

const props = defineProps<{ disabled?: boolean }>()

const update = Update()
const {
  status, release, checking, refreshing, disconnected, monitorExpired, statusError,
  phase, active, confirming, busy, alertType,
} = storeToRefs(update)
const { check, resumeMonitoring } = update
const confirmUpdate = ref(false)

const install = async () => {
  if (props.disabled || busy.value) return
  confirmUpdate.value = false
  await update.install()
}
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
