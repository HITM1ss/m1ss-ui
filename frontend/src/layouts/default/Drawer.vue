<template>
  <v-navigation-drawer
    id="panel-navigation"
    v-model="showDrawer"
    class="panel-drawer"
    :temporary="isMobile"
    :rail="!isMobile && rail"
    :permanent="!isMobile"
    :width="232"
    :rail-width="72"
  >
    <div
      class="panel-brand"
      :class="{ 'panel-brand--rail': !isMobile && rail }"
    >
      <div class="panel-brand__mark">
        <img
          src="@/assets/logo.webp"
          alt=""
          width="40"
          height="40"
        >
      </div>
      <div
        v-if="isMobile || !rail"
        class="panel-brand__text"
      >
        <strong>M1ss-UI</strong>
        <span>{{ $t('ui.controlPanel') }}</span>
      </div>
      <v-btn
        v-if="isMobile"
        icon="mdi-close"
        size="small"
        variant="text"
        :aria-label="$t('actions.close')"
        @click="$emit('toggleDrawer')"
      />
    </div>
    <nav
      :aria-label="$t('ui.navigation')"
      class="panel-nav"
    >
      <div
        v-for="group in menuGroups"
        :key="group.title"
        class="panel-nav__group"
      >
        <div
          v-if="isMobile || !rail"
          class="panel-nav__label"
        >
          {{ $t(group.title) }}
        </div>
        <v-list
          density="compact"
          nav
        >
          <v-list-item
            v-for="item in group.items"
            :key="item.path"
            v-tooltip="!isMobile && rail ? $t(item.title) : undefined"
            :to="item.path"
            :active="router.currentRoute.value.path === item.path"
            :aria-label="$t(item.title)"
            :prepend-icon="item.icon"
            :title="$t(item.title)"
            @click="isMobile && $emit('toggleDrawer')"
          />
        </v-list>
      </div>
    </nav>
    <template #append>
      <div class="panel-drawer__footer">
        <v-list
          density="compact"
          nav
        >
          <v-list-item
            v-tooltip="!isMobile && rail ? $t('menu.logout') : undefined"
            prepend-icon="mdi-logout"
            :title="$t('menu.logout')"
            :aria-label="$t('menu.logout')"
            @click="logout()"
          />
        </v-list>
      </div>
    </template>
  </v-navigation-drawer>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import router from '@/router'
import { logout } from '@/plugins/httputil'

const props = defineProps<{ isMobile: boolean, displayDrawer: boolean, rail: boolean }>()
const emit = defineEmits<{ toggleDrawer: [] }>()

const showDrawer = computed({
  get: () => props.displayDrawer,
  set: (value: boolean) => { if (value !== props.displayDrawer) emit('toggleDrawer') },
})

const menuGroups = [
  { title: 'ui.overview', items: [
    { title: 'pages.home', icon: 'mdi-view-dashboard-outline', path: '/' },
  ] },
  { title: 'ui.connections', items: [
    { title: 'pages.inbounds', icon: 'mdi-cloud-download-outline', path: '/inbounds' },
    { title: 'pages.clients', icon: 'mdi-account-multiple-outline', path: '/clients' },
    { title: 'pages.outbounds', icon: 'mdi-cloud-upload-outline', path: '/outbounds' },
    { title: 'pages.endpoints', icon: 'mdi-access-point-network', path: '/endpoints' },
    { title: 'pages.services', icon: 'mdi-server-outline', path: '/services' },
  ] },
  { title: 'network', items: [
    { title: 'pages.tls', icon: 'mdi-shield-key-outline', path: '/tls' },
    { title: 'pages.basics', icon: 'mdi-tune-variant', path: '/basics' },
    { title: 'pages.rules', icon: 'mdi-routes', path: '/rules' },
    { title: 'pages.dns', icon: 'mdi-dns-outline', path: '/dns' },
  ] },
  { title: 'ui.system', items: [
    { title: 'pages.admins', icon: 'mdi-account-cog-outline', path: '/admins' },
    { title: 'pages.settings', icon: 'mdi-cog-outline', path: '/settings' },
  ] },
]
</script>

<style scoped>
.panel-drawer { border-inline-end: 1px solid var(--panel-border) !important; }
.panel-brand { display: flex; align-items: center; gap: 10px; min-height: 96px; padding: 24px 20px; }
.panel-brand--rail { justify-content: center; padding-inline: 12px; }
.panel-brand__mark { display: grid; place-items: center; flex: 0 0 40px; height: 40px; border-radius: 24%; overflow: hidden; border: 0; }
.panel-brand__mark img { display: block; width: 100%; height: 100%; object-fit: cover; }
.panel-brand__text { display: flex; flex: 1; flex-direction: column; gap: 1px; white-space: nowrap; }
.panel-brand__text strong { font-size: 1.15rem; letter-spacing: -.025em; }
.panel-brand__text span { font-size: .7rem; color: var(--panel-muted); }
.panel-nav { padding: 6px 12px 14px; }
.panel-nav__group + .panel-nav__group { margin-top: 16px; }
.panel-nav__label { padding: 0 14px 4px; color: var(--panel-muted); font-size: .65rem; font-weight: 600; letter-spacing: .045em; }
.panel-nav .v-list { --v-list-prepend-gap: 12px; padding: 0; background: transparent; }
.panel-nav :deep(.v-list-item) { min-height: 44px; margin-bottom: 3px; border-radius: 8px; }
.panel-nav :deep(.v-list-item__prepend > .v-icon) { opacity: .7; font-size: 20px; }
.panel-nav :deep(.v-list-item-title) { font-size: .82rem; font-weight: 500; }
.panel-nav :deep(.v-list-item--active) { background: var(--panel-tint); color: rgb(var(--v-theme-primary)); box-shadow: inset 3px 0 rgb(var(--v-theme-primary)); }
.panel-nav :deep(.v-list-item--active .v-list-item__overlay) { opacity: 0; }
.panel-nav :deep(.v-list-item--active .v-icon) { opacity: 1; }
.panel-drawer__footer { padding: 8px 12px; border-top: 1px solid var(--panel-border); }
.panel-drawer__footer .v-list { background: transparent; }
.panel-drawer__footer :deep(.v-list-item) { min-height: 44px; border-radius: 8px; color: var(--panel-muted); }
.panel-drawer__footer :deep(.v-list-item-title) { font-size: .82rem; }
.v-locale--is-rtl .panel-nav :deep(.v-list-item--active) { box-shadow: inset -3px 0 rgb(var(--v-theme-primary)); }
.v-navigation-drawer--rail .panel-nav__group + .panel-nav__group { border-top: 1px solid var(--panel-border); padding-top: 12px; margin-top: 12px; }
.v-navigation-drawer--rail .panel-nav { padding-inline: 12px; }
.v-navigation-drawer--rail :deep(.v-list) { --v-list-prepend-gap: 0px; }
.v-navigation-drawer--rail :deep(.v-list-item) { grid-template-columns: 1fr; grid-template-areas: "prepend"; justify-items: center; padding-inline: 0; }
.v-navigation-drawer--rail :deep(.v-list-item__content) { position: absolute; width: 1px; height: 1px; padding: 0; overflow: hidden; clip-path: inset(50%); white-space: nowrap; }
</style>
