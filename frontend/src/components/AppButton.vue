<template>
  <n-button
    v-bind="attrs"
    :type="resolvedType"
    :secondary="resolvedSecondary"
    :class="['app-button', `app-button--${resolvedVariant}`]"
  >
    <slot />
  </n-button>
</template>

<script setup lang="ts">
import { computed, useAttrs } from 'vue'
import { NButton } from 'naive-ui'

type AppButtonVariant = 'primary' | 'secondary'
type AppButtonType = 'default' | 'tertiary' | 'primary' | 'info' | 'success' | 'warning' | 'error'

defineOptions({ inheritAttrs: false })

const props = defineProps<{
  variant?: AppButtonVariant
  type?: AppButtonType
}>()

const attrs = useAttrs()
const resolvedVariant = computed<AppButtonVariant>(() =>
  props.variant ?? (props.type === 'primary' ? 'primary' : 'secondary')
)
const resolvedType = computed<AppButtonType>(() =>
  props.type ?? (resolvedVariant.value === 'primary' ? 'primary' : 'default')
)
const resolvedSecondary = computed(() => {
  if (attrs.secondary !== undefined) return attrs.secondary !== false
  if (attrs.tertiary !== undefined || attrs.quaternary !== undefined) return false
  return resolvedVariant.value === 'secondary'
})
</script>

<style>
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

.theme-handdrawn .app-button:hover {
  color: var(--hd-blue-dark) !important;
  background-color: transparent !important;
  box-shadow: none !important;
}

.theme-handdrawn .app-button--secondary:not(.n-button--tertiary):not(.n-button--quaternary) {
  box-sizing: border-box;
  border: 1px solid transparent !important;
  border-image: url('../assets/sketch-frame-blue.svg') 24 / 7px stretch !important;
  border-radius: 0 !important;
  background-color: rgba(255, 255, 255, 0.72) !important;
  background-image: none !important;
}

.theme-handdrawn .app-button--secondary:not(.n-button--tertiary):not(.n-button--quaternary):hover,
.theme-handdrawn .app-button--secondary.n-button--disabled:not(.n-button--tertiary):not(.n-button--quaternary) {
  background-color: rgba(239, 250, 255, 0.84) !important;
  background-image: none !important;
}

.theme-handdrawn .app-button--primary {
  border-color: #0b7bd6 !important;
  color: #ffffff !important;
  background-color: var(--hd-blue) !important;
  background-image: url('../assets/sketch-blue-fill.svg') !important;
  background-size: 120px 40px !important;
  background-position: 0 0 !important;
  background-repeat: repeat !important;
  box-shadow: none !important;
}

.theme-handdrawn .app-button--primary:hover {
  color: #ffffff !important;
  border-color: #096fc5 !important;
  background-color: #1288e2 !important;
  box-shadow: none !important;
}

.theme-handdrawn .app-button--secondary.n-button--error-type {
  color: #d94f5c !important;
}
</style>
