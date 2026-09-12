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
      :is-demo-mode="isDemoMode"
      :app-version="appVersion"
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
import { CancelRunningCommand, CheckForUpdates, GetAppVersion } from '../../wailsjs/go/backend/App'
import { BrowserOpenURL, EventsOn, OnFileDrop, OnFileDropOff } from '../../wailsjs/runtime/runtime'
import { backend as main } from '../../wailsjs/go/models'
import type { AppTheme } from '../ui/theme'
import { useAppDialog } from '../ui/feedback'
import { bundledAppVersion, formatAppVersion } from '../ui/version'

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

// 解析启动初始页：优先使用 URL 参数；若未显式指定，根据本地登录凭证状态决定（未登录默认打开账号登录页面）
const urlParams = typeof window !== 'undefined' ? new URLSearchParams(window.location.search) : null
const initialTabParam = urlParams?.get('tab') || null
const isDemoMode = urlParams?.get('demo') === '1'
const storedLoggedIn = typeof window !== 'undefined' ? localStorage.getItem('apple_vault_logged_in') === 'true' : false

const activeTab = ref(initialTabParam || (isDemoMode || storedLoggedIn ? 'search' : 'account'))
const isAnyOperationRunning = ref(false)
const statusText = ref('就绪')
const appVersion = ref(bundledAppVersion)
const dialog = useAppDialog()
const updateCheckStorageKey = 'apple_vault_update_checked_at_v1'
const updateCheckInterval = 24 * 60 * 60 * 1000

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
const currentAccount = ref<main.AccountInfo>({ id: '', name: '', email: '', region: '', active: false, success: false })
const isLoggedIn = computed(() => currentAccount.value.success && !!currentAccount.value.email)
const currentSettings = ref<main.Settings>({
  keychainPassphrase: '',
  themeHintShown: false,
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
  searchViewRef.value?.searchResults.splice(0)
  versionsViewRef.value?.versionItems.splice(0)
  purchasedViewRef.value?.purchasedApps.splice(0)
  if (purchasedViewRef.value) purchasedViewRef.value.purchasedTotal = 0
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

async function loadAppVersion() {
  try {
    appVersion.value = formatAppVersion(await GetAppVersion())
  } catch {
    // Keep the bundled version when the Wails bridge is unavailable (for example, browser preview).
  }
}

async function checkForUpdatesOnStartup() {
  const lastChecked = Number(localStorage.getItem(updateCheckStorageKey) || 0)
  if (Date.now() - lastChecked < updateCheckInterval) return
  try {
    const info = await CheckForUpdates()
    localStorage.setItem(updateCheckStorageKey, String(Date.now()))
    if (!info.available) return
    const notes = (info.releaseNotes || '').trim()
    const summary = notes.length > 500 ? `${notes.slice(0, 500)}…` : notes
    dialog.info({
      title: `发现新版本 v${formatAppVersion(info.latestVersion)}`,
      content: summary || info.releaseName || '新版本已发布，可前往 GitHub 查看并下载。',
      positiveText: '立即查看',
      negativeText: '稍后',
      onPositiveClick: () => {
        if (info.releaseURL) BrowserOpenURL(info.releaseURL)
      }
    })
  } catch {
    // Automatic update checks are best-effort and must not interrupt startup.
  }
}

// 演示模式提示文案字典
const demoStatusByTab: Record<string, string> = {
  search: '搜索完成，找到 6 个应用。',
  versions: '共获取 8 个历史版本记录。',
  download: '当前有 1 个任务正在下载，速度 12.8 MB/s',
  purchased: '已购应用加载完成（共 6 款）。',
  installer: '已就绪，已检测到 iPhone 15 Pro Max',
  account: '已登录 果仓助手用户（applevault.user@icloud.com）',
  settings: '底层引擎就绪，配置已加载',
  about: `果仓助手 (AppleVault) v${bundledAppVersion}`
}

/**
 * 填充演示与截图所需的 Mock 数据
 */
function applyDemoMockData(tabParam: string | null) {
  currentAccount.value = { id: 'demo-cn', name: '果仓助手用户', email: 'applevault.user@icloud.com', region: 'CN', active: true, success: true }
  statusText.value = demoStatusByTab[tabParam || activeTab.value] || '就绪'

  if (searchViewRef.value) {
    searchViewRef.value.searchForm.term = '微信'
    searchViewRef.value.searchResults = [
      { id: 414478124, bundleID: 'com.tencent.xin', name: '微信 (WeChat)', version: '8.0.50', price: 0, displayPrice: '免费' },
      { id: 333206289, bundleID: 'com.alipay.iphoneclient', name: '支付宝 - 生活好 支付宝', version: '10.5.88', price: 0, displayPrice: '免费' },
      { id: 736536022, bundleID: 'tv.danmaku.bilianime', name: '哔哩哔哩 (Bilibili)', version: '7.82.0', price: 0, displayPrice: '免费' },
      { id: 590338362, bundleID: 'com.netease.cloudmusic', name: '网易云音乐', version: '9.0.70', price: 0, displayPrice: '免费' },
      { id: 835599320, bundleID: 'com.zhiliaoapp.musically', name: 'TikTok - 精彩短视频与音乐', version: '35.8.0', price: 0, displayPrice: '免费' },
      { id: 444934666, bundleID: 'com.tencent.mqq', name: 'QQ - 轻松做自己', version: '9.0.65', price: 0, displayPrice: '免费' }
    ]
  }

  if (versionsViewRef.value) {
    versionsViewRef.value.versionForm.bundleId = 'com.tencent.xin'
    versionsViewRef.value.versionForm.appName = '微信'
    versionsViewRef.value.versionItems = [
      { versionId: '868192301', displayVersion: '8.0.50', fileSize: '286.4 MB', releaseDate: '2024-08-15' },
      { versionId: '867204918', displayVersion: '8.0.49', fileSize: '284.1 MB', releaseDate: '2024-07-20' },
      { versionId: '865819021', displayVersion: '8.0.48', fileSize: '279.8 MB', releaseDate: '2024-06-12' },
      { versionId: '864201990', displayVersion: '8.0.47', fileSize: '275.2 MB', releaseDate: '2024-05-08' },
      { versionId: '862901124', displayVersion: '8.0.46', fileSize: '270.5 MB', releaseDate: '2024-04-01' },
      { versionId: '861502391', displayVersion: '8.0.45', fileSize: '268.0 MB', releaseDate: '2024-03-05' },
      { versionId: '859810234', displayVersion: '8.0.44', fileSize: '263.8 MB', releaseDate: '2024-01-22' },
      { versionId: '858201992', displayVersion: '8.0.43', fileSize: '260.1 MB', releaseDate: '2023-12-18' }
    ]
  }

  if (downloadViewRef.value) {
    downloadViewRef.value.downloadTasks = [
      { id: '1', appName: '微信 (WeChat)', bundleID: 'com.tencent.xin', appId: 414478124, version: '8.0.50', versionId: '868192301', fileSize: '286.4 MB', totalBytes: 300312000, currBytes: 195202800, progress: 65, speed: '12.8 MB/s', status: 'downloading', outputPath: '', errorMessage: '', createdAt: '2024-08-20 15:30:12' },
      { id: '2', appName: '支付宝', bundleID: 'com.alipay.iphoneclient', appId: 333206289, version: '10.5.88', versionId: '865001129', fileSize: '142.0 MB', totalBytes: 148897000, currBytes: 148897000, progress: 100, speed: '已完成', status: 'completed', outputPath: 'data/downloads/user@icloud.com/alipay.ipa', errorMessage: '', createdAt: '2024-08-20 15:24:05' },
      { id: '3', appName: 'Infuse · 精彩影音播放器', bundleID: 'com.firecore.infuse', appId: 1136220934, version: '7.7.2', versionId: '863920191', fileSize: '118.5 MB', totalBytes: 124256256, currBytes: 124256256, progress: 100, speed: '已完成', status: 'completed', outputPath: 'data/downloads/user@icloud.com/infuse.ipa', errorMessage: '', createdAt: '2024-08-20 14:10:33' },
      { id: '4', appName: '网易云音乐', bundleID: 'com.netease.cloudmusic', appId: 590338362, version: '9.0.70', versionId: '864201991', fileSize: '215.3 MB', totalBytes: 225758413, currBytes: 0, progress: 0, speed: '等待中', status: 'pending', outputPath: '', errorMessage: '', createdAt: '2024-08-20 15:31:00' }
    ]
  }

  if (purchasedViewRef.value) {
    purchasedViewRef.value.purchasedApps = [
      { id: 414478124, bundleID: 'com.tencent.xin', name: '微信 (WeChat)', version: '8.0.50', price: 0, purchaseDate: '2024-01-15 10:20', displayPrice: '已购' },
      { id: 333206289, bundleID: 'com.alipay.iphoneclient', name: '支付宝', version: '10.5.88', price: 0, purchaseDate: '2024-02-08 14:12', displayPrice: '已购' },
      { id: 1136220934, bundleID: 'com.firecore.infuse', name: 'Infuse • 精彩影音播放器', version: '7.7.2', price: 0, purchaseDate: '2024-03-22 09:45', displayPrice: '已购' },
      { id: 916364737, bundleID: 'com.procreate.pocket', name: 'Procreate Pocket', version: '4.0.11', price: 0, purchaseDate: '2024-04-10 18:30', displayPrice: '已购' },
      { id: 1596487405, bundleID: 'com.taguirov.adam.WebDAV', name: 'WebDAV Manager', version: '2.1.0', price: 0, purchaseDate: '2024-05-19 16:30', displayPrice: '已购' },
      { id: 736536022, bundleID: 'tv.danmaku.bilianime', name: '哔哩哔哩 (Bilibili)', version: '7.82.0', price: 0, purchaseDate: '2024-06-01 21:05', displayPrice: '已购' }
    ]
    purchasedViewRef.value.purchasedTotal = 6
  }

  if (installerViewRef.value) {
    installerViewRef.value.devices = [
      { udid: '00008130-001A49021E28001C', name: 'iPhone 15 Pro Max', productType: 'iPhone 15 Pro Max', productVersion: '17.5.1', connectionType: 'USB 3.0' }
    ]
    installerViewRef.value.selectedDeviceUDID = '00008130-001A49021E28001C'
    installerViewRef.value.selectedIPAPath = 'D:\\Dev\\AppleVault\\data\\downloads\\applevault.user@icloud.com\\WeChat_8.0.50.ipa'
    installerViewRef.value.ipaInspection = {
      success: true,
      appName: '微信',
      bundleID: 'com.tencent.xin',
      version: '8.0.50',
      buildVersion: '868192301',
      minimumOSVersion: '13.0',
      supportedPlatforms: ['iPhoneOS'],
      supportedDeviceTypes: ['iPhone/iPod', 'iPad'],
      signed: true,
      fileSize: 300312000,
      displayFileSize: '286.4 MB',
      compatible: true,
      compatibilityMessage: 'IPA 完整性、签名结构和设备兼容性检查通过'
    }
  }

  if (accountViewRef.value) {
    accountViewRef.value.account = { id: 'demo-cn', name: '果仓助手用户', email: 'applevault.user@icloud.com', region: 'CN', active: true, success: true }
    accountViewRef.value.accounts = [
      accountViewRef.value.account,
      { id: 'demo-us', name: 'AppleVault US', email: 'applevault.us@icloud.com', region: 'US', active: false, success: true }
    ]
  }

  if (settingsViewRef.value) {
    settingsViewRef.value.settings.proxyUrl = 'http://127.0.0.1:7890'
    settingsViewRef.value.settings.defaultDownloadDir = 'D:\\Dev\\AppleVault\\data\\downloads\\applevault.user'
  }
}

// 生命周期与 Wails 系统事件
onMounted(() => {
  const urlParams = new URLSearchParams(window.location.search)
  const tabParam = urlParams.get('tab')
  const isDemo = urlParams.get('demo') === '1'

  if (tabParam) {
    activeTab.value = tabParam
  }

  if (isDemo) {
    applyDemoMockData(tabParam)
  } else {
    void loadAppVersion()
    void checkForUpdatesOnStartup()
    void settingsViewRef.value?.loadSettings()
    void accountViewRef.value?.refreshAccount(!tabParam)
    void downloadViewRef.value?.loadDownloadTasks()
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
