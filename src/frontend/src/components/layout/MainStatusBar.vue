<template>
  <div class="bottom-status-container">
    <!-- Bottom Minimal Status Bar -->
    <div class="bottom-status-bar">
      <div class="status-indicator-section">
        <span class="status-pulse" :class="{ 'pulse-busy': isAnyOperationRunning }"></span>
        <span class="status-summary-text text-ellipsis">{{ statusText }}</span>
      </div>

      <div class="status-bar-tools">
        <AppButton
          v-if="isAnyOperationRunning"
          size="tiny"
          type="error"
          secondary
          class="mr-2"
          @click="emit('cancel')"
        >
          终止操作
        </AppButton>

        <button class="status-btn" @click="emit('toggleLogs')">
          {{ showLogs ? '收起日志' : '实时日志' }}
        </button>
      </div>
    </div>

    <!-- Collapsible Sleek Terminal Drawer -->
    <div v-show="showLogs" class="sleek-console-drawer">
      <div class="console-action-bar">
        <div class="console-title">实时执行日志</div>
        <div class="console-btns">
          <button class="console-bar-btn" @click="emit('clearLogs')">清空</button>
          <button class="console-bar-btn" @click="emit('toggleLogs')">收起</button>
        </div>
      </div>
      <div ref="logContainerRef" class="console-scroll-screen">
        <div v-if="logLines.length === 0" class="console-empty-tip">等待命令输出...</div>
        <div
          v-for="(log, idx) in logLines"
          :key="idx"
          class="console-row"
          :class="{ 'row-err': log.includes('[ERR]') }"
        >
          {{ log }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import AppButton from '../../ui/components/AppButton.vue'

defineProps<{
  isAnyOperationRunning: boolean
  statusText: string
  showLogs: boolean
  logLines: string[]
}>()

const emit = defineEmits<{
  (e: 'cancel'): void
  (e: 'toggleLogs'): void
  (e: 'clearLogs'): void
}>()

const logContainerRef = ref<HTMLElement | null>(null)

defineExpose({
  logContainerRef
})
</script>
