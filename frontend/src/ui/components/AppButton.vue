<template>
  <button
    :type="nativeType"
    :disabled="disabled || loading"
    :aria-busy="loading || undefined"
    :class="[
      'app-button',
      `app-button--${resolvedVariant}`,
      `app-button--${size}`,
      `app-button--type-${resolvedType}`,
      {
        'app-button--tertiary': tertiary,
        'app-button--quaternary': quaternary,
        'app-button--disabled': disabled || loading,
        'app-button--block': block,
        'app-button--loading': loading
      }
    ]"
    @click="onClick"
  >
    <span v-if="loading" class="app-button__spinner" aria-hidden="true" />
    <span class="app-button__content"><slot /></span>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type AppButtonVariant = 'primary' | 'secondary'
type AppButtonType = 'default' | 'tertiary' | 'primary' | 'info' | 'success' | 'warning' | 'error'
type AppButtonSize = 'tiny' | 'small' | 'medium' | 'large'

const props = withDefaults(defineProps<{
  variant?: AppButtonVariant
  type?: AppButtonType
  nativeType?: 'button' | 'submit' | 'reset'
  size?: AppButtonSize
  loading?: boolean
  disabled?: boolean
  block?: boolean
  secondary?: boolean
  tertiary?: boolean
  quaternary?: boolean
}>(), {
  nativeType: 'button',
  size: 'medium',
  loading: false,
  disabled: false,
  block: false,
  secondary: false,
  tertiary: false,
  quaternary: false
})

const emit = defineEmits<{ click: [event: MouseEvent] }>()

const resolvedVariant = computed<AppButtonVariant>(() =>
  props.variant ?? (props.type === 'primary' ? 'primary' : 'secondary')
)
const resolvedType = computed<AppButtonType>(() =>
  props.type ?? (resolvedVariant.value === 'primary' ? 'primary' : 'default')
)

function onClick(event: MouseEvent) {
  if (!props.disabled && !props.loading) emit('click', event)
}
</script>

<style>
.app-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  min-width: 0;
  border: 1px solid transparent;
  border-radius: 8px;
  padding: 0 14px;
  background: transparent;
  color: var(--ui-text, #334155);
  font: inherit;
  font-size: 14px;
  font-weight: 500;
  line-height: 1;
  white-space: nowrap;
  cursor: pointer;
  box-sizing: border-box;
  transition: color .15s ease, background-color .15s ease, border-color .15s ease, opacity .15s ease;
}

.app-button--tiny { height: 22px; padding: 0 7px; border-radius: 6px; font-size: 12px; }
.app-button--small { height: 28px; padding: 0 12px; font-size: 12px; }
.app-button--medium { height: 34px; }
.app-button--large { height: 40px; padding: 0 18px; }
.app-button--block { display: flex; width: 100%; }
.app-button--primary { border-color: var(--ui-primary); background: var(--ui-primary); color: #fff; }
.app-button--primary:hover:not(:disabled) { background: var(--ui-primary-hover); }
.app-button--secondary { border-color: var(--ui-border); background: var(--ui-control-bg); color: var(--ui-text); }
.app-button--secondary:hover:not(:disabled) { background: var(--ui-control-hover); }
.app-button:disabled { cursor: not-allowed; opacity: .5; }
.app-button__content { display: inline-flex; align-items: center; justify-content: center; }
.app-button__spinner { width: 12px; height: 12px; border: 2px solid currentColor; border-right-color: transparent; border-radius: 50%; animation: app-spin .7s linear infinite; }
@keyframes app-spin { to { transform: rotate(360deg); } }

.theme-handdrawn .app-button {
  border-radius: 9px !important;
  border-color: transparent !important;
  background-color: transparent !important;
  color: #31547e !important;
  font-family: "Microsoft YaHei", "Segoe UI", "PingFang SC", "DengXian", sans-serif;
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0.15px;
  box-shadow: none !important;
}

.theme-handdrawn .app-button:hover { color: var(--hd-blue-dark) !important; box-shadow: none !important; }
.theme-handdrawn .app-button--secondary:not(.app-button--tertiary):not(.app-button--quaternary) {
  border: 1px solid transparent !important;
  border-image: url('../themes/handdrawn/assets/sketch-frame-blue.svg') 24 / 7px stretch !important;
  border-radius: 0 !important;
  background-color: rgba(255, 255, 255, 0.72) !important;
}
.theme-handdrawn .app-button--secondary:not(.app-button--tertiary):not(.app-button--quaternary):hover,
.theme-handdrawn .app-button--secondary.app-button--disabled:not(.app-button--tertiary):not(.app-button--quaternary) {
  background-color: rgba(239, 250, 255, 0.84) !important;
}
.theme-handdrawn .app-button--primary {
  /* The light controls use a hand-drawn border with transparent outer
     breathing room. Keep the same optical height for filled controls. */
  border-width: 2px !important;
  border-color: transparent !important;
  color: #fff !important;
  background-color: var(--hd-blue) !important;
  background-image: url('../themes/handdrawn/assets/handdrawn-blue-button-texture.webp') !important;
  background-size: 240px 80px !important;
  background-position: 0 0 !important;
  background-repeat: repeat !important;
  background-clip: padding-box !important;
}
.theme-handdrawn .app-button--primary:hover { color: #fff !important; border-color: transparent !important; background-color: #1288e2 !important; }
.theme-handdrawn .app-button--secondary.app-button--type-error { color: #d94f5c !important; }
.theme-handdrawn .app-button--tiny { padding-right: 6px; padding-left: 6px; font-size: 12px; }
</style>
