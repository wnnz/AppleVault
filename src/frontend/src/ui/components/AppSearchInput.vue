<template>
  <div :class="['app-search-input', `app-search-input--${size}`]">
    <svg class="app-search-input__icon" viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="11" cy="11" r="8" />
      <line x1="21" y1="21" x2="16.65" y2="16.65" />
    </svg>
    <AppInput
      :model-value="modelValue"
      :size="size"
      bare
      :placeholder="placeholder"
      @update:model-value="emit('update:modelValue', $event)"
      @keydown="$emit('keydown', $event)"
    />
    <button
      v-if="clearable && modelValue"
      type="button"
      class="app-search-input__clear"
      title="清空搜索"
      aria-label="清空搜索"
      @click="emit('update:modelValue', '')"
    >
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <line x1="18" y1="6" x2="6" y2="18" />
        <line x1="6" y1="6" x2="18" y2="18" />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import AppInput from './AppInput.vue'
import type { ControlSize } from '../control'

withDefaults(defineProps<{
  modelValue?: string
  placeholder?: string
  size?: ControlSize
  clearable?: boolean
}>(), {
  modelValue: '',
  placeholder: '',
  size: 'medium',
  clearable: true
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  keydown: [event: KeyboardEvent]
}>()
</script>

<style>
.app-search-input {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 0 10px;
  border: 1px solid var(--ui-border);
  border-radius: 8px;
  background: var(--ui-control-bg);
  box-sizing: border-box;
  transition: border-color .15s ease, background-color .15s ease, box-shadow .15s ease;
}
.app-search-input--small { min-height: var(--ui-control-height-small); height: var(--ui-control-height-small); }
.app-search-input--medium { min-height: var(--ui-control-height-medium); height: var(--ui-control-height-medium); }
.app-search-input--large { min-height: var(--ui-control-height-large); height: var(--ui-control-height-large); }
.app-search-input:focus-within { border-color: var(--ui-primary); box-shadow: 0 0 0 2px var(--ui-focus); }
.app-search-input__icon { width: 15px; height: 15px; flex: 0 0 15px; margin-right: 8px; fill: none; stroke: var(--ui-muted); stroke-width: 2; }
.app-search-input__clear { display: inline-flex; align-items: center; justify-content: center; width: 20px; height: 20px; flex: 0 0 20px; padding: 0; border: 0; border-radius: 4px; background: transparent; color: var(--ui-muted); cursor: pointer; }
.app-search-input__clear:hover { background: var(--ui-control-hover); color: var(--ui-text); }
.app-search-input__clear svg { width: 13px; height: 13px; fill: none; stroke: currentColor; stroke-width: 2.2; stroke-linecap: round; }
.theme-handdrawn .app-search-input {
  border: 1px solid transparent;
  border-image: url('../themes/handdrawn/assets/sketch-frame-blue.svg') 24 / 7px stretch;
  border-radius: 0;
  background-color: rgba(255, 255, 255, .36);
  box-shadow: none;
}
.theme-handdrawn .app-search-input:focus-within { border-color: transparent; background-color: transparent; box-shadow: none; }
</style>
