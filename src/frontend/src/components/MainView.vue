<template>
  <div
    class="app-layout"
    :class="[
      `theme-${effectiveTheme}`,
      isDarkTheme ? 'dark-mode' : 'light-mode',
      'minimal-mode',
      `active-tab-${activeTab}`
    ]"
  >
    <!-- Left Modern Sidebar -->
    <MainSidebar
      v-model:active-tab="activeTab"
      :effective-theme="effectiveTheme"
      :is-dark-theme="isDarkTheme"
      :account="currentAccount"
      :is-logged-in="isLoggedIn"
      :enable-proxy="currentSettings.enableProxy"
      :active-task-count="activeTaskCount"
      @update:current-theme="emit('update:currentTheme', $event)"
      @proxy-toggle="onProxyToggle"
    />

    <!-- Right Main Workspace -->
    <section class="app-main">
      <div class="view-content-wrapper">
        <!-- VIEW 1: 应用搜索 (search) -->
        <SearchView
          v-show="activeTab === 'search'"
          ref="searchViewRef"
          :effective-theme="effectiveTheme"
          @select-app-for-versions="handleSelectAppForVersions"
          @download-app="handleDownloadApp"
          @busy-change="handleBusyChange"
        />

        <!-- VIEW 2: 历史版本 (versions) -->
        <VersionsView
          v-show="activeTab === 'versions'"
          ref="versionsViewRef"
          :effective-theme="effectiveTheme"
          @download-version="handleDownloadVersion"
          @busy-change="handleBusyChange"
        />

        <!-- VIEW 3: 下载中心 (download) -->
        <DownloadView
          v-show="activeTab === 'download'"
          ref="downloadViewRef"
          @open-download-dir="handleOpenDefaultDownloadDir"
          @install-task="handleInstallTask"
          @active-count-changed="onActiveCountChanged"
        />

        <!-- VIEW 4: 已购应用 (purchased) -->
        <PurchasedView
          v-show="activeTab === 'purchased'"
          ref="purchasedViewRef"
          :effective-theme="effectiveTheme"
          @select-app-for-versions="handleSelectAppForVersions"
          @download-app="handleDownloadApp"
          @busy-change="handleBusyChange"
        />

        <!-- VIEW 5: 设备直装 (installer) -->
        <InstallerView
          v-show="activeTab === 'installer'"
          ref="installerViewRef"
          @busy-change="handleBusyChange"
        />

        <!-- VIEW 6: 账号中心 (account) -->
        <AccountView
          v-show="activeTab === 'account'"
          ref="accountViewRef"
          @login-success="onLoginSuccess"
          @account-changed="onAccountChanged"
          @busy-change="handleBusyChange"
          @navigate="activeTab = $event"
        />

        <!-- VIEW 7: 系统设置 (settings) -->
        <SettingsView
          v-show="activeTab === 'settings'"
          ref="settingsViewRef"
          @settings-changed="onSettingsChanged"
          @busy-change="handleBusyChange"
        />

        <!-- VIEW 8: 关于软件 (about) -->
        <AboutView v-show="activeTab === 'about'" />
      </div>

      <!-- Bottom Minimal Status Bar & Drawer -->
      <MainStatusBar
        ref="statusBarRef"
        :is-any-operation-running="isAnyOperationRunning"
        :status-text="statusText"
        @cancel="handleCancel"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import MainSidebar from './layout/MainSidebar.vue'
import MainStatusBar from './layout/MainStatusBar.vue'
import AboutView from './views/AboutView.vue'
import AccountView from './views/AccountView.vue'
import DownloadView from './views/DownloadView.vue'
import InstallerView from './views/InstallerView.vue'
import PurchasedView from './views/PurchasedView.vue'
import SearchView from './views/SearchView.vue'
import SettingsView from './views/SettingsView.vue'
import VersionsView, { type VersionItem } from './views/VersionsView.vue'
import { CancelRunningCommand } from '../../wailsjs/go/backend/App'
import { EventsOn, OnFileDrop, OnFileDropOff } from '../../wailsjs/runtime/runtime'
import { backend as main } from '../../wailsjs/go/models'
import type { AppTheme } from '../ui/theme'

const props = withDefaults(
  defineProps<{
    currentTheme?: AppTheme
    isDark?: boolean
  }>(),
  {
    currentTheme: 'minimal-light',
    isDark: false
  }
)

const emit = defineEmits<{
  (e: 'update:currentTheme', theme: AppTheme): void
  (e: 'toggleTheme'): void
}>()

const effectiveTheme = computed<AppTheme>(() => props.currentTheme || (props.isDark ? 'minimal-dark' : 'minimal-light'))
const isDarkTheme = computed(() => effectiveTheme.value.endsWith('-dark') || props.isDark)

const activeTab = ref('search')
const isAnyOperationRunning = ref(false)
const statusText = ref('就绪')

// 各子视图组件引用
const searchViewRef = ref<InstanceType<typeof SearchView> | null>(null)
const versionsViewRef = ref<InstanceType<typeof VersionsView> | null>(null)
const downloadViewRef = ref<InstanceType<typeof DownloadView> | null>(null)
const purchasedViewRef = ref<InstanceType<typeof PurchasedView> | null>(null)
const installerViewRef = ref<InstanceType<typeof InstallerView> | null>(null)
const accountViewRef = ref<InstanceType<typeof AccountView> | null>(null)
const settingsViewRef = ref<InstanceType<typeof SettingsView> | null>(null)
const statusBarRef = ref<InstanceType<typeof MainStatusBar> | null>(null)

// 共享的账号与配置信息（供 Sidebar 等外部组件展示）
const currentAccount = ref<main.AccountInfo>({ name: '', email: '', success: false })
const isLoggedIn = computed(() => currentAccount.value.success && !!currentAccount.value.email)
const currentSettings = ref<main.Settings>({
  keychainPassphrase: '123456',
  defaultDownloadDir: 'data/downloads/default',
  defaultPlatform: 'iphone',
  enableProxy: true,
  proxyUrl: 'http://127.0.0.1:10808',
  ipaToolPath: ''
})
const activeTaskCount = ref(0)

function handleBusyChange(busy: boolean, text?: string) {
  isAnyOperationRunning.value = busy
  if (text) {
    statusText.value = text
  }
}

function onAccountChanged(acc: main.AccountInfo) {
  currentAccount.value = acc
}

function onLoginSuccess(acc: main.AccountInfo) {
  currentAccount.value = acc
  activeTab.value = 'search'
  void settingsViewRef.value?.loadSettings()
}

function onSettingsChanged(s: main.Settings) {
  currentSettings.value = s
}

function onProxyToggle(val: boolean) {
  void settingsViewRef.value?.onProxyToggle(val)
}

function onActiveCountChanged(count: number) {
  activeTaskCount.value = count
}

function handleOpenDefaultDownloadDir() {
  settingsViewRef.value?.handleOpenDefaultDownloadDir()
}

// 跨视图跳转联动
function handleSelectAppForVersions(app: main.AppItem) {
  activeTab.value = 'versions'
  versionsViewRef.value?.selectAppForVersions(app)
}

function handleDownloadApp(app: main.AppItem) {
  downloadViewRef.value?.addDownloadTask(app.name, app.bundleID, app.id, app.version || '最新版')
  activeTab.value = 'download'
}

function handleDownloadVersion(row: VersionItem, form: { appName: string; bundleId: string; appId: number }) {
  const appName = form.appName || form.bundleId
  const ver = row.displayVersion !== '未查询' ? row.displayVersion : (row.versionId ? `Build ${row.versionId}` : '最新版')
  downloadViewRef.value?.addDownloadTask(appName, form.bundleId, form.appId, ver, row.versionId, row.fileSize)
  activeTab.value = 'download'
}

function handleInstallTask(outputPath: string) {
  activeTab.value = 'installer'
  if (outputPath) {
    installerViewRef.value?.useIPAPath([outputPath])
  }
  if (!installerViewRef.value?.devices.length) {
    void installerViewRef.value?.loadConnectedDevices()
  }
}

function handleCancel() {
  versionsViewRef.value?.stopAllQueries()
  CancelRunningCommand()
  isAnyOperationRunning.value = false
  statusText.value = '已终止当前操作'
}

// 生命周期与 Wails 系统事件
onMounted(() => {
  const urlParams = new URLSearchParams(window.location.search)
  const tabParam = urlParams.get('tab')
  const isDemo = urlParams.get('demo') === '1'

  if (isDemo) {
    currentAccount.value = { name: '果仓助手用户', email: 'applevault.user@icloud.com', success: true }
    statusText.value = '就绪 (Demo 模式)'
  } else {
    void settingsViewRef.value?.loadSettings()
    void accountViewRef.value?.refreshAccount(true)
    void downloadViewRef.value?.loadDownloadTasks()
  }

  if (tabParam) {
    activeTab.value = tabParam
  }

  // 接收文件拖放
  OnFileDrop((_x: number, _y: number, paths: string[]) => {
    activeTab.value = 'installer'
    installerViewRef.value?.useIPAPath(paths)
  }, true)

  // 监听后端执行日志并投递到状态栏终端
  EventsOn('log', (msg: string) => {
    statusBarRef.value?.appendLog(msg)
  })

  // 监听后端任务进度更新与重载
  EventsOn('download-task-updated', (updatedTask: main.DownloadTask) => {
    downloadViewRef.value?.updateTask(updatedTask)
  })

  EventsOn('download-tasks-reload', () => {
    void downloadViewRef.value?.loadDownloadTasks()
  })
})

onBeforeUnmount(() => {
  OnFileDropOff()
})
</script>

<style src="./MainView.css"></style>
