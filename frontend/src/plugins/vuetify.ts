/**
 * plugins/vuetify.ts
 *
 * Framework documentation: https://vuetifyjs.com`
 */

// Styles
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles/main.css'

import colors from 'vuetify/util/colors'
import { fa, en, vi, zhHans, zhHant, ru } from 'vuetify/locale'

// Composables
import { createVuetify } from 'vuetify'

// https://vuetifyjs.com/en/introduction/why-vuetify/#feature-guides
export default createVuetify({
  defaults: {
    VRow: { density: 'compact' },
    VBtn: { rounded: 'lg', elevation: 0 },
    VCard: { rounded: 'xl', elevation: 0 },
    VChip: { rounded: 'pill' },
    VTextField: {
      variant: 'solo-filled',
    },
    VSelect: {
      variant: 'solo-filled',
    },
    VCombobox: {
      variant: 'solo-filled',
    },
    VTextarea: {
      variant: 'solo-filled',
    },
  },
  theme: {
    defaultTheme: localStorage.getItem('theme') ?? 'system',
    themes: {
      light: {
        colors: {
          error: colors.red.accent2,
          primary: '#5965e8',
          secondary: '#8b5cf6',
          background: '#f5f6fc',
          surface: '#ffffff',
          'surface-variant': '#f7f8fd',
          'on-surface-variant': '#4e5370',
          'on-surface': '#24283e',
          info: '#5965e8',
          success: '#20856e',
          warning: '#a56c13',
        },
      },
      dark: {
        colors: {
          primary: '#a1a9ff',
          secondary: '#bc9bff',
          background: '#111321',
          surface: '#1b1e31',
          'surface-variant': '#22263c',
          'on-surface-variant': '#bdc2df',
          'on-surface': '#e7e9f8',
          info: '#a1a9ff',
          success: '#67ccb0',
          warning: '#e4b363',
          error: colors.red.accent2,
        },
      },
    },
  },
  locale: {
    locale: localStorage.getItem("locale") ?? 'en',
    fallback: 'en',
    messages: { en, fa, vi, zhHans, zhHant, ru },
  },
})
