<template>
  <v-app class="panel-app">
    <a
      class="panel-skip-link"
      href="#panel-main"
    >{{ $t('ui.skipToContent') }}</a>
    <drawer
      :is-mobile="isMobile"
      :display-drawer="displayDrawer"
      :rail="rail"
      @toggle-drawer="toggleDrawer"
    />
    <default-bar
      :is-mobile="isMobile"
      :rail="rail"
      :display-drawer="displayDrawer"
      @toggle-drawer="toggleDrawer"
    />
    <default-view />
  </v-app>
</template>

<script lang="ts" setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import DefaultBar from './AppBar.vue'
import Drawer from './Drawer.vue'
import DefaultView from './View.vue'
import { useDisplay } from 'vuetify'
import Data from '@/store/modules/data'
import Update from '@/store/modules/update'

const update = Update()
let mounted = false
let dataTimer: ReturnType<typeof setInterval> | undefined

const loadData = () => {
  if (!update.pauseDataPolling) void Data().loadData(() => mounted && !update.pauseDataPolling)
}

onMounted(async () => {
  mounted = true
  await update.start()
  if (!mounted) return
  loadData()
  dataTimer = setInterval(loadData, 10000)
})

watch(() => update.pauseDataPolling, (paused) => {
  if (mounted && dataTimer !== undefined && !paused) loadData()
})

onBeforeUnmount(() => {
  mounted = false
  clearInterval(dataTimer)
  update.stop()
})

const { smAndDown } = useDisplay()
const displayDrawer = ref(false)
const rail = ref(localStorage.getItem('navigation-rail') === '1')

const toggleDrawer = () => {
  if (isMobile.value) {
    displayDrawer.value = !displayDrawer.value
  } else {
    rail.value = !rail.value
    localStorage.setItem('navigation-rail', rail.value ? '1' : '0')
  }
}

const isMobile = computed((): boolean => smAndDown.value)

// keep the drawer open on desktop and closed on mobile
watch(smAndDown, (v) => { displayDrawer.value = !v }, { immediate: true })
</script>

<style>
.panel-app .v-main { min-width: 0; }
</style>
