import { reactive } from 'vue'

export type AppMessageType = 'success' | 'error' | 'warning' | 'info'
export interface AppToastMessage { id: number; type: AppMessageType; content: string; open: boolean }
export interface AppDialogOptions {
  title: string
  content: string
  positiveText?: string
  negativeText?: string
  onPositiveClick?: () => void | Promise<void>
  onNegativeClick?: () => void
}

let nextToastId = 1
export const appFeedbackState = reactive({
  toasts: [] as AppToastMessage[],
  dialog: null as (AppDialogOptions & { open: boolean; type: 'warning' | 'info' }) | null
})

function pushMessage(type: AppMessageType, content: unknown) {
  const toast: AppToastMessage = { id: nextToastId++, type, content: String(content), open: true }
  appFeedbackState.toasts.push(toast)
  window.setTimeout(() => dismissMessage(toast.id), 3200)
}

export function dismissMessage(id: number) {
  const toast = appFeedbackState.toasts.find(item => item.id === id)
  if (toast) toast.open = false
  window.setTimeout(() => {
    const index = appFeedbackState.toasts.findIndex(item => item.id === id)
    if (index >= 0) appFeedbackState.toasts.splice(index, 1)
  }, 220)
}

export function useAppMessage() {
  return {
    success: (content: unknown) => pushMessage('success', content),
    error: (content: unknown) => pushMessage('error', content),
    warning: (content: unknown) => pushMessage('warning', content),
    info: (content: unknown) => pushMessage('info', content)
  }
}

function showDialog(type: 'warning' | 'info', options: AppDialogOptions) {
  appFeedbackState.dialog = { ...options, type, open: true }
}

export function useAppDialog() {
  return {
    warning: (options: AppDialogOptions) => showDialog('warning', options),
    info: (options: AppDialogOptions) => showDialog('info', options)
  }
}
