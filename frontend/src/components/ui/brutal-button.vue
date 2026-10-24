<template>
  <button
    :type="type"
    class="brutal-button"
    :class="{ 'brutal-button--bordered': hasBorder, 'brutal-button--shadow': hasShadow }"
    :style="customStyles"
  >
    <slot />
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { CSSProperties } from 'vue'

const props = withDefaults(defineProps<{
  color?: string
  textColor?: string
  hasBorder?: boolean
  borderColor?: string
  hasShadow?: boolean
  shadowColor?: string
  radius?: number
  type?: 'button' | 'submit' | 'reset'
}>(), {
  color: undefined,
  textColor: undefined,
  borderColor: undefined,
  shadowColor: undefined,
  hasBorder: true,
  hasShadow: true,
  radius: 0,
  type: 'button',
})

const customStyles = computed((): CSSProperties => ({
  '--btn-bg': props.color || 'var(--background)',
  '--btn-text': props.textColor || 'var(--foreground)',
  '--btn-border': props.borderColor || 'var(--foreground)',
  '--btn-shadow': props.shadowColor || 'var(--foreground)',
  '--btn-radius': `${props.radius}px`,
}))
</script>

<style scoped>
.brutal-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  min-height: 48px;
  padding: 12px 24px;
  border: 0 solid var(--btn-border);
  border-radius: var(--btn-radius);
  background: var(--btn-bg);
  color: var(--btn-text);
  font: inherit;
  font-weight: 700;
  line-height: 1.5;
  cursor: pointer;
}
.brutal-button--bordered { border-width: 2px; }
.brutal-button--shadow { box-shadow: 4px 4px 0 var(--btn-shadow); }
.brutal-button:focus-visible { outline: 2px solid var(--ring); outline-offset: 6px; }
.brutal-button:disabled { cursor: wait; opacity: .65; }
</style>
