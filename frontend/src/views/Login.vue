<template>
  <v-app class="login-app">
    <v-main class="login-main">
      <div class="login-page">
        <v-card
          class="login-card"
          elevation="0"
        >
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
.login-page {
  display: grid;
  min-height: 100svh;
  place-items: center;
  padding: 32px 20px;
  background:
    radial-gradient(ellipse at 15% 15%, rgba(89, 101, 232, .11), transparent 48%),
    radial-gradient(ellipse at 90% 85%, rgba(139, 92, 246, .09), transparent 45%),
    rgb(var(--v-theme-background));
}

.login-card.v-card {
  width: 100%;
  max-width: 440px;
  padding: 36px;
  background: rgb(var(--v-theme-surface));
  color: rgb(var(--v-theme-on-surface));
  border: 1px solid var(--panel-border, rgba(120, 130, 160, .16));
  border-radius: 28px;
  box-shadow: 0 16px 56px rgba(40, 46, 96, .07) !important;
}

.login-brand {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 32px;
}

.login-brand-mark {
  display: grid;
  width: 48px;
  height: 48px;
  place-items: center;
  border-radius: 24%;
  overflow: hidden;
  border: 0;
}

.login-brand-mark img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.login-brand-name {
  font-size: 24px;
  font-weight: 750;
  letter-spacing: -.06em;
}

.login-heading h1 {
  font-size: 28px;
  font-weight: 700;
  line-height: 1.4;
  letter-spacing: -.025em;
}

.login-heading p {
  margin-top: 8px;
  color: var(--panel-muted, rgba(var(--v-theme-on-surface), .65));
  font-size: 14px;
  line-height: 1.6;
}

.login-form {
  margin-top: 28px;
}

.login-form :deep(.v-field),
.login-language :deep(.v-field) {
  border-radius: 12px;
}

.login-form :deep(.v-field__prepend-inner),
.login-language :deep(.v-field__prepend-inner) {
  margin-inline-end: 8px;
  color: var(--panel-muted, rgba(var(--v-theme-on-surface), .65));
}

.login-submit.v-btn {
  margin-top: 8px;
  border-radius: 13px;
  background: var(--panel-gradient, linear-gradient(125deg, #5965e8, #8b5cf6)) !important;
  color: #fff !important;
  font-size: 14px;
  font-weight: 650;
  letter-spacing: .025em;
}

.login-preferences {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 28px;
  padding-top: 24px;
  border-top: 1px solid var(--panel-border, rgba(120, 130, 160, .16));
}

.login-language {
  min-width: 0;
}

.login-theme.v-btn {
  flex-shrink: 0;
  border-radius: 13px;
}

@media (max-width: 480px) {
  .login-page {
    padding: 24px 16px;
  }

  .login-card.v-card {
    padding: 28px 24px;
    border-radius: 24px;
  }
}
</style>
