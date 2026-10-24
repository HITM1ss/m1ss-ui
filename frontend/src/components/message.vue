<template>
  <Notivue v-slot="item">
    <NotivueSwipe :item="item">
      <Notification
        :item="item"
        :theme="theme"
        :dir="direction"
        :icons="outlinedIcons"
        :hide-close="true"
        @click="item.clear"
      />
    </NotivueSwipe>
  </Notivue>
</template>

<script lang="ts" setup>
import { Notivue, Notification, NotivueSwipe, outlinedIcons, pastelTheme } from 'notivue'
import { computed } from 'vue'
import { useTheme } from 'vuetify'
import vuetify from '@/plugins/vuetify'

const Theme = useTheme()

const theme = computed(() =>{
  const colors = Theme.current.value.colors
  const ink = String(colors['on-background'])
  return {
    ...pastelTheme,
    '--nv-width': 'auto',
    '--nv-radius': '0px',
    '--nv-border-width': '2px',
    '--nv-shadow': `4px 4px 0 ${ink}`,
    '--nv-global-bg': String(colors.popover),
    '--nv-global-fg': String(colors['on-popover']),
    '--nv-global-border': ink,
    '--nv-global-accent': String(colors.primary),
    ...Object.fromEntries(['success', 'error', 'warning', 'info', 'promise'].flatMap(type => [
      [`--nv-${type}-bg`, String(colors.popover)],
      [`--nv-${type}-fg`, String(colors['on-popover'])],
      [`--nv-${type}-border`, ink],
      [`--nv-${type}-accent`, String(colors[type === 'promise' ? 'primary' : type])],
    ])),
  }
})

const direction = computed(() => {
  return vuetify.locale.isRtl ? 'rtl' : 'ltr'
})
</script>

<style>
:root {
  --nv-z: 10020;
}

</style>
