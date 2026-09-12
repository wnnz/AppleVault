<template>
  <div :class="['app-input-group', `app-input-group--${size}`]">
    <span v-if="prefix" class="app-input-group__prefix">{{ prefix }}</span>
    <AppInput
      :model-value="modelValue"
      :size="size"
      bare
      :placeholder="placeholder"
      @update:model-value="emit('update:modelValue', $event)"
      @keydown="$emit('keydown', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import AppInput from './AppInput.vue'
import type { ControlSize } from '../control'

withDefaults(defineProps<{
  modelValue?: string
  prefix?: string
  placeholder?: string
  size?: ControlSize
}>(), {
  modelValue: '',
  prefix: '',
  placeholder: '',
  size: 'medium'
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  keydown: [event: KeyboardEvent]
}>()
</script>

<style>
.app-input-group {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 0 8px;
  border: 1px solid var(--ui-border);
  border-radius: 8px;
  background: var(--ui-control-bg);
  box-sizing: border-box;
  transition: border-color .15s ease, background-color .15s ease, box-shadow .15s ease;
}
.app-input-group--small { min-height: var(--ui-control-height-small); height: var(--ui-control-height-small); }
.app-input-group--medium { min-height: var(--ui-control-height-medium); height: var(--ui-control-height-medium); }
.app-input-group--large { min-height: var(--ui-control-height-large); height: var(--ui-control-height-large); }
.app-input-group:focus-within { border-color: var(--ui-primary); box-shadow: 0 0 0 2px var(--ui-focus); }
.app-input-group__prefix { margin-right: 8px; padding-right: 8px; border-right: 1px solid var(--ui-border); color: var(--ui-muted); font-size: 12px; font-weight: 600; white-space: nowrap; }
.theme-handdrawn .app-input-group {
  border: 1px solid transparent;
  border-image: url('../themes/handdrawn/assets/sketch-frame-blue.svg') 24 / 7px stretch;
  border-radius: 0;
  background-color: rgba(255, 255, 255, .36);
  box-shadow: none;
}
.theme-handdrawn .app-input-group:focus-within { border-color: transparent; background-color: transparent; box-shadow: none; }
</style>
