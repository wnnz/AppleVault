<template>
  <PopoverRoot :open="modelValue" @update:open="emit('update:modelValue', $event)">
    <PopoverTrigger as-child>
      <slot name="trigger" />
    </PopoverTrigger>
    <PopoverPortal>
      <PopoverContent
        class="app-popover-content"
        :side="side"
        :align="align"
        :side-offset="8"
        :collision-padding="8"
      >
        <slot />
      </PopoverContent>
    </PopoverPortal>
  </PopoverRoot>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { PopoverContent, PopoverPortal, PopoverRoot, PopoverTrigger } from 'reka-ui'

const props = withDefaults(defineProps<{
  modelValue?: boolean
  placement?: 'top' | 'top-start' | 'top-end' | 'bottom' | 'bottom-start' | 'bottom-end' | 'left' | 'right'
}>(), {
  modelValue: false,
  placement: 'top-end'
})
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()
const side = computed(() => props.placement.split('-')[0] as 'top' | 'bottom' | 'left' | 'right')
const align = computed(() => props.placement.endsWith('-start') ? 'start' : props.placement.endsWith('-end') ? 'end' : 'center')
</script>

<style>
.app-popover-content { z-index: 10040; outline: none; }
</style>
