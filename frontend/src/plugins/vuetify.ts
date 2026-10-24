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
    global: { ripple: false, transition: false, noClickAnimation: true },
    VRow: { density: 'compact' },
    VBtn: { rounded: 0, elevation: 0 },
    VCard: { rounded: 0, elevation: 0 },
    VChip: { rounded: 0 },
    VSwitch: { inset: 'square', flat: true, color: 'primary' },
    VCheckbox: { color: 'primary' },
    VRadio: { color: 'primary' },
    // Tabs and floating labels also animate through the Web Animations API.
    VTab: { hideSlider: true },
    VTabs: { hideSlider: true },
    VTextField: {
      variant: 'outlined',
      active: true,
    },
    VSelect: {
      variant: 'outlined',
      active: true,
    },
    VCombobox: {
      variant: 'outlined',
      active: true,
    },
    VAutocomplete: {
      variant: 'outlined',
      active: true,
    },
    VTextarea: {
      variant: 'outlined',
      active: true,
    },
    VFileInput: { variant: 'outlined', active: true },
  },
  theme: {
    defaultTheme: localStorage.getItem('theme') ?? 'system',
    themes: {
      light: {
        colors: {
          primary: '#c96442',
          'on-primary': '#ffffff',
          secondary: '#e9e6dc',
          'on-secondary': '#535146',
          background: '#faf9f5',
          'on-background': '#3d3929',
          surface: '#faf9f5',
          'on-surface': '#141413',
          'surface-variant': '#ede9de',
          'on-surface-variant': '#83827d',
          'surface-bright': '#ffffff',
          'surface-light': '#e9e6dc',
          outline: '#dad9d4',
          popover: '#ffffff',
          'on-popover': '#28261b',
          'chart-1': '#b05730',
          'chart-2': '#9c87f5',
          info: '#426c8a',
          success: '#33735a',
          warning: '#996514',
          error: '#141413',
          'on-error': '#ffffff',
        },
        variables: { 'border-color': '#dad9d4', 'border-opacity': 1 },
      },
      dark: {
        colors: {
          primary: '#d97757',
          'on-primary': '#ffffff',
          secondary: '#faf9f5',
          'on-secondary': '#30302e',
          background: '#262624',
          'on-background': '#c3c0b6',
          surface: '#262624',
          'on-surface': '#faf9f5',
          'surface-variant': '#1b1b19',
          'on-surface-variant': '#b7b5a9',
          'surface-bright': '#30302e',
          'surface-light': '#1a1915',
          outline: '#3e3e38',
          popover: '#30302e',
          'on-popover': '#e5e5e2',
          'chart-1': '#b05730',
          'chart-2': '#9c87f5',
          info: '#91b4cf',
          success: '#8cc5a5',
          warning: '#d9b270',
          error: '#ef4444',
          'on-error': '#ffffff',
        },
        variables: { 'border-color': '#3e3e38', 'border-opacity': 1 },
      },
    },
  },
  locale: {
    locale: localStorage.getItem("locale") ?? 'en',
    fallback: 'en',
    messages: { en, fa, vi, zhHans, zhHant, ru },
  },
})
