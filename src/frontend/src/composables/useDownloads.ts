import { ref, computed } from 'vue'
import { backend as main } from '../../wailsjs/go/models'
import {
  AddDownloadTask,
  GetDownloadTasks,
  CancelDownloadTask,
  DeleteDownloadTask,
  ClearCompletedDownloadTasks
} from '../../wailsjs/go/backend/App'
import { useAppMessage } from '../ui/feedback'

export function useDownloads(options?: {
  onTaskAdded?: () => void
}) {
  const message = useAppMessage()
  const downloadTasks = ref<main.DownloadTask[]>([])

  const activeTaskCount = computed(() =>
    downloadTasks.value.filter(t => t.status === 'downloading' || t.status === 'pending').length
  )

  const completedTaskCount = computed(() =>
    downloadTasks.value.filter(t => t.status === 'completed').length
  )

  async function loadDownloadTasks() {
    try {
      const list = await GetDownloadTasks()
      downloadTasks.value = list || []
    } catch (err: any) {
      console.error('加载任务失败:', err)
    }
  }

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
      options?.onTaskAdded?.()
    } catch (err: any) {
      message.error(`添加下载失败: ${err}`)
    }
  }

  async function handleClearCompleted() {
    try {
      await ClearCompletedDownloadTasks()
      await loadDownloadTasks()
      message.success('已清空所有已完成任务')
    } catch (err: any) {
      message.error(`操作失败: ${err}`)
    }
  }

  async function handleCancelTask(id: string) {
    try {
      await CancelDownloadTask(id)
      message.info('正在取消任务...')
    } catch (err: any) {
      message.error(`取消失败: ${err}`)
    }
  }

  async function handleDeleteTask(id: string) {
    try {
      await DeleteDownloadTask(id)
      await loadDownloadTasks()
    } catch (err: any) {
      message.error(`删除失败: ${err}`)
    }
  }

  async function handleRetryTask(task: main.DownloadTask) {
    try {
      await AddDownloadTask(
        task.appName,
        task.bundleID,
        task.appId,
        task.version,
        task.versionId,
        task.fileSize
      )
      message.success(`已重新添加任务「${task.appName}」`)
    } catch (err: any) {
      message.error(`重试失败: ${err}`)
    }
  }

  function updateTask(updatedTask: main.DownloadTask) {
    const idx = downloadTasks.value.findIndex(t => t.id === updatedTask.id)
    if (idx !== -1) {
      downloadTasks.value[idx] = updatedTask
    } else {
      downloadTasks.value.unshift(updatedTask)
    }
  }

  return {
    downloadTasks,
    activeTaskCount,
    completedTaskCount,
    loadDownloadTasks,
    addDownloadTask,
    handleClearCompleted,
    handleCancelTask,
    handleDeleteTask,
    handleRetryTask,
    updateTask
  }
}
