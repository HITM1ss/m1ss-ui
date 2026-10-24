/**
 * main.ts
 *
 * Bootstraps Vuetify and other plugins then mounts the App`
 */

// Composables
import { createApp, ref } from 'vue'

// Components
import App from './App.vue'
import './styles/panel.css'

// Use router
import router from './router'

// Store
import store from './store'

// Plugins
import { registerPlugins } from '@/plugins'

// Locale
import { i18n } from '@/locales'
import Vue3PersianDatetimePicker from 'vue3-persian-datetime-picker'

// Notivue
import { createNotivue } from 'notivue'
import 'notivue/notification.css'
const notivue = createNotivue({
  animations: { enter: '', leave: '', clearAll: '' },
  transition: 'transform 0s linear',
  position: 'bottom-center',
  limit: 4,
  enqueue: false,
  avoidDuplicates: true,
  notifications: {
    global: {
      duration: 3000
    }
  },
})

const loading = ref(false)

const app = createApp(App)
app.provide('loading', loading)

registerPlugins(app)

app
  .use(router)
  .use(store)
  .use(i18n)
  .use(notivue)
  .component('DatePicker', Vue3PersianDatetimePicker)
  .mount('#app')
