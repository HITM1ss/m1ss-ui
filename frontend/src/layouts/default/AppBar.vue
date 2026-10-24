<template>
  <v-app-bar
    class="panel-appbar"
    :elevation="0"
    height="64"
  >
    <v-btn
      :icon="isMobile ? 'mdi-menu' : rail ? 'mdi-menu-open' : 'mdi-backburger'"
      variant="text"
      size="44"
      class="ms-3"
      :aria-label="$t('ui.toggleNavigation')"
      :aria-expanded="isMobile ? displayDrawer : !rail"
      aria-controls="panel-navigation"
      @click="$emit('toggleDrawer')"
    />
    <v-app-bar-title class="panel-appbar__title">
      <span class="panel-appbar__brand">M1ss-UI <span class="mx-3">/</span></span>
      {{ $t(String(route.name ?? 'pages.home')) }}
    </v-app-bar-title>
    <!-- A stopped core is invisible from the panel otherwise, and the flag
         survives a reboot, so it has to be said on every page. -->
    <v-chip
      v-if="maintenance"
      v-tooltip="$t('setting.maintenanceOnHint')"
      color="warning"
      variant="flat"
      density="comfortable"
      prepend-icon="mdi-wrench"
      class="mr-2"
    >
      {{ $t('setting.maintenance') }}
    </v-chip>
    <v-btn
      v-tooltip="$t('donate')"
      icon
      variant="text"
      href="https://donate.alireza0.dev"
      target="_blank"
      rel="noopener"
      class="donate-btn"
      :aria-label="$t('donate')"
    >
      <v-icon
        icon="mdi-heart"
        color="primary"
        size="1.5em"
      />
    </v-btn>
    <v-menu>
      <template #activator="{ props }">
        <v-btn
          icon
          size="44"
          v-bind="props"
          :aria-label="$t('ui.language')"
        >
          <v-icon>mdi-translate</v-icon>
        </v-btn>
      </template>
      <v-list>
        <v-list-item
          v-for="lang in languages"
          :key="lang.value"
          :active="isActiveLocale(lang.value)"
          @click="changeLocale(lang.value)"
        >
          <v-list-item-title>{{ lang.title }}</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
    <v-menu>
      <template #activator="{ props }">
        <v-btn
          icon
          size="44"
          class="me-4 ms-1"
          v-bind="props"
          :aria-label="$t('ui.appearance')"
        >
          <v-icon>mdi-theme-light-dark</v-icon>
        </v-btn>
      </template>
      <v-list>
        <v-list-item
          v-for="th in themes"
          :key="th.value"
          :prepend-icon="th.icon"
          :active="isActiveTheme(th.value)"
          @click="changeTheme(th.value)"
        >
          <v-list-item-title>{{ $t(`theme.${th.value}`) }}</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
  </v-app-bar>
</template>

<script lang="ts" setup>
import { useLocale } from 'vuetify'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { languages } from '@/locales'
import { useThemeSwitcher } from '@/composables/useThemeSwitcher'
import { computed } from 'vue'
import Data from '@/store/modules/data'

defineProps<{ isMobile: boolean, rail: boolean, displayDrawer: boolean }>()
defineEmits<{ toggleDrawer: [] }>()

const route = useRoute()
const { locale: i18nLocale } = useI18n()
const vuetifyLocale = useLocale()
const { themes, changeTheme, isActiveTheme } = useThemeSwitcher()

const changeLocale = (l: string) => {
  i18nLocale.value = l
  vuetifyLocale.current.value = l
  localStorage.setItem('locale', l)
  window.location.reload()
}
const isActiveLocale = (l: string) => i18nLocale.value === l
const maintenance = computed((): boolean => Data().maintenance)
</script>

<style scoped>
.panel-appbar { border-bottom: 2px solid var(--panel-border); }
.panel-appbar__title { font-size: .95rem; font-weight: 800; }
.panel-appbar__brand { color: var(--panel-muted); font-size: .8rem; font-weight: 500; }
.panel-appbar__brand > span { opacity: .4; }
.donate-btn { width: 36px; height: 36px; margin-inline-end: 4px; }
@media (max-width: 600px) {
  .panel-appbar__brand { display: none; }
  .panel-appbar__title { font-size: .9rem; margin-inline-start: 8px; }
  .donate-btn { display: none; }
}
</style>
