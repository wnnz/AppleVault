<template>
  <DialogRoot :open="!!dialog?.open" @update:open="onOpenChange">
    <DialogPortal>
      <DialogOverlay class="app-dialog-overlay" />
      <DialogContent class="app-dialog-content">
        <DialogTitle class="app-dialog-title">{{ dialog?.title }}</DialogTitle>
        <DialogDescription class="app-dialog-description">{{ dialog?.content }}</DialogDescription>
        <div class="app-dialog-actions">
          <AppButton variant="secondary" :disabled="isSubmitting" @click="cancel">{{ dialog?.negativeText || '取消' }}</AppButton>
          <AppButton variant="primary" :loading="isSubmitting" @click="confirm">{{ dialog?.positiveText || '确定' }}</AppButton>
        </div>
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { DialogContent, DialogDescription, DialogOverlay, DialogPortal, DialogRoot, DialogTitle } from 'reka-ui'
import { appFeedbackState } from '../feedback'
import AppButton from './AppButton.vue'
const dialog = computed(() => appFeedbackState.dialog)
const isSubmitting = ref(false)
function close() { if (appFeedbackState.dialog) appFeedbackState.dialog.open = false }
function cancel() { dialog.value?.onNegativeClick?.(); close() }
async function confirm() {
  if (isSubmitting.value) return
  isSubmitting.value = true
  try {
    await dialog.value?.onPositiveClick?.()
  } finally {
    isSubmitting.value = false
    close()
  }
}
function onOpenChange(open: boolean) { if (!open && !isSubmitting.value) cancel() }
</script>

<style>
.app-dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}
.app-dialog-actions .app-button {
  min-width: 90px;
  height: 34px;
  box-sizing: border-box;
}
</style>
