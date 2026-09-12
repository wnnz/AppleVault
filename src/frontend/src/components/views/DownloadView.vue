<template>
  <div class="view-panel">
    <div class="view-header">
      <div>
        <div class="title-with-badge"><h2 class="view-title">下载中心</h2></div>
        <p class="view-desc">正在下载与已完成的 IPA 任务管理，自动归档至对应账号目录</p>
      </div>
      <div class="btn-group-row">
        <AppButton
          variant="secondary"
          size="small"
          :disabled="completedTaskCount === 0"
          @click="handleClearCompleted"
        >
          清空已完成
        </AppButton>
        <AppButton variant="secondary" size="small" @click="emit('openDownloadDir')">打开存储目录</AppButton>
      </div>
    </div>

    <div class="tasks-container">
      <div v-if="downloadTasks.length === 0" class="empty-state-card">
        <div class="empty-title">暂无下载任务</div>
        <div class="empty-desc">前往「应用搜索」或「历史版本」点击下载即可在此实时追踪进度</div>
      </div>

      <div v-else class="task-grid">
        <div v-for="task in downloadTasks" :key="task.id" class="modern-task-card">
          <div class="task-card-header">
            <div class="task-app-title-group">
              <span class="task-name">{{ task.appName }}</span>
              <span class="task-pill task-pill-ver">{{ task.version }}</span>
              <span v-if="task.versionId" class="task-pill task-pill-build">Build {{ task.versionId }}</span>
            </div>
            <div class="task-actions-group">
              <AppButton
                v-if="task.status === 'downloading'"
                size="tiny"
                type="error"
                secondary
                @click="handleCancelTask(task.id)"
              >
                取消
              </AppButton>
              <AppButton
                v-if="task.status === 'completed'"
                size="tiny"
                type="primary"
                @click="emit('installTask', task.outputPath)"
              >
                安装到设备
              </AppButton>
              <AppButton
                v-if="task.status === 'error' || task.status === 'canceled'"
                size="tiny"
                secondary
                @click="handleRetryTask(task)"
              >
                重试
              </AppButton>
              <AppButton size="tiny" quaternary @click="handleDeleteTask(task.id)">删除</AppButton>
            </div>
          </div>

          <div class="task-meta-bundle">{{ task.bundleID }}</div>
          <div class="task-progress-section">
            <AppProgress
              :percentage="task.progress"
              :status="getProgressStatus(task.status)"
              :class="'task-progress-' + task.status"
              :height="6"
            />
          </div>

          <div class="task-card-footer">
            <div class="footer-status-tag">
              <span class="status-indicator" :class="'indicator-' + task.status"></span>
              <span class="status-text">{{ getStatusText(task.status) }}</span>
              <span v-if="task.status === 'downloading' && task.speed" class="task-speed">{{ task.speed }}</span>
              <span v-if="task.status === 'downloading' && task.totalBytes > 0" class="task-bytes-info">
                {{ formatBytes(task.currBytes) }} / {{ formatBytes(task.totalBytes) }}
              </span>
              <span v-else-if="task.fileSize && task.fileSize !== '-'" class="task-bytes-info">体积: {{ task.fileSize }}</span>
              <span v-if="task.status === 'error'" class="task-error-text" :title="task.errorMessage">{{ task.errorMessage }}</span>
            </div>
            <div class="footer-right-info">
              <span class="pct-text font-bold">{{ task.progress }}%</span>
              <span class="time-text">{{ task.createdAt }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { backend as main } from '../../../wailsjs/go/models'
import {
  AddDownloadTask,
  RetryDownloadTask,
  GetDownloadTasks,
  CancelDownloadTask,
  DeleteDownloadTask,
  ClearCompletedDownloadTasks
} from '../../../wailsjs/go/backend/App'
import AppButton from '../../ui/components/AppButton.vue'
import AppProgress from '../../ui/components/AppProgress.vue'
import { useAppMessage } from '../../ui/feedback'

const emit = defineEmits<{
  (e: 'openDownloadDir'): void
  (e: 'installTask', outputPath: string): void
  (e: 'activeCountChanged', count: number): void
}>()

const message = useAppMessage()
const downloadTasks = ref<main.DownloadTask[]>([])

const activeTaskCount = computed(() =>
  downloadTasks.value.filter(t => t.status === 'downloading' || t.status === 'pending').length
)

const completedTaskCount = computed(() =>
  downloadTasks.value.filter(t => t.status === 'completed').length
)

/**
 * 载入下载任务列表
 */
async function loadDownloadTasks() {
  try {
    const list = await GetDownloadTasks()
    downloadTasks.value = list || []
    emit('activeCountChanged', activeTaskCount.value)
  } catch (err: any) {
    console.error('加载任务失败:', err)
  }
}

/**
 * 添加下载任务
 */
async function addDownloadTask(
  appName: string,
  bundleID: string,
  appId: number,
  version: string,
  versionId: string = '',
  fileSize: string = ''
) {
  try {
    await AddDownloadTask(appName, bundleID, appId, version, versionId, fileSize)
    message.success(`已添加任务「${appName}」到下载中心`)
    await loadDownloadTasks()
  } catch (err: any) {
    message.error(`添加下载失败: ${err}`)
  }
}

/**
 * 清空所有已完成任务
 */
async function handleClearCompleted() {
  try {
    await ClearCompletedDownloadTasks()
    await loadDownloadTasks()
    message.success('已清空所有已完成任务')
  } catch (err: any) {
    message.error(`操作失败: ${err}`)
  }
}

/**
 * 取消正在下载的任务
 */
async function handleCancelTask(id: string) {
  try {
    await CancelDownloadTask(id)
    message.info('正在取消任务...')
  } catch (err: any) {
    message.error(`取消失败: ${err}`)
  }
}

/**
 * 删除指定任务
 */
async function handleDeleteTask(id: string) {
  try {
    await DeleteDownloadTask(id)
    await loadDownloadTasks()
  } catch (err: any) {
    message.error(`删除失败: ${err}`)
  }
}

/**
 * 重试失败或已取消的任务
 */
async function handleRetryTask(task: main.DownloadTask) {
  try {
    await RetryDownloadTask(task.id)
    message.success(`已重新添加任务「${task.appName}」`)
    await loadDownloadTasks()
  } catch (err: any) {
    message.error(`重试失败: ${err}`)
  }
}

/**
 * 实时更新单个任务进度状态
 */
function updateTask(updatedTask: main.DownloadTask) {
  const idx = downloadTasks.value.findIndex(t => t.id === updatedTask.id)
  if (idx !== -1) {
    downloadTasks.value[idx] = updatedTask
  } else {
    downloadTasks.value.unshift(updatedTask)
  }
  emit('activeCountChanged', activeTaskCount.value)
}

function getProgressStatus(status: string) {
  if (status === 'completed') return 'success'
  if (status === 'error') return 'error'
  return 'info'
}

function getStatusText(status: string) {
  const labels: Record<string, string> = {
    pending: '等待中',
    downloading: '下载中',
    completed: '已完成',
    error: '下载失败',
    canceled: '已取消'
  }
  return labels[status] || status
}

function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B'
  const sizes = ['B', 'KB', 'MB', 'GB']
  const index = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, index)).toFixed(2) + ' ' + sizes[index]
}

defineExpose({
  downloadTasks,
  activeTaskCount,
  loadDownloadTasks,
  addDownloadTask,
  updateTask
})
</script>
