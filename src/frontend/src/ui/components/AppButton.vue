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
import type { ControlSize } from '../control'

type AppButtonVariant = 'primary' | 'secondary'
type AppButtonType = 'default' | 'tertiary' | 'primary' | 'info' | 'success' | 'warning' | 'error'
type AppButtonSize = 'tiny' | ControlSize

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
  flex-shrink: 0;
  border: 1px solid transparent;
  border-radius: 4px;
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

.app-button--tiny { height: var(--ui-control-height-small); padding: 0 7px; font-size: 12px; }
.app-button--small { height: var(--ui-control-height-small); padding: 0 12px; font-size: 12px; }
.app-button--medium { height: var(--ui-control-height-medium); }
.app-button--large { height: var(--ui-control-height-large); padding: 0 18px; }
.app-button--block { display: flex; width: 100%; }
.app-button--primary { border-color: var(--ui-primary); background: var(--ui-primary); color: #fff; }
.app-button--primary:hover:not(:disabled) { background: var(--ui-primary-hover); }
.app-button--secondary { border-color: var(--ui-border); background: var(--ui-control-bg); color: var(--ui-text); }
.app-button--secondary:hover:not(:disabled) { background: var(--ui-control-hover); }
.app-button:disabled { cursor: not-allowed; opacity: .5; }
.app-button__content { position: relative; z-index: 1; display: inline-flex; align-items: center; justify-content: center; }
.app-button__spinner { position: relative; z-index: 1; width: 12px; height: 12px; border: 2px solid currentColor; border-right-color: transparent; border-radius: 50%; animation: app-spin .7s linear infinite; }
@keyframes app-spin { to { transform: rotate(360deg); } }

.theme-handdrawn .app-button {
  border-radius: 4px !important;
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
.theme-handdrawn .app-button__spinner {
  width: 14px;
  height: 4px;
  flex: 0 0 14px;
  border: 0;
  border-radius: 0;
  background:
    radial-gradient(circle at 2px 2px, currentColor 0 2px, transparent 2.2px),
    radial-gradient(circle at 7px 2px, currentColor 0 2px, transparent 2.2px),
    radial-gradient(circle at 12px 2px, currentColor 0 2px, transparent 2.2px);
  animation: handdrawn-button-loading 1s ease-in-out infinite;
}

@keyframes handdrawn-button-loading {
  0%, 100% { opacity: .45; transform: translateY(0); }
  50% { opacity: 1; transform: translateY(-1px); }
}

.theme-handdrawn .app-button--secondary:not(.app-button--tertiary):not(.app-button--quaternary) {
  border: 1px solid transparent !important;
  border-image: url('../themes/handdrawn/assets/sketch-frame-blue.svg') 24 / 6px stretch !important;
  background-color: rgba(255, 255, 255, 0.72) !important;
}
.theme-handdrawn .app-button--secondary:not(.app-button--tertiary):not(.app-button--quaternary):hover,
.theme-handdrawn .app-button--secondary.app-button--disabled:not(.app-button--tertiary):not(.app-button--quaternary) {
  background-color: rgba(239, 250, 255, 0.84) !important;
}
.theme-handdrawn .app-button--primary {
  position: relative;
  isolation: isolate;
  border-width: 1px !important;
  border-color: transparent !important;
  color: #fff !important;
  background-color: transparent !important;
  background-image: none !important;
}
.theme-handdrawn .app-button--primary::before {
  content: '';
  position: absolute;
  z-index: 0;
  inset: 2px 0;
  border-radius: 4px;
  background-color: var(--hd-blue, #168ff0);
  background-image: url('../themes/handdrawn/assets/handdrawn-blue-button-texture.webp');
  background-size: 240px 80px !important;
  background-position: 0 0 !important;
  background-repeat: repeat !important;
}
.theme-handdrawn .app-button--primary:hover:not(:disabled):not(.app-button--disabled) {
  color: #fff !important;
  border-color: transparent !important;
  background-color: transparent !important;
}
.theme-handdrawn .app-button--primary:hover:not(:disabled):not(.app-button--disabled)::before {
  background-color: var(--hd-blue-dark, #1288e2);
}
.theme-handdrawn .app-button--primary:disabled,
.theme-handdrawn .app-button--primary.app-button--disabled {
  opacity: .55 !important;
  cursor: not-allowed !important;
}
.theme-handdrawn .app-button--primary:disabled::before,
.theme-handdrawn .app-button--primary.app-button--disabled::before {
  background-color: var(--hd-blue, #168ff0);
}
.theme-handdrawn .app-button--secondary.app-button--type-error { color: #d94f5c !important; }
.theme-handdrawn .app-button--tiny { padding-right: 6px; padding-left: 6px; font-size: 12px; }
</style>
