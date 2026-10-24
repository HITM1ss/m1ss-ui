<template>
  <v-text-field
    :id="elId"
    v-model="dateFormatted"
    :label="label || $t('date.expiry')"
    prepend-inner-icon="mdi-calendar"
    readonly
    hide-details
  />
  <DatePicker
    v-model="Input"
    color="var(--primary)"
    :locale="locale"
    :element="elId"
    compact-time
    type="datetime"
    @input="Input=$event"
  >
    <template #next-month>
      <v-icon icon="mdi-chevron-right" />
    </template>
    <template #prev-month>
      <v-icon icon="mdi-chevron-left" />
    </template>
    <template #submit-btn="{ submit: submitPicker, canSubmit }">
      <v-btn
        :disabled="!canSubmit"
        @click="submitPicker"
      >
        {{ $t('submit') }}
      </v-btn>
    </template>
    <template #cancel-btn="{ vm }">
      <v-btn
        @click="reset(vm)"
      >
        {{ $t('reset') }}
      </v-btn>
    </template>
    <template #now-btn="{ goToday }">
      <v-btn
        @click="goToday"
      >
        {{ $t('now') }}
      </v-btn>
    </template>
  </DatePicker>
</template>

<script lang="ts">
import DatePicker from 'vue3-persian-datetime-picker'
import { i18n, locale } from '@/locales'
import 'moment/locale/ru'
import 'moment/locale/vi'
import 'moment/locale/zh-cn'
import 'moment/locale/zh-tw'

// The picker hands its own view-model to the cancel slot; closing the picker is
// all this component needs from it.
interface PickerVm {
  visible: boolean
}

export default {
  components: { DatePicker },
  props: {
    // A unix timestamp in seconds; 0 means no expiry.
    expiry: { type: Number, required: true },
    label: { type: String, required: false, default: undefined },
    inputId: { type: String, required: false, default: undefined }
  },
  emits: ['submit'],
  data() {
    return {
      menu: false,
      input: new Date(),
    }
  },
  computed: {
    locale() {
      return locale
    },
    elId() {
      return this.inputId || 'expiry'
    },
    dateFormatted() {
      if (this.expDate == 0) return i18n.global.t('unlimited')
      const date = new Date(this.expDate*1000)
      return date.toLocaleString(locale)
    },
    expDate() {
      return parseInt(String(this.expiry?? 0))
    },
    Input: {
      get() { return this.expDate == 0 ? new Date() : new Date(this.expDate*1000) },
      set(v:string) {
        this.input = new Date(v)
        this.submit()
      }
    }
  },
  watch: {
    menu(v) {
      if (v) {
        this.input = this.expiry == 0 ? new Date() : new Date(this.expDate*1000)
      }
    }
  },
  methods: {
    updateInput(v:Date) {
      this.input = v
    },
    setNow() {
      this.input = new Date()
    },
    submit() {
      this.$emit('submit',Math.floor(this.input.getTime()/1000))
    },
    reset(vm:PickerVm) {
      this.$emit('submit',0)
      this.input = new Date()
      vm.visible = false
    }
  }
}
</script>

<style>
.vpd-addon-list,
.vpd-addon-list-item {
  background-color: var(--popover) !important;
  border-color: var(--border) !important;
  color: var(--popover-foreground);
}
.vpd-content {
  background-color: var(--popover) !important;
  color: var(--popover-foreground);
  border: 2px solid var(--foreground);
  border-radius: 0;
  overflow: hidden;
  box-shadow: 6px 6px 0 var(--foreground);
}
.vpd-header { color: var(--primary-foreground); border-bottom: 2px solid var(--foreground); }
.vpd-day-effect { border-radius: 0; }
.vpd-week { color: var(--muted-foreground); }
.vpd-addon-list-item.vpd-selected,
.vpd-addon-list-item:hover {
  background-color: var(--accent) !important;
  color: var(--accent-foreground) !important;
}
.vpd-close-addon {
  color: var(--popover-foreground) !important;
  background-color: transparent;
}
.vpd-controls {
  overflow-x: hidden;
}
.vpd-month-label {
  width: auto;
}
.vpd-actions button:hover {
  background-color: transparent;
}
.vpd-wrapper[data-type=datetime].vpd-compact-time .vpd-time {
  border-top: 0;
}
.vpd-time .vpd-time-h .vpd-counter-item,
.vpd-time .vpd-time-m .vpd-counter-item {
  vertical-align: top;
}
</style>
