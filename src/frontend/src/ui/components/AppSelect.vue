<template>
  <SelectRoot
    :model-value="modelValue as any"
    :disabled="disabled || loading"
    @update:model-value="emit('update:modelValue', $event as SelectValue)"
  >
    <SelectTrigger
      v-bind="$attrs"
      class="app-select"
      :class="[`app-select--${size}`, { 'app-select--disabled': disabled || loading }]"
      :aria-label="placeholder"
    >
      <SelectValue class="app-select__label" :placeholder="loading ? '加载中…' : placeholder" />
      <SelectIcon class="app-select__icon" aria-hidden="true" />
    </SelectTrigger>
    <SelectPortal>
      <SelectContent
        class="app-select-menu"
        position="popper"
        :side-offset="4"
        :collision-padding="8"
      >
        <SelectViewport class="app-select-menu__viewport">
          <SelectItem
            v-for="option in options"
            :key="String(option.value)"
            :value="option.value as any"
            :disabled="option.disabled"
            class="app-select-option"
          >
            <SelectItemText>{{ option.label }}</SelectItemText>
            <SelectItemIndicator class="app-select-option__indicator">✓</SelectItemIndicator>
          </SelectItem>
        </SelectViewport>
      </SelectContent>
    </SelectPortal>
  </SelectRoot>
</template>

<script setup lang="ts">
import {
  SelectContent,
  SelectIcon,
  SelectItem,
  SelectItemIndicator,
  SelectItemText,
  SelectPortal,
  SelectRoot,
  SelectTrigger,
  SelectValue,
  SelectViewport
} from 'reka-ui'
import type { ControlSize } from '../control'

defineOptions({ inheritAttrs: false })

export type SelectValue = string | number | null
export interface AppSelectOption {
  label: string
  value: string | number
  disabled?: boolean
}

withDefaults(defineProps<{
  modelValue?: SelectValue
  options: AppSelectOption[]
  placeholder?: string
  disabled?: boolean
  loading?: boolean
  size?: ControlSize
}>(), {
  modelValue: null,
  placeholder: '请选择',
  disabled: false,
  loading: false,
  size: 'medium'
})

const emit = defineEmits<{ 'update:modelValue': [value: SelectValue] }>()
</script>

<style>
.app-select {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  width: 100%;
  padding: 0 10px;
  border: 1px solid var(--ui-border);
  border-radius: 8px;
  background: var(--ui-control-bg);
  color: var(--ui-text);
  font: inherit;
  font-size: 13px;
  cursor: pointer;
  box-sizing: border-box;
  outline: none;
}
.app-select--small { min-height: var(--ui-control-height-small); height: var(--ui-control-height-small); }
.app-select--medium { min-height: var(--ui-control-height-medium); height: var(--ui-control-height-medium); }
.app-select--large { min-height: var(--ui-control-height-large); height: var(--ui-control-height-large); }
.app-select:hover { border-color: var(--ui-border); }
.app-select[data-state='open'] { border-color: var(--ui-primary); }
.app-select:focus-visible { box-shadow: 0 0 0 2px var(--ui-focus); }
.app-select:disabled { cursor: not-allowed; opacity: .5; }
.app-select__icon {
  display: block;
  width: 7px;
  height: 7px;
  flex: 0 0 7px;
  margin-right: 1px;
  border-right: 1.5px solid currentColor;
  border-bottom: 1.5px solid currentColor;
  color: var(--ui-muted);
  font-size: 0;
  line-height: 0;
  transform: translateY(-2px) rotate(45deg) !important;
}
.app-select-menu { z-index: 10050; min-width: var(--reka-select-trigger-width); max-height: min(300px, var(--reka-select-content-available-height)); overflow: hidden; border: 1px solid var(--ui-border); border-radius: 8px; background: var(--ui-surface); color: var(--ui-text); box-shadow: 0 10px 30px rgba(15, 23, 42, .16); }
.app-select-menu__viewport { padding: 4px; }
.app-select-option { position: relative; display: flex; align-items: center; min-height: 30px; padding: 0 28px 0 10px; border-radius: 6px; font-size: 13px; cursor: pointer; outline: none; user-select: none; }
.app-select-option[data-highlighted] { background: var(--ui-control-hover); }
.app-select-option[data-disabled] { opacity: .45; pointer-events: none; }
.app-select-option__indicator { position: absolute; right: 9px; color: var(--ui-primary); }
</style>
