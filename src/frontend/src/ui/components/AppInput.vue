<template>
  <input
    ref="inputRef"
    :value="modelValue"
    :type="type"
    :disabled="disabled"
    :readonly="readonly"
    :class="['app-input', `app-input--${size}`, { 'app-input--bare': bare }]"
    @input="onInput"
  />
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { ControlSize } from '../control'

withDefaults(defineProps<{
  modelValue?: string | number | null
  type?: string
  size?: ControlSize
  bare?: boolean
  disabled?: boolean
  readonly?: boolean
}>(), {
  modelValue: '',
  type: 'text',
  size: 'medium',
  bare: false,
  disabled: false,
  readonly: false
})

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()
const inputRef = ref<HTMLInputElement | null>(null)

function onInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).value)
}

defineExpose({
  focus: () => inputRef.value?.focus(),
  select: () => inputRef.value?.select(),
  element: inputRef
})
</script>

<style>
.app-input {
  width: 100%;
  padding: 0 12px;
  border: 1px solid var(--ui-border);
  border-radius: 8px;
  background: var(--ui-control-bg);
  color: var(--ui-text);
  font: inherit;
  font-size: 13px;
  line-height: normal;
  outline: none;
  box-sizing: border-box;
  transition: border-color .15s ease, background-color .15s ease, box-shadow .15s ease;
}

.app-input--small { min-height: var(--ui-control-height-small); height: var(--ui-control-height-small); }
.app-input--medium { min-height: var(--ui-control-height-medium); height: var(--ui-control-height-medium); }
.app-input--large { min-height: var(--ui-control-height-large); height: var(--ui-control-height-large); }
.theme-handdrawn .app-input.app-input--bare { border: 0 !important; border-image: none !important; background: transparent !important; }
.app-input--bare {
  min-height: 0;
  height: 100%;
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}
.app-input:focus { border-color: var(--ui-primary); box-shadow: 0 0 0 2px var(--ui-focus); }
.app-input--bare:focus { border-color: transparent; box-shadow: none; }
.theme-handdrawn .app-input:focus { border-color: transparent; outline: none; box-shadow: none; }
.app-input:disabled { cursor: not-allowed; opacity: .5; }
</style>
