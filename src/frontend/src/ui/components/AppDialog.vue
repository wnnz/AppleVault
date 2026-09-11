<template>
  <DialogRoot :open="modelValue" @update:open="emit('update:modelValue', $event)">
    <DialogPortal>
      <DialogOverlay class="app-dialog-overlay" />
      <DialogContent class="app-dialog-content" v-bind="$attrs" @escape-key-down="onEscape">
        <DialogTitle class="app-dialog-title">{{ title }}</DialogTitle>
        <DialogDescription v-if="description" class="app-dialog-description">{{ description }}</DialogDescription>
        <slot />
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>

<script setup lang="ts">
import { DialogContent, DialogDescription, DialogOverlay, DialogPortal, DialogRoot, DialogTitle } from 'reka-ui'
defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{ modelValue: boolean; title: string; description?: string; maskClosable?: boolean }>(), { description: '', maskClosable: true })
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()
function onEscape(event: Event) { if (!props.maskClosable) event.preventDefault() }
</script>

<style>
.app-dialog-overlay { position: fixed; z-index: 10000; inset: 0; background: rgba(15, 23, 42, .46); }
.app-dialog-content { position: fixed; z-index: 10001; top: 50%; left: 50%; width: min(420px, calc(100vw - 32px)); padding: 18px; transform: translate(-50%, -50%); border: 1px solid var(--ui-border); border-radius: 14px; background: var(--ui-surface); color: var(--ui-text); box-shadow: 0 20px 60px rgba(15,23,42,.28); box-sizing: border-box; outline: none; }
.app-dialog-title { margin: 0 0 8px; font-size: 17px; font-weight: 700; }
.app-dialog-description { margin: 0 0 12px; color: var(--ui-muted); font-size: 13px; line-height: 1.5; }
</style>
