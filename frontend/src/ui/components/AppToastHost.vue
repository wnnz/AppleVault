<template>
  <ToastProvider :duration="3200" swipe-direction="right">
    <ToastRoot
      v-for="toast in appFeedbackState.toasts"
      :key="toast.id"
      :open="toast.open"
      class="app-toast"
      :class="`app-toast--${toast.type}`"
      @update:open="!$event && dismissMessage(toast.id)"
    >
      <ToastTitle class="app-toast__title">{{ labels[toast.type] }}</ToastTitle>
      <ToastDescription class="app-toast__description">{{ toast.content }}</ToastDescription>
      <ToastClose class="app-toast__close" aria-label="关闭">×</ToastClose>
    </ToastRoot>
    <ToastViewport class="app-toast-viewport" />
  </ToastProvider>
</template>

<script setup lang="ts">
import { ToastClose, ToastDescription, ToastProvider, ToastRoot, ToastTitle, ToastViewport } from 'reka-ui'
import { appFeedbackState, dismissMessage, type AppMessageType } from '../feedback'
const labels: Record<AppMessageType, string> = { success: '成功', error: '错误', warning: '提示', info: '信息' }
</script>

<style>
.app-toast-viewport { position: fixed; z-index: 11000; top: 18px; right: 18px; display: flex; flex-direction: column; gap: 8px; width: min(360px, calc(100vw - 36px)); margin: 0; padding: 0; list-style: none; outline: none; }
.app-toast { position: relative; display: grid; grid-template-columns: 1fr auto; gap: 2px 12px; padding: 11px 36px 11px 14px; border: 1px solid var(--ui-border); border-left: 4px solid var(--ui-primary); border-radius: 10px; background: var(--ui-surface); color: var(--ui-text); box-shadow: 0 10px 30px rgba(15,23,42,.16); }
.app-toast[data-state='open'] { animation: toast-in .2s ease-out; }
.app-toast[data-state='closed'] { animation: toast-out .18s ease-in; }
.app-toast--success { border-left-color: var(--ui-success); }
.app-toast--error { border-left-color: var(--ui-danger); }
.app-toast--warning { border-left-color: var(--ui-warning); }
.app-toast__title { font-size: 13px; font-weight: 700; }
.app-toast__description { grid-column: 1; color: var(--ui-muted); font-size: 12px; line-height: 1.4; }
.app-toast__close { position: absolute; top: 8px; right: 9px; border: 0; background: transparent; color: var(--ui-muted); font-size: 18px; cursor: pointer; }
@keyframes toast-in { from { opacity: 0; transform: translateX(16px); } }
@keyframes toast-out { to { opacity: 0; transform: translateX(16px); } }
</style>
