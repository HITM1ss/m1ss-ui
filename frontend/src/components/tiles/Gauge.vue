<script lang="ts" setup>
import { HumanReadable } from '@/plugins/utils'
import { computed } from 'vue'

// One resource reading. Memory, disk and swap report a pair; cpu is a percent.
interface Usage {
  current: number
  total: number
}

interface TilesData {
  cpu?: number
  mem?: Usage
  dsk?: Usage
  swp?: Usage
}

interface GaugeReading {
  percent: number | null
  current?: string
  total?: string
}

const props = defineProps<{
  tilesData: TilesData
  type: string
}>()

const resources: Record<string, { icon: string; label: string }> = {
  'g-cpu': { icon: 'mdi-cpu-64-bit', label: 'main.gauge.cpu' },
  'g-mem': { icon: 'mdi-memory', label: 'main.gauge.mem' },
  'g-dsk': { icon: 'mdi-harddisk', label: 'main.gauge.dsk' },
  'g-swp': { icon: 'mdi-swap-horizontal', label: 'main.gauge.swp' },
}
const resource = computed(() => resources[props.type] ?? resources['g-cpu'])
const percentage = (value: number): number | null =>
  Number.isFinite(value) ? Math.ceil(Math.min(100, Math.max(0, value))) : null
const formatSize = (value: number) => value === 0 ? '0 B' : HumanReadable.sizeFormat(value, 1)

const gaugeData = (usage?: Usage): GaugeReading => {
  if (!usage || !Number.isFinite(usage.current) || !Number.isFinite(usage.total)
    || usage.current < 0 || usage.total < 0) return { percent: null }
  return {
    percent: usage.total > 0 ? percentage(usage.current * 100 / usage.total) : null,
    current: formatSize(usage.current),
    total: formatSize(usage.total),
  }
}

const data = computed((): GaugeReading => {
  const d = props.tilesData
  switch (props.type) {
    case 'g-cpu':
      return { percent: d.cpu == null ? null : percentage(d.cpu) }
    case 'g-mem':
      return gaugeData(d.mem)
    case 'g-dsk':
      return gaugeData(d.dsk)
    case 'g-swp':
      return gaugeData(d.swp)
  }
  return { percent: null }
})

const tone = computed(() => {
  if ((data.value.percent ?? 0) > 90) return 'error'
  if ((data.value.percent ?? 0) > 70) return 'warning'
  return 'primary'
})

</script>

<template>
  <div
    class="resource-gauge"
    :class="`resource-gauge--${tone}`"
    role="progressbar"
    :aria-label="$t(resource.label)"
    :aria-valuenow="data.percent ?? undefined"
    :aria-valuetext="data.percent == null ? '—' : `${data.percent}%`"
    aria-valuemin="0"
    aria-valuemax="100"
  >
    <div class="resource-gauge__reading">
      <div
        class="resource-gauge__percent"
        dir="ltr"
      >
        <strong>{{ data.percent ?? '—' }}</strong><span v-if="data.percent != null">%</span>
      </div>
      <v-icon
        :icon="resource.icon"
        size="22"
        class="resource-gauge__icon"
      />
    </div>
    <div
      class="resource-gauge__track"
      aria-hidden="true"
    >
      <span
        v-if="data.percent != null && data.percent > 0"
        class="resource-gauge__progress"
        :style="{ transform: `scaleX(${data.percent / 100})` }"
      />
    </div>
    <div
      class="resource-gauge__summary"
      dir="ltr"
    >
      <template v-if="data.current != null">
        <span class="resource-gauge__current">{{ data.current }}</span>
        <span class="resource-gauge__separator">/</span>
        <span>{{ data.total }}</span>
      </template>
      <template v-else-if="type === 'g-cpu'">
        {{ $t('stats.usage') }}
      </template>
      <template v-else>
        —
      </template>
    </div>
  </div>
</template>

<style scoped>
.resource-gauge { --gauge-accent: rgb(var(--v-theme-primary)); width: 100%; }
.resource-gauge--warning { --gauge-accent: rgb(var(--v-theme-warning)); }
.resource-gauge--error { --gauge-accent: rgb(var(--v-theme-error)); }
.resource-gauge__reading { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 20px; }
.resource-gauge__percent { display: flex; align-items: baseline; gap: 4px; line-height: 1; }
.resource-gauge__percent strong { font-size: 2.75rem; font-weight: 800; letter-spacing: -.06em; }
.resource-gauge__percent > span { font-size: 1rem; color: var(--panel-muted); }
.resource-gauge__icon { color: var(--gauge-accent); }
.resource-gauge__track { height: 8px; border: 1px solid var(--panel-border); overflow: hidden; background: var(--accent); }
.resource-gauge__progress { display: block; width: 100%; height: 100%; background: var(--gauge-accent); transform-origin: left; }
.v-locale--is-rtl .resource-gauge__progress { transform-origin: right; }
.resource-gauge__summary { display: flex; align-items: center; gap: 6px; min-height: 18px; margin-top: 12px; font-size: .75rem; color: var(--panel-muted); flex-wrap: wrap; }
.resource-gauge__current { color: rgb(var(--v-theme-on-surface)); font-weight: 500; }
.resource-gauge__separator { opacity: .6; }
@media (max-width: 600px) {
  .resource-gauge__percent strong { font-size: 2.25rem; }
}
</style>
