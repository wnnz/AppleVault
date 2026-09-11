<template>
  <SwitchRoot
    :model-value="modelValue"
    :disabled="disabled"
    class="app-switch"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <span class="app-switch__rail">
      <SwitchThumb class="app-switch__thumb" />
    </span>
  </SwitchRoot>
</template>

<script setup lang="ts">
import { SwitchRoot, SwitchThumb } from 'reka-ui'

withDefaults(defineProps<{ modelValue?: boolean; disabled?: boolean; size?: 'small' | 'medium' }>(), {
  modelValue: false,
  disabled: false,
  size: 'medium'
})
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()
</script>

<style>
.app-switch { appearance: none; display: inline-flex; width: 40px; height: 22px; padding: 0; border: 0; background: transparent; cursor: pointer; }
.app-switch__rail { position: relative; display: block; width: 100%; height: 100%; border-radius: 999px; background: var(--ui-switch-off); transition: background-color .18s ease; }
.app-switch[data-state='checked'] .app-switch__rail { background: var(--ui-primary); }
.app-switch__thumb { display: block; width: 18px; height: 18px; margin: 2px; border-radius: 50%; background: #fff; box-shadow: 0 1px 3px rgba(15,23,42,.22); transition: transform .18s ease; }
.app-switch__thumb[data-state='checked'] { transform: translateX(18px); }
.app-switch:disabled { cursor: not-allowed; opacity: .5; }
</style>
