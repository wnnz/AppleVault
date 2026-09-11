<template>
  <ProgressRoot
    class="app-progress"
    :model-value="percentage"
    :max="100"
    :aria-label="`${percentage}%`"
  >
    <div class="app-progress__rail">
      <ProgressIndicator
        class="app-progress__fill"
        :class="status === 'success' ? 'app-progress__fill--success' : ''"
        :style="{ transform: `translateX(-${100 - clampedPercentage}%)`, backgroundColor: color }"
      />
    </div>
  </ProgressRoot>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ProgressIndicator, ProgressRoot } from 'reka-ui'
const props = withDefaults(defineProps<{ percentage?: number; color?: string; status?: 'success' | 'error' | 'info' | 'default'; height?: number }>(), {
  percentage: 0,
  status: 'default',
  height: 6
})
const clampedPercentage = computed(() => Math.min(100, Math.max(0, props.percentage)))
</script>

<style>
.app-progress { width: 100%; }
.app-progress__rail { width: 100%; height: 6px; overflow: hidden; border-radius: 999px; background: var(--ui-progress-track); }
.app-progress__fill { width: 100%; height: 100%; border-radius: inherit; background: var(--ui-primary); transition: transform .25s ease; }
.app-progress__fill--success { background: var(--ui-success); }
</style>
