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
      :account="account"
      :is-logged-in="isLoggedIn"
      :enable-proxy="settings.enableProxy"
      :active-task-count="activeTaskCount"
      @update:current-theme="emit('update:currentTheme', $event)"
      @proxy-toggle="onProxyToggle"
    />

    <!-- Right Main Workspace -->
    <section class="app-main">
      <!-- Dynamic View Container -->
      <div class="view-content-wrapper">
        <!-- VIEW 1: 应用搜索 (search) -->
        <SearchView
          v-show="activeTab === 'search'"
          :search-form="searchForm"
          :platform-options="platformOptions"
          :limit-options="limitOptions"
          :is-searching="isSearching"
          :columns="displayedSearchColumns"
          :results="searchResults"
          @search="handleSearch"
        />

        <!-- VIEW 2: 历史版本 (versions) -->
        <VersionsView
          v-show="activeTab === 'versions'"
          v-model:show-target-version-modal="showTargetVersionModal"
          v-model:target-version-input="targetVersionInput"
          :version-form="versionForm"
          :version-items="versionItems"
          :filtered-versions="filteredVersions"
          :columns="displayedVersionColumns"
          :is-listing-versions="isListingVersions"
          :is-batch-querying="isBatchQuerying"
          :is-target-querying="isTargetQuerying"
          @list-versions="handleListVersions"
          @batch-query="handleBatchQueryClick"
          @target-query="handleTargetQueryClick"
          @confirm-target-query="confirmStartTargetQuery"
        />

        <!-- VIEW 3: 下载中心 (download) -->
        <DownloadView
          v-show="activeTab === 'download'"
          :tasks="downloadTasks"
          :completed-task-count="completedTaskCount"
          @clear-completed="handleClearCompleted"
          @open-download-dir="handleOpenDefaultDownloadDir"
          @cancel-task="handleCancelTask"
          @install-task="handleInstallFromTask"
          @retry-task="handleRetryTask"
          @delete-task="handleDeleteTask"
        />

        <!-- VIEW 4: 已购应用 (purchased) -->
        <PurchasedView
          v-show="activeTab === 'purchased'"
          v-model:search-keyword="purchasedSearchKeyword"
          :is-loading="isPurchasedLoading"
          :load-total="purchasedLoadTotal"
          :load-loaded="purchasedLoadLoaded"
          :load-progress="purchasedLoadProgress"
          :columns="displayedPurchasedColumns"
          :apps="filteredPurchasedApps"
          :match-total="purchasedMatchTotal"
          :total-apps="purchasedApps.length"
          :total="purchasedTotal"
          :page="purchasedPage"
          :page-count="purchasedPageCount"
          :page-size="purchasedPageSize"
          :page-size-options="purchasedPageSizeOptions"
          @refresh="loadPurchases"
          @previous="prevPurchasedPage"
          @next="nextPurchasedPage"
          @page-size-change="onPurchasedPageSizeChange"
        />

        <!-- VIEW 5: 设备直装 (installer) -->
        <InstallerView
          v-show="activeTab === 'installer'"
          v-model:selected-device-u-d-i-d="selectedDeviceUDID"
          :selected-i-p-a-path="selectedIPAPath"
          :selected-i-p-a-file-name="selectedIPAFileName"
          :devices="devices"
          :selected-device="selectedDevice"
          :device-options="deviceOptions"
          :is-loading-devices="isLoadingDevices"
          :is-installing-i-p-a="isInstallingIPA"
          @select-i-p-a="handleSelectIPA"
          @refresh-devices="loadConnectedDevices"
          @install-i-p-a="handleInstallIPA"
        />

        <!-- VIEW 6: 账号中心 (account) -->
        <AccountView
          v-show="activeTab === 'account'"
          v-model:show2-f-a-modal="show2FAModal"
          v-model:two-f-a-code="twoFACode"
          :account="account"
          :is-logged-in="isLoggedIn"
          :is-account-loading="isAccountLoading"
          :is-revoking="isRevoking"
          :is-clearing="isClearing"
          :is-logging-in="isLoggingIn"
          :login-form="loginForm"
          @revoke="handleRevoke"
          @clear-keychain="handleClearKeychain"
          @refresh="refreshAccount(false)"
          @login="handleLogin"
          @confirm2-f-a="confirm2FA"
          @cancel2-f-a="cancel2FA"
        />

        <!-- VIEW 7: 系统设置 (settings) -->
        <SettingsView
          v-show="activeTab === 'settings'"
          :settings="settings"
          :platform-options="platformOptions"
          :is-testing-proxy="isTestingProxy"
          @save="handleSaveSettings"
          @test-proxy="handleTestProxy"
          @open-download-dir="handleOpenDefaultDownloadDir"
        />

        <!-- VIEW 8: 关于软件 (about) -->
        <AboutView
          v-show="activeTab === 'about'"
          @open-git-hub="openGitHub"
          @copy-git-hub-url="copyGitHubUrl"
        />
      </div>

      <!-- Bottom Minimal Status Bar & Drawer -->
      <MainStatusBar
        :is-any-operation-running="isAnyOperationRunning"
        :status-text="statusText"
        :show-logs="showLogs"
        :log-lines="logLines"
        @cancel="handleCancel"
        @toggle-logs="toggleLogs"
        @clear-logs="clearLogs"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useClipboard } from '@vueuse/core'
import MainSidebar from './layout/MainSidebar.vue'
import MainStatusBar from './layout/MainStatusBar.vue'
import AboutView from './views/AboutView.vue'
import AccountView from './views/AccountView.vue'
import DownloadView from './views/DownloadView.vue'
import InstallerView from './views/InstallerView.vue'
import PurchasedView from './views/PurchasedView.vue'
import SearchView from './views/SearchView.vue'
import SettingsView from './views/SettingsView.vue'
import VersionsView from './views/VersionsView.vue'
import { useAppMessage } from '../ui/feedback'
import { CancelRunningCommand } from '../../wailsjs/go/backend/App'
import { EventsOn, OnFileDrop, OnFileDropOff, BrowserOpenURL } from '../../wailsjs/runtime/runtime'
import { backend as main } from '../../wailsjs/go/models'
import type { AppTheme } from '../ui/theme'

import { useSettings } from '../composables/useSettings'
import { useAccount } from '../composables/useAccount'
import { useDownloads } from '../composables/useDownloads'
import { useInstaller } from '../composables/useInstaller'
import { useSearch } from '../composables/useSearch'
import { useVersions, type VersionItem } from '../composables/useVersions'
import { usePurchased } from '../composables/usePurchased'
import { useConsoleLogs } from '../composables/useConsoleLogs'
import { applyDemoData } from '../composables/useDemoData'

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

const message = useAppMessage()
const { copy } = useClipboard({ legacy: true })

const activeTab = ref('search')
const isAnyOperationRunning = ref(false)
const statusText = ref('就绪')

function onBusyChange(busy: boolean, text?: string) {
  isAnyOperationRunning.value = busy
  if (text) {
    statusText.value = text
  }
}

// 1. Settings Composable
const {
  settings,
  isTestingProxy,
  platformOptions,
  limitOptions,
  loadSettings,
  onProxyToggle,
  handleTestProxy,
  handleSaveSettings,
  handleOpenDefaultDownloadDir
} = useSettings({
  onStatusChange: (status) => { statusText.value = status }
})

// 2. Account Composable
const {
  account,
  isLoggedIn,
  isAccountLoading,
  isRevoking,
  isClearing,
  isLoggingIn,
  loginForm,
  show2FAModal,
  twoFACode,
  refreshAccount,
  handleLogin,
  confirm2FA,
  cancel2FA,
  handleRevoke,
  handleClearKeychain
} = useAccount({
  onSuccessLogin: () => { activeTab.value = 'search' },
  onNeedNavigate: (tab) => { activeTab.value = tab },
  onSettingsReload: loadSettings,
  onBusyChange
})

// 3. Downloads Composable
const {
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
} = useDownloads({
  onTaskAdded: () => { activeTab.value = 'download' }
})

// 4. Installer Composable
const {
  selectedIPAPath,
  selectedIPAFileName,
  devices,
  selectedDeviceUDID,
  selectedDevice,
  deviceOptions,
  isLoadingDevices,
  isInstallingIPA,
  useIPAPath,
  handleSelectIPA,
  loadConnectedDevices,
  handleInstallIPA
} = useInstaller({
  onBusyChange,
  onNavigateToInstaller: () => { activeTab.value = 'installer' }
})

function handleInstallFromTask(outputPath: string) {
  if (!outputPath) {
    message.warning('未找到下载的 IPA 文件路径')
    return
  }
  selectedIPAPath.value = outputPath
  activeTab.value = 'installer'
  statusText.value = `已选定安装包: ${outputPath}`
  if (devices.value.length === 0 && !isLoadingDevices.value) {
    void loadConnectedDevices()
  }
}

// 5. Versions Composable
const {
  isListingVersions,
  isBatchQuerying,
  isTargetQuerying,
  showTargetVersionModal,
  targetVersionInput,
  versionForm,
  versionItems,
  filteredVersions,
  displayedVersionColumns,
  handleListVersions,
  handleBatchQueryClick,
  handleTargetQueryClick,
  confirmStartTargetQuery,
  selectAppForVersions,
  stopAllQueries
} = useVersions({
  effectiveTheme,
  onDownloadVersion: (row: VersionItem, form) => {
    const appName = form.appName || form.bundleId
    const ver = row.displayVersion !== '未查询' ? row.displayVersion : (row.versionId ? `Build ${row.versionId}` : '最新版')
    void addDownloadTask(appName, form.bundleId, form.appId, ver, row.versionId, row.fileSize)
  },
  onBusyChange,
  onNavigateToVersions: () => { activeTab.value = 'versions' }
})

// 6. Search Composable
const {
  isSearching,
  searchForm,
  searchResults,
  displayedSearchColumns,
  handleSearch
} = useSearch({
  effectiveTheme,
  onSelectAppForVersions: selectAppForVersions,
  onDownloadApp: (app: main.AppItem) => {
    void addDownloadTask(app.name, app.bundleID, app.id, app.version || '最新版')
  },
  onBusyChange
})

// 7. Purchased Composable
const {
  isPurchasedLoading,
  purchasedLoadLoaded,
  purchasedLoadTotal,
  purchasedLoadProgress,
  purchasedPage,
  purchasedPageSize,
  purchasedTotal,
  purchasedApps,
  purchasedSearchKeyword,
  filteredPurchasedApps,
  purchasedMatchTotal,
  purchasedPageCount,
  displayedPurchasedColumns,
  purchasedPageSizeOptions,
  loadPurchases,
  onPurchasedPageSizeChange,
  prevPurchasedPage,
  nextPurchasedPage
} = usePurchased({
  effectiveTheme,
  onSelectAppForVersions: selectAppForVersions,
  onDownloadApp: (app: main.AppItem) => {
    void addDownloadTask(app.name, app.bundleID, app.id, app.version || '最新版')
  },
  onBusyChange
})

// 8. Console Logs Composable
const {
  showLogs,
  logLines,
  toggleLogs,
  appendLog,
  clearLogs
} = useConsoleLogs()

function openGitHub() {
  BrowserOpenURL('https://github.com/wnnz/AppleVault')
}

async function copyGitHubUrl() {
  await copy('https://github.com/wnnz/AppleVault')
  message.success('已复制仓库地址到剪贴板！')
}

function handleCancel() {
  stopAllQueries()
  CancelRunningCommand()
  isAnyOperationRunning.value = false
  statusText.value = '已终止当前操作'
}

// Lifecycle
onMounted(() => {
  const urlParams = new URLSearchParams(window.location.search)
  const tabParam = urlParams.get('tab')
  const isDemo = urlParams.get('demo') === '1'

  if (isDemo) {
    applyDemoData({
      tabParam,
      account,
      statusText,
      searchForm,
      settings,
      searchResults,
      versionForm,
      versionItems,
      downloadTasks,
      purchasedApps,
      purchasedTotal,
      devices,
      selectedDeviceUDID,
      selectedIPAPath
    })
  } else {
    loadSettings()
    refreshAccount(true)
    loadDownloadTasks()
  }

  if (tabParam) {
    activeTab.value = tabParam
  }

  OnFileDrop((_x: number, _y: number, paths: string[]) => {
    useIPAPath(paths)
  }, true)

  EventsOn('log', (msg: string) => {
    appendLog(msg)
  })

  EventsOn('download-task-updated', (updatedTask: main.DownloadTask) => {
    updateTask(updatedTask)
  })

  EventsOn('download-tasks-reload', () => {
    loadDownloadTasks()
  })
})

onBeforeUnmount(() => {
  OnFileDropOff()
})
</script>

<style src="./MainView.css"></style>
