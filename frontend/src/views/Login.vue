<template>
  <v-app class="login-app">
    <v-main class="login-main">
      <div class="login-page">
        <div class="login-shell">
          <aside class="login-intro">
            <div class="login-brand">
              <span class="login-brand-mark">
                <img
                  src="@/assets/logo.webp"
                  alt=""
                  width="48"
                  height="48"
                >
              </span>
              <span class="login-brand-name">M1ss-UI</span>
            </div>

            <div class="login-intro__body">
              <span class="login-eyebrow">{{ $t('ui.controlPanel') }}</span>
              <h2>{{ $t('login.workspaceTitle') }}</h2>
              <p>{{ $t('login.workspaceDescription') }}</p>
            </div>
            <div class="login-index">
              <div><span>01</span>{{ $t('ui.connections') }}</div>
              <div><span>02</span>{{ $t('network') }}</div>
              <div><span>03</span>{{ $t('ui.system') }}</div>
            </div>
          </aside>

          <v-card
            class="login-card"
            elevation="0"
          >
            <header class="login-heading">
              <h1>{{ $t('login.title') }}</h1>
              <p>{{ $t('login.subtitle') }}</p>
            </header>

            <v-form
              ref="form"
              class="login-form"
              @submit.prevent="login"
            >
              <v-text-field
                v-model="username"
                :label="$t('login.username')"
                :rules="usernameRules"
                prepend-inner-icon="mdi-account-outline"
                autocomplete="username"
                name="username"
                variant="outlined"
                density="comfortable"
                color="primary"
                required
              />
              <v-text-field
                v-model="password"
                :label="$t('login.password')"
                :rules="passwordRules"
                prepend-inner-icon="mdi-lock-outline"
                autocomplete="current-password"
                name="password"
                type="password"
                variant="outlined"
                density="comfortable"
                color="primary"
                required
              />
              <v-btn
                :loading="loading"
                type="submit"
                color="primary"
                block
                height="48"
                class="login-submit"
                elevation="0"
              >
                {{ $t('login.title') }}
                <v-icon
                  icon="mdi-arrow-right"
                  size="18"
                  end
                />
              </v-btn>
            </v-form>

            <div class="login-preferences">
              <v-select
                v-model="$i18n.locale"
                :aria-label="$t('ui.language')"
                :label="$t('ui.language')"
                :items="languages"
                prepend-inner-icon="mdi-translate"
                class="login-language"
                density="compact"
                hide-details
                variant="outlined"
                @update:model-value="changeLocale"
              />
              <v-menu>
                <template #activator="{ props }">
                  <v-btn
                    v-bind="props"
                    :aria-label="$t('ui.appearance')"
                    :title="$t('ui.appearance')"
                    class="login-theme"
                    icon="mdi-theme-light-dark"
                    variant="tonal"
                    color="primary"
                    size="44"
                  />
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
            </div>
          </v-card>
        </div>
      </div>
    </v-main>
  </v-app>
</template>

<script lang="ts" setup>
import { ref } from "vue"
import { useLocale } from 'vuetify'
import { i18n, languages } from '@/locales'
import { useRouter } from 'vue-router'
import HttpUtil from '@/plugins/httputil'
import { setAuthenticated } from '@/plugins/auth'
import { useThemeSwitcher } from '@/composables/useThemeSwitcher'


const locale = useLocale()
const { themes, changeTheme, isActiveTheme } = useThemeSwitcher()

const username = ref('')
const usernameRules = [
  (value: string) => {
    if (value?.length > 0) return true
    return i18n.global.t('login.unRules')
  },
]

const password = ref('')
const passwordRules = [
  (value: string) => {
    if (value?.length > 0) return true
    return i18n.global.t('login.pwRules')
  },
]

const loading = ref(false)
const router = useRouter()

const login = async () => {
  if (username.value == '' || password.value == '') return
  loading.value=true
  const response = await HttpUtil.post('api/login',{user: username.value, pass: password.value})
  if(response.success){
    setAuthenticated()
    setTimeout(() => {
      loading.value=false
      router.push('/')
    }, 500)
  } else {
    loading.value=false
  }
}
const changeLocale = (l: string | null) => {
  locale.current.value = l ?? 'en'
  localStorage.setItem('locale', locale.current.value)
}
</script>

<style scoped>
.login-page { display: grid; min-height: 100dvh; place-items: center; padding: 40px 24px; }
.login-shell { display: grid; grid-template-columns: 1.08fr 1fr; width: 100%; max-width: 960px; border: 1px solid var(--panel-border); border-radius: 20px; overflow: hidden; background: rgb(var(--v-theme-surface)); }
.login-intro { display: flex; flex-direction: column; padding: 40px; background: var(--panel-tint); border-inline-end: 1px solid var(--panel-border); }
.login-brand { display: flex; align-items: center; gap: 12px; }
.login-brand-mark { display: grid; width: 40px; height: 40px; place-items: center; border-radius: 10px; overflow: hidden; }
.login-brand-mark img { display: block; width: 100%; height: 100%; object-fit: cover; }
.login-brand-name { font-size: 1.25rem; font-weight: 600; letter-spacing: -.04em; }
.login-intro__body { margin: 68px 0 48px; }
.login-eyebrow { font-size: .6875rem; color: rgb(var(--v-theme-primary)); font-weight: 600; letter-spacing: .08em; }
.login-intro h2 { margin-top: 16px; font-size: 2.6rem; line-height: 1.3; font-weight: 600; letter-spacing: -.055em; white-space: pre-line; text-wrap: balance; }
.login-intro p { margin-top: 20px; max-width: 32em; font-size: .875rem; line-height: 1.9; color: var(--panel-muted); text-wrap: pretty; }
.login-index { margin-top: auto; }
.login-index > div { display: flex; align-items: center; gap: 16px; padding: 12px 0; border-top: 1px solid var(--panel-border); font-size: .75rem; }
.login-index span { color: var(--panel-muted); font-size: .625rem; }
.login-card.v-card { align-self: center; width: 100%; padding: 48px 40px; border-radius: 0; background: transparent; }
.login-heading h1 { font-size: 1.75rem; font-weight: 600; line-height: 1.4; letter-spacing: -.035em; }
.login-heading p { margin-top: 10px; color: var(--panel-muted); font-size: .8125rem; line-height: 1.8; }
.login-form { margin-top: 32px; }
.login-form :deep(.v-field__prepend-inner), .login-language :deep(.v-field__prepend-inner) { margin-inline-end: 8px; color: var(--panel-muted); }
.login-submit.v-btn { margin-top: 8px; border-radius: 8px; font-size: .875rem; }
.login-preferences { display: flex; align-items: center; gap: 12px; margin-top: 32px; padding-top: 24px; border-top: 1px solid var(--panel-border); }
.login-language { min-width: 0; }
.login-theme.v-btn { flex-shrink: 0; }
@media (max-width: 700px) {
  .login-page { padding: 20px 16px; }
  .login-shell { grid-template-columns: 1fr; max-width: 440px; }
  .login-intro { padding: 24px; border-inline-end: 0; border-bottom: 1px solid var(--panel-border); }
  .login-intro__body, .login-index { display: none; }
  .login-card.v-card { padding: 28px 24px 32px; }
  .login-heading h1 { font-size: 1.5rem; }
}
</style>
