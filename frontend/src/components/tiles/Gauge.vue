<script lang="ts" setup>
import { HumanReadable } from '@/plugins/utils'
import { computed, useId } from 'vue'

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

const gradientId = `resource-gauge-${useId()}`
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

const ticks = Array.from({ length: 21 }, (_, index) => {
  const angle = (135 + index * 13.5) * Math.PI / 180
  const inner = index % 5 === 0 ? 76 : 78
  return {
    x1: 100 + Math.cos(angle) * inner,
    y1: 82 + Math.sin(angle) * inner,
    x2: 100 + Math.cos(angle) * 81,
    y2: 82 + Math.sin(angle) * 81,
  }
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
    <div class="resource-gauge__dial">
      <svg viewBox="0 0 200 160" aria-hidden="true">
        <defs>
          <linearGradient :id="gradientId" x1="0%" y1="100%" x2="100%" y2="0%">
            <stop offset="0%" stop-color="rgb(var(--v-theme-primary))" />
            <stop offset="100%" stop-color="rgb(var(--v-theme-secondary))" />
          </linearGradient>
        </defs>
        <g class="resource-gauge__ticks">
          <line v-for="(tick, index) in ticks" :key="index" v-bind="tick" />
        </g>
        <path class="resource-gauge__track" d="M 53.331 128.669 A 66 66 0 1 1 146.669 128.669" />
        <path
          v-if="data.percent != null && data.percent > 0"
          class="resource-gauge__progress"
          d="M 53.331 128.669 A 66 66 0 1 1 146.669 128.669"
          pathLength="100"
          :stroke-dasharray="`${data.percent} 100`"
          :stroke="tone === 'primary' ? `url(#${gradientId})` : 'var(--gauge-accent)'"
        />
      </svg>
      <div class="resource-gauge__reading" dir="ltr">
        <span class="resource-gauge__icon"><v-icon :icon="resource.icon" size="16" /></span>
        <div class="resource-gauge__percent">
          <strong>{{ data.percent ?? '—' }}</strong><span v-if="data.percent != null">%</span>
        </div>
        <span class="resource-gauge__label">{{ $t('stats.usage') }}</span>
      </div>
    </div>
    <div class="resource-gauge__summary" dir="ltr">
      <template v-if="data.current != null">
        <span class="resource-gauge__current">{{ data.current }}</span>
        <span class="resource-gauge__separator">/</span>
        <span>{{ data.total }}</span>
      </template>
      <template v-else-if="type === 'g-cpu'">
        <span>0%</span><span class="resource-gauge__scale" /><span>100%</span>
      </template>
      <template v-else>—</template>
    </div>
  </div>
</template>

<style scoped>
.resource-gauge {
  --gauge-accent: rgb(var(--v-theme-primary));
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  width: 100%;
  font-variant-numeric: tabular-nums;
}

.resource-gauge--warning {
  --gauge-accent: rgb(var(--v-theme-warning));
}

.resource-gauge--error {
  --gauge-accent: rgb(var(--v-theme-error));
}

.resource-gauge__dial {
  position: relative;
  width: min(100%, 172px);
  aspect-ratio: 5 / 4;
}

.resource-gauge__dial svg {
  display: block;
  width: 100%;
  height: 100%;
}

.resource-gauge__ticks {
  stroke: rgba(var(--v-theme-on-surface), .18);
  stroke-width: 1.25;
  stroke-linecap: round;
}

.resource-gauge__track,
.resource-gauge__progress {
  fill: none;
  stroke-width: 11;
  stroke-linecap: round;
}

.resource-gauge__track {
  stroke: rgba(var(--v-theme-on-surface), .065);
}

.resource-gauge__reading {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
}

.resource-gauge__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border-radius: 9px;
  color: var(--gauge-accent);
  background: color-mix(in srgb, var(--gauge-accent) 9%, transparent);
}

.resource-gauge__percent {
  display: flex;
  align-items: baseline;
  gap: 2px;
  line-height: 1;
}

.resource-gauge__percent strong {
  font-size: 34px;
  font-weight: 700;
  letter-spacing: -.055em;
}

.resource-gauge__percent > span {
  font-size: 13px;
  color: rgba(var(--v-theme-on-surface), .5);
}

.resource-gauge__label {
  font-size: 10px;
  color: rgba(var(--v-theme-on-surface), .5);
}

.resource-gauge__summary {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  min-height: 18px;
  font-size: 12px;
  color: rgba(var(--v-theme-on-surface), .5);
  white-space: nowrap;
}

.resource-gauge__current {
  color: rgba(var(--v-theme-on-surface), .85);
  font-weight: 600;
}

.resource-gauge__separator {
  opacity: .6;
}

.resource-gauge__scale {
  width: 28px;
  height: 1px;
  background: rgba(var(--v-theme-on-surface), .14);
}
</style>
