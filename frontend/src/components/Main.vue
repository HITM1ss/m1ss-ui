<template>
  <LogVue
    v-model="logModal.visible"
    :control="logModal"
    :visible="logModal.visible"
  />
  <Backup
    v-model="backupModal.visible"
    :control="backupModal"
    :visible="backupModal.visible"
  />
  <UsageStats v-model:visible="usageStatsModal.visible" />
  <v-container
    class="dashboard"
    :loading="loading"
  >
    <v-responsive>
      <div class="dashboard-header">
        <div class="dashboard-heading">
          <div>
            <span class="dashboard-eyebrow">M1ss-UI / {{ $t('ui.controlPanel') }}</span>
            <h1>{{ $t('ui.overview') }}</h1>
            <p>{{ tilesData.sys?.hostName || 'M1ss-UI' }}</p>
          </div>
        </div>
        <div class="dashboard-actions">
          <v-dialog
            v-model="menu"
            :close-on-content-click="false"
            transition="scale-transition"
            max-width="800"
          >
            <template #activator="{ props }">
              <v-btn
                v-bind="props"
                class="dashboard-customize"
                variant="flat"
                prepend-icon="mdi-view-dashboard-edit-outline"
              >
                {{ $t('main.tiles') }}
              </v-btn>
            </template>
            <v-card rounded="xl">
              <v-card-title>
                <v-row>
                  <v-col>
                    {{ $t('main.tiles') }}
                  </v-col>
                  <v-spacer />
                  <v-col cols="auto">
                    <v-btn
                      :aria-label="$t('actions.close')"
                      icon="mdi-close"
                      size="small"
                      variant="text"
                      @click="menu = false"
                    />
                  </v-col>
                </v-row>
              </v-card-title>
              <v-divider />
              <v-row
                v-for="(items, mi) in menuItems"
                :key="mi"
                density="compact"
              >
                <v-col cols="12">
                  <v-card
                    :subtitle="items.title"
                    variant="flat"
                  >
                    <v-card-text>
                      <v-row density="compact">
                        <v-col
                          v-for="item in items.value"
                          :key="item.title ?? item"
                          cols="12"
                          md="6"
                          lg="3"
                        >
                          <v-switch
                            v-model="reloadItems"
                            density="compact"
                            :value="item.value"
                            color="primary"
                            :label="item.title"
                            hide-details
                          />
                        </v-col>
                      </v-row>
                    </v-card-text>
                  </v-card>
                </v-col>
              </v-row>
            </v-card>
          </v-dialog>
          <v-btn
            variant="tonal"
            prepend-icon="mdi-backup-restore"
            @click="backupModal.visible = true"
          >
            {{ $t('main.backup.title') }}
          </v-btn>
          <v-btn
            variant="tonal"
            prepend-icon="mdi-list-box-outline"
            @click="logModal.visible = true"
          >
            {{ $t('basic.log.title') }}
          </v-btn>
          <v-btn
            variant="tonal"
            prepend-icon="mdi-chart-box-outline"
            @click="usageStatsModal.visible = true"
          >
            {{ $t('main.stats.title') }}
          </v-btn>
        </div>
      </div>
      <v-card
        v-if="reloadItems.length === 0"
        class="dashboard-empty"
        variant="flat"
      >
        <v-icon
          icon="mdi-view-dashboard-plus-outline"
          color="primary"
          size="40"
        />
        <h2>{{ $t('main.tiles') }}</h2>
        <p>{{ $t('main.gauges') }} · {{ $t('main.charts') }} · {{ $t('main.infos') }}</p>
        <v-btn
          color="primary"
          variant="tonal"
          prepend-icon="mdi-plus"
          @click="menu = true"
        >
          {{ $t('actions.add') }}
        </v-btn>
      </v-card>
      <v-row class="dashboard-grid">
        <v-col
          v-if="selectedGauges.length > 0"
          cols="12"
        >
          <v-card
            class="dashboard-gauges"
            variant="flat"
            tag="section"
            :aria-label="$t('main.gauges')"
          >
            <v-card-text
              class="dashboard-gauges-content"
              :style="{
                '--gauge-count': selectedGauges.length,
                '--gauge-small-count': Math.min(2, selectedGauges.length),
              }"
            >
              <div
                v-for="item in selectedGauges"
                :key="item.value"
                class="dashboard-gauge"
              >
                <h2 class="dashboard-gauge-title">
                  {{ item.title }}
                </h2>
                <Gauge
                  :tiles-data="tilesData"
                  :type="item.value"
                />
              </div>
            </v-card-text>
          </v-card>
        </v-col>
        <v-col
          v-for="i in otherTiles"
          :key="i"
          cols="12"
          sm="6"
          md="6"
          lg="4"
        >
          <v-card
            class="dashboard-tile"
            :class="{ 'dashboard-tile--info': i.charAt(0) === 'i' }"
            variant="flat"
          >
            <v-card-title class="dashboard-tile-title">
              {{ menuItems.flatMap(cat => cat.value).find(m => m.value == i)?.title }}
              <template v-if="i == 'i-sys'">
                <v-btn
                  v-tooltip:top="$t('actions.update')"
                  :aria-label="$t('actions.update')"
                  icon="mdi-refresh"
                  variant="text"
                  color="primary"
                  size="32"
                  class="ms-auto"
                  @click="reloadSys()"
                />
              </template>
              <template v-if="i == 'h-net'">
                <v-icon
                  v-tooltip:top="'↓' +
                    HumanReadable.sizeFormat(tilesData.net?.recv) + ' - ' +
                    HumanReadable.sizeFormat(tilesData.net?.sent) + '↑'"
                  icon="mdi-information"
                  color="primary"
                  size="small"
                  style="margin-inline-start: 10px;"
                />
              </template>
            </v-card-title>
            <v-card-text
              class="dashboard-tile-content"
            >
              <History
                v-if="i.charAt(0) == 'h'"
                :tiles-data="tilesData"
                :type="i"
              />
              <template v-if="i == 'i-sys'">
                <v-row
                  class="dashboard-info"
                  dense
                >
                  <v-col cols="3">
                    {{ $t('main.info.host') }}
                  </v-col>
                  <v-col
                    cols="9"
                    class="dashboard-host"
                  >
                    {{ tilesData.sys?.hostName }}
                  </v-col>
                  <v-col cols="3">
                    {{ $t('main.info.cpu') }}
                  </v-col>
                  <v-col cols="9">
                    <v-chip
                      density="compact"
                      variant="flat"
                    >
                      <v-tooltip
                        activator="parent"
                        location="top"
                        style="direction: ltr;"
                      >
                        {{ tilesData.sys?.cpuType }}
                      </v-tooltip>
                      {{ tilesData.sys?.cpuCount }} {{ $t('main.info.core') }}
                    </v-chip>
                  </v-col>
                  <v-col cols="3">
                    IP
                  </v-col>
                  <v-col cols="9">
                    <v-chip
                      v-if="tilesData.sys?.ipv4?.length>0"
                      density="compact"
                      color="primary"
                      variant="flat"
                    >
                      <v-tooltip
                        activator="parent"
                        location="top"
                        style="direction: ltr;"
                      >
                        <span style="white-space: pre-line">{{ (tilesData.sys?.ipv4 ?? []).join('\n') }}</span>
                      </v-tooltip>
                      IPv4
                    </v-chip>
                    <v-chip
                      v-if="tilesData.sys?.ipv6?.length>0"
                      density="compact"
                      color="primary"
                      variant="flat"
                    >
                      <v-tooltip
                        activator="parent"
                        location="top"
                        style="direction: ltr;"
                      >
                        <span style="white-space: pre-line">{{ (tilesData.sys?.ipv6 ?? []).join('\n') }}</span>
                      </v-tooltip>
                      IPv6
                    </v-chip>
                  </v-col>
                  <v-col cols="3">
                    M1ss-UI
                  </v-col>
                  <v-col cols="9">
                    <v-chip
                      density="compact"
                      color="primary"
                    >
                      v{{ tilesData.sys?.appVersion }}
                    </v-chip>
                  </v-col>
                  <v-col cols="3">
                    {{ $t('main.info.uptime') }}
                  </v-col>
                  <v-col
                    v-tooltip:top="$t('main.info.startupTime')
                      + ': ' + new Date((tilesData.sys?.bootTime || 0) * 1000).toLocaleString(locale)"
                    cols="9"
                  >
                    {{ HumanReadable.formatSecond((Date.now()/1000) - tilesData.sys?.bootTime) }}
                  </v-col>
                </v-row>
              </template>
              <template v-if="i == 'i-sbd'">
                <v-row
                  class="dashboard-info"
                  dense
                >
                  <v-col cols="4">
                    {{ $t('main.info.running') }}
                  </v-col>
                  <v-col cols="8">
                    <v-chip
                      v-if="tilesData.sbd?.running"
                      density="compact"
                      color="success"
                      variant="flat"
                    >
                      {{ $t('yes') }}
                    </v-chip>
                    <!-- A core stopped on purpose reads the same as a crashed
                         one here, and the operator has to be able to tell. -->
                    <v-chip
                      v-else-if="tilesData.sbd?.maintenance"
                      density="compact"
                      color="warning"
                      variant="flat"
                    >
                      {{ $t('setting.maintenance') }}
                    </v-chip>
                    <v-chip
                      v-else
                      density="compact"
                      color="error"
                      variant="flat"
                    >
                      {{ $t('no') }}
                    </v-chip>
                    <v-chip
                      v-if="tilesData.sbd?.running && !loading"
                      density="compact"
                      color="transparent"
                      style="cursor: pointer;"
                      @click="restartSingbox()"
                    >
                      <v-tooltip
                        activator="parent"
                        location="top"
                      >
                        {{ $t('actions.restartSb') }}
                      </v-tooltip>
                      <v-icon
                        icon="mdi-restart"
                        color="warning"
                      />
                    </v-chip>
                  </v-col>
                  <v-col cols="4">
                    {{ $t('main.info.memory') }}
                  </v-col>
                  <v-col cols="8">
                    <v-chip
                      v-if="tilesData.sbd?.stats?.Alloc"
                      density="compact"
                      color="primary"
                      variant="flat"
                    >
                      {{ HumanReadable.sizeFormat(tilesData.sbd?.stats?.Alloc) }}
                    </v-chip>
                  </v-col>
                  <v-col cols="4">
                    {{ $t('main.info.threads') }}
                  </v-col>
                  <v-col cols="8">
                    <v-chip
                      v-if="tilesData.sbd?.stats?.NumGoroutine"
                      density="compact"
                      color="primary"
                      variant="flat"
                    >
                      {{ tilesData.sbd?.stats?.NumGoroutine }}
                    </v-chip>
                  </v-col>
                  <v-col cols="4">
                    {{ $t('main.info.uptime') }}
                  </v-col>
                  <v-col cols="8">
                    {{ HumanReadable.formatSecond(tilesData.sbd?.stats?.Uptime) }}
                  </v-col>
                  <v-col cols="4">
                    {{ $t('online') }}
                  </v-col>
                  <v-col cols="8">
                    <template v-if="tilesData.sbd?.running">
                      <v-chip
                        v-if="Data().onlines.user"
                        density="compact"
                        color="primary"
                        variant="flat"
                      >
                        <v-tooltip
                          activator="parent"
                          location="top"
                          overflow="auto"
                        >
                          <span
                            style="font-weight: bold;"
                            v-text="$t('pages.clients')"
                          /><br>
                          <span
                            v-for="user in Data().onlines.user"
                            :key="user"
                          >{{ user }}<br></span>
                        </v-tooltip>
                        {{ Data().onlines.user?.length }}
                      </v-chip>
                      <v-chip
                        v-if="Data().onlines.inbound"
                        density="compact"
                        color="success"
                        variant="flat"
                      >
                        <v-tooltip
                          activator="parent"
                          location="top"
                          :text="$t('pages.inbounds')"
                        >
                          <span
                            style="font-weight: bold;"
                            v-text="$t('pages.inbounds')"
                          /><br>
                          <span
                            v-for="tag in Data().onlines.inbound"
                            :key="tag"
                          >{{ tag }}<br></span>
                        </v-tooltip>
                        {{ Data().onlines.inbound?.length }}
                      </v-chip>
                      <v-chip
                        v-if="Data().onlines.outbound"
                        density="compact"
                        color="info"
                        variant="flat"
                      >
                        <v-tooltip
                          activator="parent"
                          location="top"
                          :text="$t('pages.outbounds')"
                        >
                          <span
                            style="font-weight: bold;"
                            v-text="$t('pages.outbounds')"
                          /><br>
                          <span
                            v-for="o in Data().onlines.outbound"
                            :key="o"
                          >{{ o }}<br></span>
                        </v-tooltip>
                        {{ Data().onlines.outbound?.length }}
                      </v-chip>
                    </template>
                  </v-col>
                </v-row>
              </template>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
import HttpUtils from '@/plugins/httputil'
import { HumanReadable } from '@/plugins/utils'
import Data from '@/store/modules/data'
import Update from '@/store/modules/update'
import Gauge from '@/components/tiles/Gauge.vue'
import History from '@/components/tiles/History.vue'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { i18n, locale } from '@/locales'
import LogVue from '@/layouts/modals/Logs.vue'
import Backup from '@/layouts/modals/Backup.vue'
import UsageStats from '@/layouts/modals/UsageStats.vue'

const loading = ref(false)
const menu = ref(false)
const menuItems = [
  { title: i18n.global.t('main.gauges'), value: [
    { title: i18n.global.t('main.gauge.cpu'), value: "g-cpu" },
    { title: i18n.global.t('main.gauge.mem'), value: "g-mem" },
    { title: i18n.global.t('main.gauge.dsk'), value: "g-dsk" },
    { title: i18n.global.t('main.gauge.swp'), value: "g-swp" },
    ]
  },
  { title: i18n.global.t('main.charts'), value: [
    { title: i18n.global.t('main.chart.cpu'), value: "h-cpu" },
    { title: i18n.global.t('main.chart.mem'), value: "h-mem" },
    { title: i18n.global.t('main.chart.net'), value: "h-net" },
    { title: i18n.global.t('main.chart.pnet'), value: "hp-net" },
    { title: i18n.global.t('main.chart.dio'), value: "h-dio" },
    ]
  },
  { title: i18n.global.t('main.infos'), value: [
    { title: i18n.global.t('main.info.sys'), value: "i-sys" },
    { title: i18n.global.t('main.info.sbd'), value: "i-sbd" },
    ]
  },
]

// What api/status answers with. The panel asks for only the sections it is
// showing, so a section is absent until its first reload; every reader below
// and in the template guards with ?. before it reaches in.
interface Usage {
  current: number
  total: number
}

interface NetUsage {
  recv: number
  sent: number
  precv: number
  psent: number
}

interface DiskIo {
  read: number
  write: number
}

interface SysInfo {
  hostName: string
  cpuType: string
  cpuCount: number
  ipv4: string[]
  ipv6: string[]
  appVersion: string
  bootTime: number
}

interface SingBoxInfo {
  running: boolean
  maintenance: boolean
  stats: {
    Alloc: number
    NumGoroutine: number
    Uptime: number
  }
}

interface TilesData {
  cpu: number
  mem: Usage
  dsk: Usage
  swp: Usage
  net: NetUsage
  dio: DiskIo
  sys: SysInfo
  sbd: SingBoxInfo
}

const tilesData = ref(<TilesData>{})
const update = Update()
let mounted = false
const canLoadData = () => mounted && !update.pauseDataPolling

const reloadItems = computed({
  get() { return Data().reloadItems },
  set(v:string[]) {
    if (Data().reloadItems.length == 0 && v.length>0) startTimer()
    if (Data().reloadItems.length > 0 && v.length == 0) stopTimer()
    Data().reloadItems = v
    if (v.length > 0) localStorage.setItem("reloadItems", v.join(','))
    else localStorage.removeItem("reloadItems")
  }
})

const selectedGauges = computed(() => menuItems[0].value.filter(item => reloadItems.value.includes(item.value)))
const otherTiles = computed(() => reloadItems.value.filter(item => !item.startsWith('g-')))

const reloadData = async () => {
  if (!canLoadData()) return
  const request = [...new Set(reloadItems.value.map(r => r.split('-')[1]))]
  if (tilesData.value?.sys?.appVersion) request.filter(r => r != 'sys')
  const data = await HttpUtils.get<TilesData>('api/status', { r: request.join(',') }, {}, canLoadData)
  if (data.success) {
    tilesData.value = data.obj
  }
}

const reloadSys = async () => {
  if (!canLoadData()) return
  const data = await HttpUtils.get<TilesData>('api/status', { r: 'sys' }, {}, canLoadData)
  if (data.success) {
    tilesData.value.sys = data.obj.sys
  }
}

let intervalId: ReturnType<typeof setInterval> | null = null

const startTimer = () => {
  if (!mounted || intervalId) return
  intervalId = setInterval(() => {
    reloadData()
  }, 2000)
}

const stopTimer = () => {
  if (intervalId) {
    clearInterval(intervalId)
    intervalId = null
  }
}

onMounted(async () => {
  mounted = true
  loading.value = true
  if (Data().reloadItems.length != 0) {
    await reloadData()
    startTimer()
  }
  loading.value = false
})

onBeforeUnmount(() => {
  mounted = false
  stopTimer()
})

const logModal = ref({ visible: false })

const backupModal = ref({ visible: false })

const usageStatsModal = ref({ visible: false })

const restartSingbox = async () => {
  loading.value = true
  await HttpUtils.post('api/restartSb',{})
  loading.value = false
}
</script>

<style scoped>
.dashboard {
  max-width: 100%;
  padding: 0;
}

.dashboard-header,
.dashboard-heading,
.dashboard-actions {
  display: flex;
  align-items: center;
}

.dashboard-header {
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 28px;
  flex-wrap: wrap;
}

.dashboard-heading {
  gap: 14px;
  min-width: 0;
}

.dashboard-eyebrow { display: block; margin-bottom: 8px; font-size: .6875rem; font-weight: 600; letter-spacing: .06em; color: var(--panel-muted); }

.dashboard-heading h1 {
  font-size: 2rem;
  font-weight: 600;
  line-height: 1.3;
  letter-spacing: -.025em;
}

.dashboard-heading p {
  color: var(--panel-muted);
  font-size: .8125rem;
  margin-top: 8px;
  overflow-wrap: anywhere;
}

.dashboard-actions {
  flex-wrap: wrap;
  gap: 8px;
}

.dashboard-actions :deep(.v-btn) {
  border-radius: 8px;
  font-size: .8125rem;
  letter-spacing: 0;
}

.dashboard-actions .dashboard-customize {
  background: rgb(var(--v-theme-primary));
  color: rgb(var(--v-theme-on-primary));
}

.dashboard-grid {
  margin: -8px;
}

.dashboard-empty {
  padding: 64px 24px;
  text-align: center;
  border: 1px dashed var(--panel-border);
  border-radius: var(--panel-radius);
}

.dashboard-empty h2 {
  margin-top: 16px;
  font-size: 1.125rem;
  font-weight: 600;
}

.dashboard-empty p {
  margin: 8px 0 24px;
  color: var(--panel-muted);
  font-size: .875rem;
}

.dashboard-grid > [class*="v-col"] {
  padding: 8px;
}

.dashboard-tile,
.dashboard-gauges {
  border: 1px solid var(--panel-border);
  border-radius: var(--panel-radius);
}

.dashboard-gauges-content {
  display: grid;
  grid-template-columns: repeat(var(--gauge-count), minmax(0, 1fr));
  row-gap: 24px;
  padding: 28px 8px;
}

.dashboard-gauge {
  min-width: 0;
  padding-inline: 24px;
}

.dashboard-gauge + .dashboard-gauge {
  border-inline-start: 1px solid var(--panel-border);
}

.dashboard-gauge-title {
  margin-bottom: 18px;
  color: var(--panel-muted);
  text-align: start;
  font-size: .75rem;
  font-weight: 500;
  line-height: 1.5;
}

.dashboard-tile {
  height: 268px;
}

.dashboard-tile-title {
  display: flex;
  align-items: center;
  padding: 20px 20px 12px;
  font-size: .875rem;
  font-weight: 600;
  line-height: 1.5;
}

.dashboard-tile-content {
  height: 202px;
  padding: 0 20px 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dashboard-tile--info .dashboard-tile-content {
  display: block;
  height: auto;
  min-height: 202px;
}

.dashboard-tile--info {
  height: auto;
  min-height: 268px;
}

.dashboard-info {
  display: grid;
  grid-template-columns: 72px minmax(0, 1fr);
  gap: 0;
  margin: 0 -4px;
  align-items: center;
  font-size: .8125rem;
}

.dashboard-info > [class*="v-col"] {
  width: 100%;
  max-width: 100%;
  min-width: 0;
  padding: 5px 4px;
}

.dashboard-info > [class*="v-col"]:nth-child(odd) {
  color: var(--panel-muted);
}

.dashboard-info :deep(.v-chip) {
  margin-inline-end: 4px;
  font-size: .75rem;
}

.dashboard-host {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 1199px) {
  .dashboard-gauges-content {
    grid-template-columns: repeat(var(--gauge-small-count), minmax(0, 1fr));
  }

  .dashboard-gauge:nth-child(odd) {
    border-inline-start: 0;
  }
}

@media (max-width: 599px) {
  .dashboard-gauge { padding-inline: 12px; }
  .dashboard-heading h1 { font-size: 1.6rem; }
  .dashboard-header {
    gap: 18px;
    margin-bottom: 22px;
  }

  .dashboard-actions {
    width: 100%;
  }

  .dashboard-actions :deep(.v-btn) {
    flex: 1 1 calc(50% - 8px);
    padding-inline: 10px;
  }
}
</style>
