/**
 * plugins/vuetify.ts
 *
 * Framework documentation: https://vuetifyjs.com`
 */

// Styles
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles/main.css'

import { fa, en, vi, zhHans, zhHant, ru } from 'vuetify/locale'

// Composables
import { createVuetify } from 'vuetify'

// https://vuetifyjs.com/en/introduction/why-vuetify/#feature-guides
export default createVuetify({
  defaults: {
    VRow: { density: 'compact' },
    VBtn: { rounded: 'lg', elevation: 0 },
    VCard: { rounded: 'lg', elevation: 0 },
    VChip: { rounded: 'sm' },
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
          error: '#b64b48',
          primary: '#29685e',
          secondary: '#586b65',
          background: '#f3f5f4',
          surface: '#ffffff',
          'surface-variant': '#edf1ef',
          'on-surface-variant': '#52615b',
          'on-surface': '#26332e',
          info: '#426c8a',
          success: '#33735a',
          warning: '#996514',
        },
      },
      dark: {
        colors: {
          primary: '#8fc6b8',
          secondary: '#acbdb5',
          background: '#151b18',
          surface: '#1d2621',
          'surface-variant': '#27312b',
          'on-surface-variant': '#afbeb5',
          'on-surface': '#e1e9e4',
          info: '#91b4cf',
          success: '#8cc5a5',
          warning: '#d9b270',
          error: '#ed9991',
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
