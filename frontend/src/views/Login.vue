<template>
  <v-app class="login-app">
    <v-main class="login-main">
      <div class="login-page">
        <div class="login-workspace">
          <header class="login-masthead">
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
            <span class="login-masthead__label">{{ $t('ui.controlPanel') }}</span>
          </header>

          <div class="login-frame">
            <aside class="login-story">
              <span class="login-eyebrow">
                <span aria-hidden="true" />
                {{ $t('ui.controlPanel') }}
              </span>
              <div class="login-story__body">
                <h2>{{ $t('login.workspaceTitle') }}</h2>
                <p>{{ $t('login.workspaceDescription') }}</p>
              </div>
              <div class="login-index">
                <div>
                  <span>01</span>{{ $t('ui.connections') }}<v-icon
                    icon="mdi-arrow-top-right"
                    aria-hidden="true"
                  />
                </div>
                <div>
                  <span>02</span>{{ $t('network') }}<v-icon
                    icon="mdi-arrow-top-right"
                    aria-hidden="true"
                  />
                </div>
                <div>
                  <span>03</span>{{ $t('ui.system') }}<v-icon
                    icon="mdi-arrow-top-right"
                    aria-hidden="true"
                  />
                </div>
              </div>
            </aside>

            <section
              class="login-credentials"
              aria-labelledby="login-title"
            >
              <header class="login-heading">
                <span
                  class="login-section-number"
                  aria-hidden="true"
                >↳</span>
                <h1 id="login-title">
                  {{ $t('login.title') }}
                </h1>
                <p>{{ $t('login.subtitle') }}</p>
              </header>

              <v-form
                ref="form"
                class="login-form"
                @submit.prevent="login"
              >
                <v-text-field
                  id="login-username"
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
                  :disabled="loading"
                />
                <v-text-field
                  id="login-password"
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
                  :disabled="loading"
                />
                <BrutalButton
                  :disabled="loading"
                  :aria-busy="loading"
                  type="submit"
                  color="var(--primary)"
                  text-color="var(--primary-foreground)"
                  class="login-submit"
                >
                  {{ loading ? $t('loading') : $t('login.title') }}
                  <v-icon
                    icon="mdi-arrow-right"
                    size="18"
                    end
                  />
                </BrutalButton>
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
                <v-menu content-class="login-appearance-menu">
                  <template #activator="{ props }">
                    <BrutalButton
                      v-bind="props"
                      :aria-label="$t('ui.appearance')"
                      :title="$t('ui.appearance')"
                      class="login-theme"
                    >
                      <v-icon
                        icon="mdi-theme-light-dark"
                        aria-hidden="true"
                      />
                    </BrutalButton>
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
            </section>
          </div>
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
import BrutalButton from '@/components/ui/brutal-button.vue'


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
  if (loading.value || username.value == '' || password.value == '') return
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
.login-page { display: grid; min-height: 100dvh; place-items: center; padding: 48px 32px 60px; }
.login-workspace { width: 100%; max-width: 1040px; }
.login-masthead { display: flex; justify-content: space-between; align-items: center; gap: 24px; margin-bottom: 24px; }
.login-brand { display: flex; align-items: center; gap: 14px; }
.login-brand-mark { display: grid; width: 44px; height: 44px; place-items: center; border: 2px solid var(--foreground); overflow: hidden; }
.login-brand-mark img { display: block; width: 100%; height: 100%; object-fit: cover; }
.login-brand-name { font-size: 1.5rem; font-weight: 800; letter-spacing: -.055em; }
.login-masthead__label { font-size: .75rem; font-weight: 600; }
.login-frame { display: grid; grid-template-columns: 1fr 1fr; border: 2px solid var(--foreground); box-shadow: 10px 10px 0 var(--foreground); background: var(--card); }
.login-story { display: flex; flex-direction: column; min-width: 0; padding: 40px; background: var(--accent); border-inline-end: 2px solid var(--foreground); }
.login-eyebrow { display: inline-flex; align-items: center; align-self: flex-start; gap: 10px; border: 2px solid var(--foreground); padding: 8px 12px; background: var(--background); font-size: .6875rem; font-weight: 700; }
.login-eyebrow > span { width: 8px; height: 8px; background: var(--primary); }
.login-story__body { margin: 36px 0 40px; }
.login-story h2 { max-width: 9em; font-size: clamp(2.25rem, 3.7vw, 3.5rem); font-weight: 800; letter-spacing: -.06em; line-height: 1.22; white-space: pre-line; text-wrap: balance; }
.login-story p { margin-top: 24px; max-width: 30em; font-size: .875rem; line-height: 1.85; color: var(--foreground); text-wrap: pretty; }
.login-index { margin-top: auto; border-bottom: 1px solid var(--foreground); }
.login-index > div { display: flex; align-items: center; gap: 16px; padding: 14px 0; border-top: 1px solid var(--foreground); font-size: .8125rem; font-weight: 600; }
.login-index span { font-size: .6875rem; font-weight: 500; }
.login-index .v-icon { margin-inline-start: auto; font-size: 18px; }
.login-credentials { min-width: 0; padding: 40px; }
.login-heading { position: relative; padding-bottom: 24px; border-bottom: 2px solid var(--foreground); }
.login-section-number { position: absolute; inset-block-start: 0; inset-inline-end: 0; font-size: 2.5rem; line-height: 1; color: var(--primary); }
.login-heading h1 { padding-inline-end: 40px; font-size: 2rem; font-weight: 800; line-height: 1.25; letter-spacing: -.045em; }
.login-heading p { margin-top: 12px; color: var(--muted-foreground); font-size: .8125rem; line-height: 1.7; }
.login-form { margin-top: 28px; }
.login-form > .v-input { margin-bottom: 12px; }
.login-form :deep(.v-field), .login-language :deep(.v-field) { border-radius: 0; background: var(--background); }
.login-form :deep(.v-field__outline), .login-language :deep(.v-field__outline) { --v-field-border-opacity: 1; --v-field-border-width: 2px; color: var(--foreground); }
.login-form :deep(.v-field__prepend-inner), .login-language :deep(.v-field__prepend-inner) { margin-inline-end: 10px; color: var(--foreground); opacity: 1; }
.login-form :deep(.v-field .v-label), .login-language :deep(.v-field .v-label) { color: var(--foreground); font-weight: 600; }
.login-form :deep(.v-field--focused:not(.v-field--error)), .login-language :deep(.v-field--focused) { outline-offset: 3px; }
.login-submit { width: calc(100% - 4px); min-height: 52px; margin-top: 8px; }
.login-submit .v-icon { margin-inline-start: auto; }
.login-preferences { display: flex; align-items: center; gap: 16px; margin-top: 36px; padding-top: 24px; border-top: 1px solid var(--foreground); }
.login-language { min-width: 0; }
.login-theme { flex-shrink: 0; width: 44px; height: 44px; padding: 0; min-height: 44px; margin-inline-end: 4px; }
:global(.login-appearance-menu > .v-list) { border: 2px solid var(--foreground) !important; border-radius: 0 !important; box-shadow: 4px 4px 0 var(--foreground) !important; }
@media (max-width: 900px) {
  .login-story, .login-credentials { padding: 32px; }
}
@media (max-width: 700px) {
  .login-page { padding: 28px 22px 40px 16px; }
  .login-workspace { max-width: 440px; }
  .login-masthead { margin-bottom: 20px; }
  .login-masthead__label { display: none; }
  .login-frame { grid-template-columns: 1fr; box-shadow: 6px 6px 0 var(--foreground); }
  .login-story { padding: 24px; border-inline-end: 0; border-bottom: 2px solid var(--foreground); }
  .login-eyebrow { padding: 6px 10px; }
  .login-story__body { margin: 20px 0 0; }
  .login-story h2 { max-width: none; font-size: 1.75rem; white-space: normal; }
  .login-story p, .login-index { display: none; }
  .login-credentials { padding: 28px 24px 32px; }
  .login-heading h1 { font-size: 1.75rem; }
  .login-preferences { gap: 12px; }
}
</style>
