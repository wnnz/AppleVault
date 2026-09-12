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
    <aside class="app-sidebar">
      <!-- Brand Header -->
      <div class="sidebar-brand">
        <div class="brand-icon-box">
          <img
            class="brand-logo-image"
            :src="standardLogo"
            alt="AppleVault"
          />
        </div>
        <div class="brand-text">
          <div class="brand-title-row">
            <div class="brand-name">果仓助手</div>
            <div class="brand-tag">v1.1</div>
          </div>
          <div class="brand-sub">AppleVault</div>
        </div>
      </div>

      <!-- Navigation List -->
      <nav class="sidebar-nav">
        <div class="nav-section-title">
          <span>核心功能</span>
        </div>
        <button
          class="nav-item"
          :class="{ active: activeTab === 'search' }"
          @click="activeTab = 'search'"
        >
          <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="search" /></span>
          <span class="nav-label">应用搜索</span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'versions' }"
          @click="activeTab = 'versions'"
        >
          <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="history" /></span>
          <span class="nav-label">历史版本</span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'download' }"
          @click="activeTab = 'download'"
        >
          <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="download" /></span>
          <span class="nav-label">下载中心</span>
          <span v-if="activeTaskCount > 0" class="nav-badge">{{ activeTaskCount }}</span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'purchased' }"
          @click="activeTab = 'purchased'"
        >
          <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="purchased" /></span>
          <span class="nav-label">已购应用</span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'installer' }"
          @click="activeTab = 'installer'"
        >
          <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="device" /></span>
          <span class="nav-label">设备直装</span>
        </button>

        <div class="nav-section-title mt-4">
          <span>偏好与设置</span>
        </div>
        <button
          class="nav-item"
          :class="{ active: activeTab === 'account' }"
          @click="activeTab = 'account'"
        >
          <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="account" /></span>
          <span class="nav-label">账号中心</span>
          <span class="account-dot" :class="isLoggedIn ? 'dot-online' : 'dot-offline'"></span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'settings' }"
          @click="activeTab = 'settings'"
        >
          <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="settings" /></span>
          <span class="nav-label">系统设置</span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'about' }"
          @click="activeTab = 'about'"
        >
          <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="about" /></span>
          <span class="nav-label">关于软件</span>
        </button>
      </nav>

      <div v-if="effectiveTheme === 'handdrawn'" class="handdrawn-sidebar-floating-leaf" aria-hidden="true">
        <img :src="handdrawnSidebarFloatingLeaf" alt="" />
      </div>

      <div v-if="effectiveTheme === 'handdrawn'" class="handdrawn-sidebar-art" aria-hidden="true">
        <img :src="handdrawnSidebarArt" alt="" />
      </div>

      <!-- Sidebar Footer (Account & Quick Controls) -->
      <div class="sidebar-footer">

        <!-- Account Quick Card -->
        <div class="sidebar-account-card" @click="activeTab = 'account'">
          <div class="user-avatar" :class="{ 'avatar-logged': isLoggedIn }">
            {{ isLoggedIn ? (account.name ? account.name.charAt(0).toUpperCase() : '') : '?' }}
          </div>
          <div class="user-info-text">
            <div class="user-name text-ellipsis">{{ account.name || '未登录 Apple ID' }}</div>
            <div class="user-email text-ellipsis">{{ account.email || '点击前往登录' }}</div>
          </div>
        </div>

        <!-- Quick Switch Bar -->
        <div class="sidebar-actions-bar">
          <!-- Proxy Switch -->
          <div class="quick-tool" title="网络代理">
            <AppSwitch
              v-model="settings.enableProxy"
              size="small"
              @update:model-value="onProxyToggle"
            />
            <span class="quick-tool-label">代理</span>
          </div>

          <!-- Theme Switcher (靠右显示) -->
          <div class="theme-switcher-wrapper">
            <AppPopover
              v-model="showThemePopover"
              placement="top-end"
            >
              <template #trigger>
                <button
                  class="icon-action-btn theme-action-btn"
                  :class="{ active: showThemePopover }"
                  title="选择外观主题"
                >
                  <!-- 调色板/主题 图标 -->
                  <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="13.5" cy="6.5" r=".5" fill="currentColor"></circle>
                    <circle cx="17.5" cy="10.5" r=".5" fill="currentColor"></circle>
                    <circle cx="8.5" cy="7.5" r=".5" fill="currentColor"></circle>
                    <circle cx="6.5" cy="12.5" r=".5" fill="currentColor"></circle>
                    <path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.648-.746 1.648-1.688 0-.437-.18-.835-.437-1.125-.29-.289-.438-.652-.438-1.125a1.64 1.64 0 0 1 1.668-1.668h1.996c3.051 0 5.563-2.512 5.563-5.563C22 6.5 17.5 2 12 2z"></path>
                  </svg>
                </button>
              </template>
              <div
                class="theme-popover-menu"
                :class="[
                  `theme-${effectiveTheme}`,
                  isDarkTheme ? 'is-dark' : 'is-light'
                ]"
              >
                <div class="theme-popover-header">
                  <div class="theme-popover-title">
                    <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 5px;">
                      <circle cx="13.5" cy="6.5" r=".5" fill="currentColor"></circle>
                      <circle cx="17.5" cy="10.5" r=".5" fill="currentColor"></circle>
                      <circle cx="8.5" cy="7.5" r=".5" fill="currentColor"></circle>
                      <circle cx="6.5" cy="12.5" r=".5" fill="currentColor"></circle>
                      <path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.648-.746 1.648-1.688 0-.437-.18-.835-.437-1.125-.29-.289-.438-.652-.438-1.125a1.64 1.64 0 0 1 1.668-1.668h1.996c3.051 0 5.563-2.512 5.563-5.563C22 6.5 17.5 2 12 2z"></path>
                    </svg>
                    选择外观主题
                  </div>
                </div>
                
                <div class="theme-options-list">
                  <div
                    v-for="item in themeList"
                    :key="item.key"
                    class="theme-select-item"
                    :class="{ active: effectiveTheme === item.key }"
                    @click="selectTheme(item.key)"
                  >
                    <div class="theme-color-preview" :style="{ background: item.preview.bg, borderColor: item.preview.border }">
                      <div class="theme-color-card" :style="{ background: item.preview.card, borderColor: item.preview.border }">
                        <span class="theme-color-dot" :style="{ background: item.preview.accent }"></span>
                      </div>
                    </div>
                    <div class="theme-info-box">
                      <div class="theme-item-name">
                        <span>{{ item.name }}</span>
                        <span v-if="item.recommended" class="theme-thumb-badge" title="精选推荐">
                          <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor" aria-hidden="true">
                            <path d="M1 21h4V9H1v12zm22-11c0-1.1-.9-2-2-2h-6.31l.95-4.57.03-.32c0-.41-.17-.79-.44-1.06L14.17 1 7.59 7.59C7.22 7.95 7 8.45 7 9v10c0 1.1.9 2 2 2h9c.83 0 1.54-.5 1.84-1.22l3.02-7.05c.09-.23.14-.47.14-.73v-2z"></path>
                          </svg>
                        </span>
                      </div>
                      <div class="theme-item-desc">{{ item.description }}</div>
                    </div>
                    <div v-if="effectiveTheme === item.key" class="theme-check-icon">
                      <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                        <polyline points="20 6 9 17 4 12"></polyline>
                      </svg>
                    </div>
                  </div>
                </div>
              </div>
            </AppPopover>
          </div>
        </div>
      </div>
    </aside>

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
            @click="handleCancel"
            class="mr-2"
          >
            终止操作
          </AppButton>

          <button class="status-btn" @click="toggleLogs">
            {{ showLogs ? '收起日志' : '实时日志' }}
          </button>
        </div>
      </div>

      <!-- Collapsible Sleek Terminal Drawer -->
      <div v-show="showLogs" class="sleek-console-drawer">
        <div class="console-action-bar">
          <div class="console-title">实时执行日志</div>
          <div class="console-btns">
            <button class="console-bar-btn" @click="clearLogs">清空</button>
            <button class="console-bar-btn" @click="toggleLogs">收起</button>
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
    </section>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick, h, defineComponent } from 'vue'
import { useClipboard } from '@vueuse/core'
import AppButton from '../ui/components/AppButton.vue'
import AppPopover from '../ui/components/AppPopover.vue'
import AppSwitch from '../ui/components/AppSwitch.vue'
import AboutView from './views/AboutView.vue'
import AccountView from './views/AccountView.vue'
import DownloadView from './views/DownloadView.vue'
import InstallerView from './views/InstallerView.vue'
import PurchasedView from './views/PurchasedView.vue'
import SearchView from './views/SearchView.vue'
import SettingsView from './views/SettingsView.vue'
import VersionsView from './views/VersionsView.vue'
import { useAppDialog, useAppMessage } from '../ui/feedback'
import {
  AddDownloadTask,
  GetDownloadTasks,
  CancelDownloadTask,
  DeleteDownloadTask,
  ClearCompletedDownloadTasks,
  GetAccountInfo,
  Login,
  Revoke,
  ClearKeychainCache,
  Search,
  ListVersions,
  GetVersionMetadata,
  Purchase,
  ListPurchases,
  GetSettings,
  SaveSettings,
  SelectDirectory,
  TestProxy,
  OpenInExplorer,
  CancelRunningCommand,
  SelectIPA,
  ListDevices,
  InstallIPA
} from '../../wailsjs/go/backend/App'
import { EventsOn, OnFileDrop, OnFileDropOff, BrowserOpenURL } from '../../wailsjs/runtime/runtime'
import { backend as main } from '../../wailsjs/go/models'
import { type AppTheme, themeOptions } from '../ui/theme'
import standardLogo from '../assets/applevault-logo.webp'
import handdrawnSidebarArt from '../ui/themes/handdrawn/assets/sketch-sidebar-art.svg'
import handdrawnSidebarFloatingLeaf from '../ui/themes/handdrawn/assets/sketch-floating-leaf.svg'

const sketchNavPaths: Record<string, string[]> = {
  search: ['M10.8 4.2a6.6 6.6 0 1 0 0 13.2 6.6 6.6 0 0 0 0-13.2Z', 'm15.8 15.8 4 4'],
  history: ['M12 4.1a7.9 7.9 0 1 1-5.8 2.5', 'M4.2 4.6v4h4', 'M12 7.7v4.8l3.2 2'],
  download: ['M12 3.5v11', 'm7.8 11.2 4.2 4.1 4.2-4.1', 'M4.5 18v2h15v-2'],
  purchased: ['M5.5 8h13l-.5 12H6L5.5 8Z', 'M8.2 8V6.5a3.8 3.8 0 0 1 7.6 0V8'],
  device: ['M7.2 3.2h9.6c1 0 1.7.8 1.7 1.8v14c0 1-.7 1.8-1.7 1.8H7.2c-1 0-1.7-.8-1.7-1.8V5c0-1 .7-1.8 1.7-1.8Z', 'M10 6h4', 'M11 18h2'],
  account: ['M12 4.2a3.8 3.8 0 1 0 0 7.6 3.8 3.8 0 0 0 0-7.6Z', 'M5.2 20c.7-4 3.2-6.1 6.8-6.1s6.1 2.1 6.8 6.1'],
  settings: ['M12 8.4a3.6 3.6 0 1 0 0 7.2 3.6 3.6 0 0 0 0-7.2Z', 'M9.8 3.5 9.2 5a7.6 7.6 0 0 0-1.8 1l-1.5-.6-1.6 2.8 1.2 1a7.8 7.8 0 0 0-.1 2.1l-1.4.9.8 3.1 1.7-.1a7.5 7.5 0 0 0 1.5 1.4l-.2 1.7 3.1.9.9-1.4a7.3 7.3 0 0 0 2-.2l1.1 1.3 2.8-1.6-.6-1.5a7.6 7.6 0 0 0 1-1.8l1.6-.5-.1-3.2-1.6-.4a7.7 7.7 0 0 0-1.1-1.8l.5-1.6-2.8-1.5-1 1.3a7.7 7.7 0 0 0-2-.1l-.9-1.4Z'],
  about: ['M12 3.8a8.2 8.2 0 1 0 0 16.4 8.2 8.2 0 0 0 0-16.4Z', 'M12 10.5v5.8', 'M12 7.3h.01'],
}

const SketchNavIcon = defineComponent({
  name: 'SketchNavIcon',
  props: { name: { type: String, required: true } },
  setup(props) {
    return () => {
      const paths = sketchNavPaths[props.name] || sketchNavPaths.about
      const makePaths = (opacity: number, transform?: string) =>
        h('g', { opacity, transform }, paths.map((d) => h('path', { d })))
      return h('svg', {
        class: 'nav-icon-image sketch-nav-icon',
        viewBox: '0 0 24 24',
        fill: 'none',
        stroke: 'currentColor',
        'stroke-width': '1.75',
        'stroke-linecap': 'round',
        'stroke-linejoin': 'round',
      }, [makePaths(1), makePaths(0.28, 'translate(.35 .25)')])
    }
  },
})

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

const showThemePopover = ref(false)

const effectiveTheme = computed<AppTheme>(() => props.currentTheme || (props.isDark ? 'minimal-dark' : 'minimal-light'))
const isDarkTheme = computed(() => effectiveTheme.value.endsWith('-dark') || props.isDark)

const themeList = themeOptions

function selectTheme(themeKey: AppTheme) {
  emit('update:currentTheme', themeKey)
  showThemePopover.value = false
}

const message = useAppMessage()
const dialog = useAppDialog()
const { copy } = useClipboard({ legacy: true })

const activeTab = ref('search')
const isAnyOperationRunning = ref(false)
const statusText = ref('就绪')

// Settings
const settings = ref<main.Settings>({
  keychainPassphrase: '123456',
  defaultDownloadDir: 'data/downloads/default',
  defaultPlatform: 'iphone',
  enableProxy: true,
  proxyUrl: 'http://127.0.0.1:10808',
  ipaToolPath: ''
})

const platformOptions = [
  { label: 'iPhone (iOS)', value: 'iphone' },
  { label: 'iPad (iPadOS)', value: 'ipad' },
  { label: 'Apple TV (tvOS)', value: 'appletv' },
  { label: 'VisionOS', value: 'visionos' }
]

const limitOptions = [
  { label: '5 个', value: 5 },
  { label: '10 个', value: 10 },
  { label: '20 个', value: 20 },
  { label: '50 个', value: 50 }
]

const purchasedPageSizeOptions = [
  { label: '20', value: 20 },
  { label: '50', value: 50 },
  { label: '100', value: 100 }
]

// Account State
const account = ref<main.AccountInfo>({ name: '', email: '', success: false })
const isLoggedIn = computed(() => account.value.success && !!account.value.email)
const isAccountLoading = ref(false)
const isRevoking = ref(false)
const isClearing = ref(false)
const isLoggingIn = ref(false)

const loginForm = ref({
  email: '',
  password: ''
})

// 2FA Modal
const show2FAModal = ref(false)
const twoFACode = ref('')

// Search State
const isSearching = ref(false)
const searchForm = ref({
  term: '',
  limit: 10,
  platform: 'iphone'
})
const searchResults = ref<main.AppItem[]>([])

const searchColumns = [
  { title: '应用名称', key: 'name', minWidth: 160, ellipsis: { tooltip: true } },
  { title: 'Bundle ID', key: 'bundleID', minWidth: 180, ellipsis: { tooltip: true } },
  { title: 'App ID', key: 'id', width: 100 },
  { title: '最新版本', key: 'version', width: 90 },
  { title: '价格', key: 'displayPrice', width: 80 },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    fixed: 'right' as const,
    render(row: main.AppItem) {
      return h('div', { class: 'app-button-group' }, [
        h(AppButton, { size: 'tiny', secondary: true, onClick: () => selectAppForVersions(row) }, () => '历史版本'),
        h(AppButton, { size: 'tiny', type: 'primary', onClick: () => downloadFromSearch(row) }, () => '下载'),
        h(AppButton, { size: 'tiny', secondary: true, onClick: () => handlePurchaseApp(row.bundleID) }, () => '获取许可')
      ])
    }
  }
]

const displayedSearchColumns = computed(() => {
  if (effectiveTheme.value !== 'handdrawn') return searchColumns

  const widths = [196, 163, 94, 87, 83, 217]
  return searchColumns.map((column, index) => ({
    ...column,
    minWidth: undefined,
    width: widths[index]
  }))
})

// Versions State
const isListingVersions = ref(false)
const isBatchQuerying = ref(false)
const shouldStopBatchQuery = ref(false)
const isTargetQuerying = ref(false)
const shouldStopTargetQuery = ref(false)
const showTargetVersionModal = ref(false)
const targetVersionInput = ref('')
const versionForm = ref({
  bundleId: '',
  appId: 0,
  appName: '',
  filter: ''
})

interface VersionItem {
  versionId: string
  displayVersion: string
  fileSize: string
  releaseDate: string
  isQuerying?: boolean
}

function parseVersionParts(v: string): number[] {
  if (!v || v === '未查询') return []
  const clean = v.trim().toLowerCase().replace(/^v/, '')
  const match = clean.match(/^(\d+(?:\.\d+)*)/)
  if (!match) return []
  return match[1].split('.').map(num => parseInt(num, 10))
}

function compareVersions(v1: string, v2: string): number {
  const p1 = parseVersionParts(v1)
  const p2 = parseVersionParts(v2)
  if (p1.length === 0 || p2.length === 0) {
    const clean1 = v1.trim().toLowerCase().replace(/^v/, '')
    const clean2 = v2.trim().toLowerCase().replace(/^v/, '')
    if (clean1 === clean2) return 0
    return clean1 > clean2 ? 1 : -1
  }
  const maxLen = Math.max(p1.length, p2.length)
  for (let i = 0; i < maxLen; i++) {
    const num1 = p1[i] || 0
    const num2 = p2[i] || 0
    if (num1 > num2) return 1
    if (num1 < num2) return -1
  }
  return 0
}

function isVersionMatch(actual: string, target: string): boolean {
  if (!actual || actual === '未查询') return false
  const a = actual.trim().toLowerCase().replace(/^v/, '')
  const t = target.trim().toLowerCase().replace(/^v/, '')
  if (a === t) return true
  return compareVersions(actual, target) === 0
}

const versionItems = ref<VersionItem[]>([])

async function querySingleVersionMetadata(row: VersionItem) {
  row.isQuerying = true
  isAnyOperationRunning.value = true
  statusText.value = `正在查询版本 ID ${row.versionId} 的详情...`

  try {
    const res = await GetVersionMetadata(versionForm.value.bundleId, row.versionId, versionForm.value.appId)
    row.displayVersion = res.displayVersion
    row.fileSize = res.displayFileSize
    row.releaseDate = res.releaseDate ? res.releaseDate.replace(/T.*/, '') : '-'
    versionItems.value = [...versionItems.value]
    statusText.value = `构建 ID ${row.versionId} 对应版本: ${res.displayVersion} (${res.displayFileSize})`
  } catch (err: any) {
    message.error(`查询详情失败: ${err}`)
  } finally {
    row.isQuerying = false
    isAnyOperationRunning.value = false
  }
}

async function downloadFromVersions(row: VersionItem) {
  const appName = versionForm.value.appName || versionForm.value.bundleId
  const ver = row.displayVersion !== '未查询' ? row.displayVersion : (row.versionId ? `Build ${row.versionId}` : '最新版')
  try {
    await AddDownloadTask(
      appName,
      versionForm.value.bundleId,
      versionForm.value.appId,
      ver,
      row.versionId,
      row.fileSize
    )
    message.success(`已添加任务: ${appName} (${ver})`)
    activeTab.value = 'download'
  } catch (err: any) {
    message.error(`添加下载失败: ${err}`)
  }
}

const filteredVersions = computed(() => {
  const f = versionForm.value.filter.trim().toLowerCase()
  if (!f) return versionItems.value
  return versionItems.value.filter(item =>
    item.versionId.toLowerCase().includes(f) ||
    item.displayVersion.toLowerCase().includes(f) ||
    item.fileSize.toLowerCase().includes(f) ||
    item.releaseDate.toLowerCase().includes(f)
  )
})

const versionColumns = [
  { title: '构建 ID', key: 'versionId', minWidth: 160 },
  {
    title: '版本号',
    key: 'displayVersion',
    width: 120,
    render(row: VersionItem) {
      if (row.displayVersion === '未查询') {
        return h('span', { class: 'text-dim italic' }, '未查询')
      }
      return h('span', { class: 'tag-version-highlight' }, row.displayVersion)
    }
  },
  {
    title: '文件体积',
    key: 'fileSize',
    width: 110,
    render(row: VersionItem) {
      return h('span', row.fileSize && row.fileSize !== '-' ? row.fileSize : '-')
    }
  },
  {
    title: '发布日期',
    key: 'releaseDate',
    width: 120,
    render(row: VersionItem) {
      return h('span', row.releaseDate && row.releaseDate !== '-' ? row.releaseDate : '-')
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 160,
    fixed: 'right' as const,
    render(row: VersionItem) {
      return h('div', { class: 'app-button-group' }, [
        h(AppButton, {
          size: 'tiny',
          secondary: true,
          disabled: isListingVersions.value || isBatchQuerying.value || isTargetQuerying.value,
          loading: row.isQuerying,
          onClick: () => querySingleVersionMetadata(row)
        }, () => '查详情'),
        h(AppButton, {
          size: 'tiny',
          type: 'primary',
          onClick: () => downloadFromVersions(row)
        }, () => '下载')
      ])
    }
  }
]

// Download Tasks Manager State
const downloadTasks = ref<main.DownloadTask[]>([])
const activeTaskCount = computed(() =>
  downloadTasks.value.filter(t => t.status === 'downloading' || t.status === 'pending').length
)
const completedTaskCount = computed(() =>
  downloadTasks.value.filter(t => t.status === 'completed').length
)

// Purchased State
const isPurchasedLoading = ref(false)
const purchasedLoadLoaded = ref(0)
const purchasedLoadTotal = ref(0)
const purchasedPage = ref(1)
const purchasedPageSize = ref(20)
const purchasedTotal = ref(0)
const purchasedApps = ref<main.AppItem[]>([])
const purchasedSearchKeyword = ref('')

const purchasedLoadProgress = computed(() => {
  if (!purchasedLoadTotal.value) return 0
  return Math.min(100, Math.round((purchasedLoadLoaded.value / purchasedLoadTotal.value) * 100))
})

const matchingPurchasedApps = computed(() => {
  const kw = purchasedSearchKeyword.value.trim().toLowerCase()
  if (!kw) return purchasedApps.value
  return purchasedApps.value.filter(app => {
    const nameMatch = Boolean(app.name && app.name.toLowerCase().includes(kw))
    const bundleMatch = Boolean(app.bundleID && app.bundleID.toLowerCase().includes(kw))
    const idMatch = Boolean(app.id && String(app.id).includes(kw))
    return nameMatch || bundleMatch || idMatch
  })
})

const purchasedMatchTotal = computed(() => matchingPurchasedApps.value.length)
const purchasedPageCount = computed(() => Math.max(1, Math.ceil(purchasedMatchTotal.value / purchasedPageSize.value)))
const filteredPurchasedApps = computed(() => {
  const start = (purchasedPage.value - 1) * purchasedPageSize.value
  return matchingPurchasedApps.value.slice(start, start + purchasedPageSize.value)
})

watch(purchasedSearchKeyword, () => {
  purchasedPage.value = 1
})

// IPA Installer State
const selectedIPAPath = ref('')
const devices = ref<main.DeviceInfo[]>([])
const selectedDeviceUDID = ref<string | null>(null)
const isLoadingDevices = ref(false)
const isInstallingIPA = ref(false)
const selectedIPAFileName = computed(() => selectedIPAPath.value.split(/[\\/]/).pop() || selectedIPAPath.value)
const selectedDevice = computed(() => devices.value.find(device => device.udid === selectedDeviceUDID.value))
const deviceOptions = computed(() => devices.value.map(device => {
  const modelShort = device.productType
  const osPrefix = device.productType.toLowerCase().includes('ipad') ? 'iPadOS' : 'iOS'
  const versionStr = device.productVersion ? `${osPrefix} ${device.productVersion}` : ''
  const details = (device.name === device.productType
    ? [versionStr, device.connectionType]
    : [modelShort, versionStr, device.connectionType]
  ).filter(Boolean).join(' · ')
  return {
    label: details ? `${device.name} (${details})` : device.name,
    value: device.udid
  }
}))

function formatReadableDate(raw?: string): string {
  if (!raw) return '-'
  return raw.replace(/T/, ' ').replace(/:\d{2}Z$/, '').replace(/Z$/, '')
}

const purchasedColumns = [
  { title: '应用名称', key: 'name', minWidth: 160, ellipsis: { tooltip: true } },
  { title: 'Bundle ID', key: 'bundleID', minWidth: 180, ellipsis: { tooltip: true } },
  { title: 'App ID', key: 'id', width: 100 },
  {
    title: '获取时间',
    key: 'purchaseDate',
    width: 140,
    render(row: main.AppItem) {
      return h('span', { class: 'text-dim' }, formatReadableDate(row.purchaseDate))
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    fixed: 'right' as const,
    render(row: main.AppItem) {
      return h('div', { class: 'app-button-group' }, [
        h(AppButton, { size: 'tiny', secondary: true, onClick: () => selectAppForVersions(row) }, () => '历史版本'),
        h(AppButton, { size: 'tiny', type: 'primary', onClick: () => downloadFromPurchased(row) }, () => '下载')
      ])
    }
  }
]

const displayedVersionColumns = computed(() => {
  if (effectiveTheme.value !== 'handdrawn') return versionColumns

  const widths = [314, 139, 101, 129, 155]
  return versionColumns.map((column, index) => ({
    ...column,
    minWidth: undefined,
    width: widths[index]
  }))
})

const displayedPurchasedColumns = computed(() => {
  if (effectiveTheme.value !== 'handdrawn') return purchasedColumns

  const widths = [227, 200, 104, 153, 156]
  return purchasedColumns.map((column, index) => ({
    ...column,
    minWidth: undefined,
    width: widths[index]
  }))
})

// Settings State
const isTestingProxy = ref(false)

// Logs State
const showLogs = ref<boolean>(localStorage.getItem('apple_vault_show_logs') === 'true')
const logLines = ref<string[]>([])
const logContainerRef = ref<HTMLElement | null>(null)

function toggleLogs() {
  showLogs.value = !showLogs.value
  localStorage.setItem('apple_vault_show_logs', String(showLogs.value))
}

function appendLog(line: string) {
  logLines.value.push(line)
  if (logLines.value.length > 600) {
    logLines.value.shift()
  }
  nextTick(() => {
    if (logContainerRef.value) {
      logContainerRef.value.scrollTop = logContainerRef.value.scrollHeight
    }
  })
}

function clearLogs() {
  logLines.value = []
}

// --- Methods ---

async function loadSettings() {
  try {
    const s = await GetSettings()
    settings.value = s
  } catch (err: any) {
    console.error(err)
  }
}

async function onProxyToggle(val: boolean) {
  settings.value.enableProxy = val
  await SaveSettings(settings.value)
  message.info(val ? '已开启网络代理' : '已关闭网络代理')
}

async function refreshAccount(autoNavigate = false) {
  isAccountLoading.value = true
  statusText.value = '正在获取账号信息...'
  try {
    const res = await GetAccountInfo()
    account.value = res
    if (res.success && res.email) {
      statusText.value = `已登录: ${res.name} (${res.email})`
      await loadSettings()
      if (autoNavigate) {
        activeTab.value = 'search'
      }
    } else {
      statusText.value = '未检测到已登录的 Apple ID'
      activeTab.value = 'account'
    }
  } catch (err: any) {
    account.value = { name: '', email: '', success: false }
    statusText.value = '未登录'
    activeTab.value = 'account'
  } finally {
    isAccountLoading.value = false
  }
}

async function handleLogin() {
  if (!loginForm.value.email || !loginForm.value.password) {
    message.warning('请输入 Apple ID 邮箱和密码！')
    return
  }

  isLoggingIn.value = true
  isAnyOperationRunning.value = true
  statusText.value = '正在验证 Apple ID 账号与密码...'

  try {
    const res = await Login(loginForm.value.email, loginForm.value.password, '')
    if (res.requires2FA) {
      twoFACode.value = ''
      show2FAModal.value = true
      statusText.value = '等待输入双重认证验证码...'
      return
    }

    if (res.success) {
      account.value = res.account
      message.success(`登录成功: ${res.account.name}`)
      loginForm.value.password = ''
      statusText.value = `登录成功: ${res.account.name}`
      await loadSettings()
      activeTab.value = 'search'
    } else {
      message.error(`登录失败: ${res.errorMessage}`)
      statusText.value = `登录失败: ${res.errorMessage}`
    }
  } catch (err: any) {
    message.error(`执行异常: ${err}`)
  } finally {
    isLoggingIn.value = false
    isAnyOperationRunning.value = false
  }
}

async function confirm2FA() {
  if (twoFACode.value.length !== 6) {
    message.warning('请输入正确的 6 位数字验证码！')
    return
  }

  isLoggingIn.value = true
  isAnyOperationRunning.value = true
  statusText.value = '正在提交验证码并完成登录...'

  try {
    const res = await Login(loginForm.value.email, loginForm.value.password, twoFACode.value)
    if (res.success) {
      show2FAModal.value = false
      account.value = res.account
      message.success(`登录成功: ${res.account.name}`)
      loginForm.value.password = ''
      twoFACode.value = ''
      statusText.value = `登录成功: ${res.account.name}`
      await loadSettings()
      activeTab.value = 'search'
    } else {
      message.error(`验证失败: ${res.errorMessage}`)
      statusText.value = `验证失败: ${res.errorMessage}`
    }
  } catch (err: any) {
    message.error(`验证异常: ${err}`)
  } finally {
    isLoggingIn.value = false
    isAnyOperationRunning.value = false
  }
}

function cancel2FA() {
  show2FAModal.value = false
  twoFACode.value = ''
  statusText.value = '用户取消了验证。'
}

async function handleRevoke() {
  dialog.warning({
    title: '确认退出登录',
    content: '确定要退出当前 Apple ID 账号并清除本地授权凭证吗？',
    positiveText: '确定退出',
    negativeText: '取消',
    onPositiveClick: async () => {
      isRevoking.value = true
      isAnyOperationRunning.value = true
      try {
        const ok = await Revoke()
        if (ok) {
          account.value = { name: '', email: '', success: false }
          message.success('已成功注销登录凭据')
          statusText.value = '已退出登录'
          await loadSettings()
          activeTab.value = 'account'
        }
      } catch (err: any) {
        message.error(`注销失败: ${err}`)
      } finally {
        isRevoking.value = false
        isAnyOperationRunning.value = false
      }
    }
  })
}

async function handleClearKeychain() {
  dialog.warning({
    title: '清空本地密钥库缓存',
    content: '此操作将删除本地存储的 .ipatool 密钥数据并重置缓存。确定清理吗？',
    positiveText: '确定清理',
    negativeText: '取消',
    onPositiveClick: async () => {
      isClearing.value = true
      try {
        await ClearKeychainCache()
        message.success('本地密钥缓存已清除')
        await refreshAccount()
      } catch (err: any) {
        message.error(`清理失败: ${err}`)
      } finally {
        isClearing.value = false
      }
    }
  })
}

// Search
async function handleSearch() {
  if (!searchForm.value.term) {
    message.warning('请输入搜索关键词')
    return
  }

  isSearching.value = true
  isAnyOperationRunning.value = true
  statusText.value = `正在搜索 "${searchForm.value.term}"...`

  try {
    const res = await Search(searchForm.value.term, searchForm.value.limit, searchForm.value.platform)
    searchResults.value = res.apps || []
    statusText.value = `搜索完成，找到 ${searchResults.value.length} 个应用。`
  } catch (err: any) {
    message.error(`搜索失败: ${err}`)
    statusText.value = '搜索失败'
  } finally {
    isSearching.value = false
    isAnyOperationRunning.value = false
  }
}

function selectAppForVersions(app: main.AppItem) {
  versionForm.value.appName = app.name
  versionForm.value.bundleId = app.bundleID
  versionForm.value.appId = app.id
  versionForm.value.filter = ''
  targetVersionInput.value = ''
  activeTab.value = 'versions'
  handleListVersions()
}

async function downloadFromSearch(app: main.AppItem) {
  try {
    await AddDownloadTask(
      app.name,
      app.bundleID,
      app.id,
      app.version || '最新版',
      '',
      ''
    )
    message.success(`已添加任务「${app.name}」到下载中心`)
    activeTab.value = 'download'
  } catch (err: any) {
    message.error(`添加下载失败: ${err}`)
  }
}

async function downloadFromPurchased(app: main.AppItem) {
  try {
    await AddDownloadTask(
      app.name,
      app.bundleID,
      app.id,
      app.version || '最新版',
      '',
      ''
    )
    message.success(`已添加任务「${app.name}」到下载中心`)
    activeTab.value = 'download'
  } catch (err: any) {
    message.error(`添加下载失败: ${err}`)
  }
}

async function handlePurchaseApp(bundleId: string) {
  isAnyOperationRunning.value = true
  statusText.value = `正在获取应用授权: ${bundleId}...`
  try {
    const res = await Purchase(bundleId)
    if (res.alreadyOwned) {
      message.info('你已经拥有该应用的正版许可，无需重复获取。')
    } else {
      message.success('获取许可成功')
    }
  } catch (err: any) {
    message.error(`获取许可失败: ${err}`)
  } finally {
    isAnyOperationRunning.value = false
  }
}

// Versions
async function handleListVersions() {
  if (!versionForm.value.bundleId) {
    message.warning('请输入 Bundle ID')
    return
  }

  isListingVersions.value = true
  isAnyOperationRunning.value = true
  statusText.value = `正在查询 ${versionForm.value.bundleId} 的历史版本列表...`

  try {
    const res = await ListVersions(versionForm.value.bundleId, versionForm.value.appId)
    versionItems.value = (res.externalVersionIdentifiers || []).map(id => ({
      versionId: id,
      displayVersion: '未查询',
      fileSize: '-',
      releaseDate: '-'
    }))
    statusText.value = `共获取到 ${versionItems.value.length} 个历史版本记录。`
  } catch (err: any) {
    message.error(`查询历史版本失败: ${err}`)
    statusText.value = '查询历史版本失败'
  } finally {
    isListingVersions.value = false
    isAnyOperationRunning.value = false
  }
}

async function cancellableSleep(ms: number, stopRef: { value: boolean }): Promise<boolean> {
  const start = Date.now()
  while (Date.now() - start < ms) {
    if (stopRef.value) return false
    await new Promise(r => setTimeout(r, 20))
  }
  return !stopRef.value
}

async function fetchItemMetadata(index: number): Promise<string> {
  if (index < 0 || index >= versionItems.value.length) return ''
  const item = versionItems.value[index]
  if (item.displayVersion && item.displayVersion !== '未查询') {
    return item.displayVersion
  }
  item.isQuerying = true
  try {
    const res = await GetVersionMetadata(versionForm.value.bundleId, item.versionId, versionForm.value.appId)
    item.displayVersion = res.displayVersion
    item.fileSize = res.displayFileSize
    item.releaseDate = res.releaseDate ? res.releaseDate.replace(/T.*/, '') : '-'
    versionItems.value = [...versionItems.value]
    return item.displayVersion
  } catch (err) {
    console.warn(`查询构建 ID ${item.versionId} 失败:`, err)
    return ''
  } finally {
    item.isQuerying = false
  }
}

async function runBatchQuery(count: number) {
  isBatchQuerying.value = true
  shouldStopBatchQuery.value = false
  isAnyOperationRunning.value = true
  statusText.value = `正在批量查询前 ${count} 个历史版本号...`

  try {
    for (let i = 0; i < count; i++) {
      if (shouldStopBatchQuery.value) {
        statusText.value = '批量查询已手动停止。'
        message.info('批量查询已停止')
        break
      }

      const item = versionItems.value[i]
      if (item.displayVersion !== '未查询') continue

      statusText.value = `正在查询 (${i + 1}/${count}): ${item.versionId}...`
      await fetchItemMetadata(i)

      if (shouldStopBatchQuery.value) break
      if (!await cancellableSleep(120, shouldStopBatchQuery)) break
    }

    if (!shouldStopBatchQuery.value) {
      statusText.value = '批量查询完成。'
      message.success('批量查询完成')
    }
  } catch (err: any) {
    statusText.value = `批量查询失败: ${err}`
    message.error(`批量查询失败: ${err}`)
  } finally {
    isBatchQuerying.value = false
    shouldStopBatchQuery.value = false
    isAnyOperationRunning.value = false
  }
}

function handleBatchQueryClick() {
  if (isBatchQuerying.value) {
    stopBatchQuery()
    return
  }
  handleBatchQuery()
}

function stopBatchQuery() {
  shouldStopBatchQuery.value = true
  statusText.value = '正在停止批量查询...'
}

function handleBatchQuery() {
  const count = Math.min(versionItems.value.length, 30)
  dialog.info({
    title: '批量解析确认',
    content: `即将批量解析前 ${count} 个版本的详细版本号与体积，是否继续？`,
    positiveText: '开始解析',
    negativeText: '取消',
    onPositiveClick: () => {
      void runBatchQuery(count)
    }
  })
}

function handleTargetQueryClick() {
  if (isTargetQuerying.value) {
    stopTargetQuery()
    return
  }
  if (versionItems.value.length === 0) {
    message.warning('请先点击「获取历史版本」')
    return
  }
  showTargetVersionModal.value = true
}

function stopTargetQuery() {
  shouldStopTargetQuery.value = true
  statusText.value = '正在停止版本检索...'
}

function confirmStartTargetQuery() {
  const target = targetVersionInput.value.trim()
  if (!target) {
    message.warning('请输入目标版本号')
    return
  }
  showTargetVersionModal.value = false
  void runTargetQuery(target)
}

async function runTargetQuery(targetVersion: string) {
  isTargetQuerying.value = true
  shouldStopTargetQuery.value = false
  isAnyOperationRunning.value = true
  statusText.value = `正在检索目标版本: ${targetVersion}...`

  const total = versionItems.value.length
  let found = false
  let foundItem: VersionItem | null = null

  try {
    let low = 0
    let high = total - 1

    statusText.value = `正在核对最新版本 (1/${total})...`
    const v0 = await fetchItemMetadata(0)
    if (shouldStopTargetQuery.value) {
      statusText.value = `已停止查询指定版本 (${targetVersion})。`
      return
    }

    if (v0 && isVersionMatch(v0, targetVersion)) {
      found = true
      foundItem = versionItems.value[0]
    } else if (v0 && compareVersions(v0, targetVersion) < 0) {
      statusText.value = `当前最新版本 (${v0}) 小于目标版本 ${targetVersion}，未找到该版本。`
      message.warning(`当前最新版本为 ${v0}，未找到更高版本 ${targetVersion}`)
      return
    } else {
      low = 1
    }

    if (!found && high >= low) {
      statusText.value = `正在核对早期版本 (${total}/${total})...`
      const vLast = await fetchItemMetadata(high)
      if (shouldStopTargetQuery.value) {
        statusText.value = `已停止查询指定版本 (${targetVersion})。`
        return
      }

      if (vLast && isVersionMatch(vLast, targetVersion)) {
        found = true
        foundItem = versionItems.value[high]
      } else if (vLast && compareVersions(vLast, targetVersion) > 0) {
        statusText.value = `当前早期版本 (${vLast}) 大于目标版本 ${targetVersion}，未找到该版本。`
        message.warning(`当前早期版本为 ${vLast}，未找到更低版本 ${targetVersion}`)
        return
      } else {
        high = high - 1
      }
    }

    if (!found) {
      while (low <= high && !shouldStopTargetQuery.value) {
        const mid = Math.floor((low + high) / 2)
        statusText.value = `正在核对版本 (构建 ID: ${versionItems.value[mid].versionId})...`
        const vMid = await fetchItemMetadata(mid)

        if (shouldStopTargetQuery.value) break
        if (!await cancellableSleep(120, shouldStopTargetQuery)) break

        if (vMid && isVersionMatch(vMid, targetVersion)) {
          found = true
          foundItem = versionItems.value[mid]
          break
        }

        if (vMid) {
          if (compareVersions(vMid, targetVersion) < 0) {
            high = mid - 1
          } else {
            low = mid + 1
          }
        } else {
          low++
        }
      }
    }

    if (shouldStopTargetQuery.value) {
      statusText.value = `已停止查询指定版本 (${targetVersion})。`
      message.info('已停止查询指定版本')
    } else if (found && foundItem) {
      statusText.value = `已找到目标版本 ${foundItem.displayVersion} (构建 ID: ${foundItem.versionId})！`
      message.success(`已查询到指定版本 ${foundItem.displayVersion}！`)
      versionForm.value.filter = targetVersion
    } else {
      statusText.value = `检索完成，未在历史记录中找到指定版本: ${targetVersion}`
      message.warning(`未找到版本号: ${targetVersion}`)
    }
  } catch (err: any) {
    statusText.value = `查询指定版本失败: ${err}`
    message.error(`查询失败: ${err}`)
  } finally {
    isTargetQuerying.value = false
    shouldStopTargetQuery.value = false
    isAnyOperationRunning.value = false
  }
}

// Download Tasks Operations
async function loadDownloadTasks() {
  try {
    const list = await GetDownloadTasks()
    downloadTasks.value = list || []
  } catch (err: any) {
    console.error('加载任务失败:', err)
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

function handleOpenDefaultDownloadDir() {
  OpenInExplorer(settings.value.defaultDownloadDir || '')
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

// IPA Installer
function useIPAPath(paths: string[]) {
  const ipaPath = paths.find(path => path.toLowerCase().endsWith('.ipa'))
  if (!ipaPath) {
    message.warning('请拖放 .ipa 格式的安装包')
    return
  }
  selectedIPAPath.value = ipaPath
  activeTab.value = 'installer'
  statusText.value = `已选定 IPA: ${ipaPath}`
}

async function handleSelectIPA() {
  try {
    const path = await SelectIPA()
    if (path) {
      useIPAPath([path])
    }
  } catch (err: any) {
    message.error(`选择 IPA 失败: ${err}`)
  }
}

async function loadConnectedDevices() {
  isLoadingDevices.value = true
  isAnyOperationRunning.value = true
  statusText.value = '正在检测已连接的苹果设备...'
  try {
    const result = await ListDevices()
    devices.value = result || []
    if (!devices.value.some(device => device.udid === selectedDeviceUDID.value)) {
      selectedDeviceUDID.value = devices.value.length === 1 ? devices.value[0].udid : null
    }
    if (devices.value.length === 0) {
      statusText.value = '未发现可用苹果设备'
      message.warning('未发现设备，请确认设备已连接、解锁并信任此电脑')
    } else {
      statusText.value = `发现 ${devices.value.length} 台可用设备`
    }
  } catch (err: any) {
    devices.value = []
    selectedDeviceUDID.value = null
    statusText.value = '设备检测失败'
    message.error(`设备检测失败: ${err}`)
  } finally {
    isLoadingDevices.value = false
    isAnyOperationRunning.value = false
  }
}

async function handleInstallIPA() {
  if (!selectedIPAPath.value) {
    message.warning('请先选择 IPA 文件')
    return
  }
  if (!selectedDeviceUDID.value) {
    message.warning('请选择要安装的苹果设备')
    return
  }

  isInstallingIPA.value = true
  isAnyOperationRunning.value = true
  statusText.value = `正在安装 ${selectedIPAFileName.value}（请保持设备屏幕常亮勿息屏）...`
  message.info('开始安装应用，请确保设备屏幕保持常亮解锁...')
  try {
    const result = await InstallIPA(selectedIPAPath.value, selectedDeviceUDID.value)
    if (result.success) {
      statusText.value = result.message
      message.success(result.message)
    } else {
      statusText.value = 'IPA 安装失败'
      message.error(result.message || 'IPA 安装失败')
    }
  } catch (err: any) {
    statusText.value = 'IPA 安装失败'
    message.error(`IPA 安装失败: ${err}`)
  } finally {
    isInstallingIPA.value = false
    isAnyOperationRunning.value = false
  }
}

// Purchased
async function loadPurchases() {
  isPurchasedLoading.value = true
  isAnyOperationRunning.value = true
  purchasedLoadLoaded.value = 0
  purchasedLoadTotal.value = 0
  statusText.value = '正在获取已购应用总数...'

  try {
    const pageSize = 100
    const firstPage = await ListPurchases(1, pageSize)
    const firstApps = firstPage.apps || []
    const firstTotal = firstPage.totalCount || firstApps.length
    const allApps = [...firstApps]
    const totalPages = Math.max(1, Math.ceil(firstTotal / pageSize))

    purchasedLoadTotal.value = firstTotal
    purchasedLoadLoaded.value = allApps.length
    purchasedTotal.value = firstTotal

    const maxConcurrentPages = 4
    for (let batchStart = 2; batchStart <= totalPages; batchStart += maxConcurrentPages) {
      const pages = Array.from(
        { length: Math.min(maxConcurrentPages, totalPages - batchStart + 1) },
        (_, index) => batchStart + index
      )
      const pageResults = await Promise.all(pages.map(async page => {
        const pageResult = await ListPurchases(page, pageSize)
        const pageApps = pageResult.apps || []
        purchasedLoadLoaded.value = Math.min(purchasedLoadLoaded.value + pageApps.length, firstTotal)
        statusText.value = `正在加载已购应用 (${purchasedLoadLoaded.value}/${firstTotal})...`
        return { page, apps: pageApps }
      }))

      pageResults.sort((a, b) => a.page - b.page)
      for (const result of pageResults) {
        allApps.push(...result.apps)
      }
    }

    purchasedApps.value = allApps
    purchasedTotal.value = firstTotal || allApps.length
    purchasedPage.value = 1
    statusText.value = `已购应用加载完成 (共 ${purchasedApps.value.length} 款)。`
  } catch (err: any) {
    message.error(`加载已购列表失败: ${err}`)
    statusText.value = '加载失败'
  } finally {
    isPurchasedLoading.value = false
    isAnyOperationRunning.value = false
  }
}

function onPurchasedPageSizeChange(val: string | number | null) {
  if (val === null) return
  purchasedPageSize.value = Number(val)
  purchasedPage.value = 1
}

function prevPurchasedPage() {
  if (purchasedPage.value > 1) {
    purchasedPage.value--
  }
}

function nextPurchasedPage() {
  if (purchasedPage.value < purchasedPageCount.value) {
    purchasedPage.value++
  }
}

// Settings
async function handleSelectDefaultDir() {
  try {
    const dir = await SelectDirectory('选择默认下载保存目录', settings.value.defaultDownloadDir)
    if (dir) {
      settings.value.defaultDownloadDir = dir
    }
  } catch (err: any) {
    console.error(err)
  }
}

async function handleTestProxy() {
  isTestingProxy.value = true
  statusText.value = '正在测试网络代理连接...'
  try {
    const res = await TestProxy(settings.value.proxyUrl)
    if (res.success) {
      message.success(res.message)
      statusText.value = '网络代理测试成功！'
    } else {
      message.error(res.message)
      statusText.value = '网络代理测试失败'
    }
  } catch (err: any) {
    message.error(`测试异常: ${err}`)
  } finally {
    isTestingProxy.value = false
  }
}

async function handleSaveSettings() {
  try {
    await SaveSettings(settings.value)
    message.success('设置已保存并生效！')
  } catch (err: any) {
    message.error(`保存失败: ${err}`)
  }
}

function openGitHub() {
  BrowserOpenURL('https://github.com/wnnz/AppleVault')
}

async function copyGitHubUrl() {
  await copy('https://github.com/wnnz/AppleVault')
  message.success('已复制仓库地址到剪贴板！')
}

function handleCancel() {
  shouldStopBatchQuery.value = true
  shouldStopTargetQuery.value = true
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
    account.value = { name: '果仓助手用户', email: 'applevault.user@icloud.com', success: true }
    const demoStatusByTab: Record<string, string> = {
      search: '搜索完成，找到 6 个应用。',
      versions: '共获取 8 个历史版本记录。',
      download: '当前有 1 个任务正在下载，速度 12.8 MB/s',
      purchased: '已购应用加载完成（共 6 款）。',
      installer: '已就绪，已检测到 iPhone 15 Pro Max',
      account: '已登录 果仓助手用户（applevault.user@icloud.com）',
      settings: '底层引擎就绪，配置已加载',
      about: '果仓助手 (AppleVault) v1.1.0'
    }
    statusText.value = demoStatusByTab[tabParam || 'search'] || '就绪'
    searchForm.value.term = '微信'
    settings.value.proxyUrl = 'http://127.0.0.1:7890'
    settings.value.defaultDownloadDir = 'D:\\Dev\\AppleVault\\data\\downloads\\applevault.user'
    searchResults.value = [
      { id: 414478124, bundleID: 'com.tencent.xin', name: '微信 (WeChat)', version: '8.0.50', price: 0, displayPrice: '免费' },
      { id: 333206289, bundleID: 'com.alipay.iphoneclient', name: '支付宝 - 生活好 支付宝', version: '10.5.88', price: 0, displayPrice: '免费' },
      { id: 736536022, bundleID: 'tv.danmaku.bilianime', name: '哔哩哔哩 (Bilibili)', version: '7.82.0', price: 0, displayPrice: '免费' },
      { id: 590338362, bundleID: 'com.netease.cloudmusic', name: '网易云音乐', version: '9.0.70', price: 0, displayPrice: '免费' },
      { id: 835599320, bundleID: 'com.zhiliaoapp.musically', name: 'TikTok - 精彩短视频与音乐', version: '35.8.0', price: 0, displayPrice: '免费' },
      { id: 444934666, bundleID: 'com.tencent.mqq', name: 'QQ - 轻松做自己', version: '9.0.65', price: 0, displayPrice: '免费' }
    ]
    versionForm.value.bundleId = 'com.tencent.xin'
    versionForm.value.appName = '微信'
    versionItems.value = [
      { versionId: '868192301', displayVersion: '8.0.50', fileSize: '286.4 MB', releaseDate: '2024-08-15' },
      { versionId: '867204918', displayVersion: '8.0.49', fileSize: '284.1 MB', releaseDate: '2024-07-20' },
      { versionId: '865819021', displayVersion: '8.0.48', fileSize: '279.8 MB', releaseDate: '2024-06-12' },
      { versionId: '864201990', displayVersion: '8.0.47', fileSize: '275.2 MB', releaseDate: '2024-05-08' },
      { versionId: '862901124', displayVersion: '8.0.46', fileSize: '270.5 MB', releaseDate: '2024-04-01' },
      { versionId: '861502391', displayVersion: '8.0.45', fileSize: '268.0 MB', releaseDate: '2024-03-05' },
      { versionId: '859810234', displayVersion: '8.0.44', fileSize: '263.8 MB', releaseDate: '2024-01-22' },
      { versionId: '858201992', displayVersion: '8.0.43', fileSize: '260.1 MB', releaseDate: '2023-12-18' }
    ]
    downloadTasks.value = [
      { id: '1', appName: '微信 (WeChat)', bundleID: 'com.tencent.xin', appId: 414478124, version: '8.0.50', versionId: '868192301', fileSize: '286.4 MB', totalBytes: 300312000, currBytes: 195202800, progress: 65, speed: '12.8 MB/s', status: 'downloading', outputPath: '', errorMessage: '', createdAt: '2024-08-20 15:30:12' },
      { id: '2', appName: '支付宝', bundleID: 'com.alipay.iphoneclient', appId: 333206289, version: '10.5.88', versionId: '865001129', fileSize: '142.0 MB', totalBytes: 148897000, currBytes: 148897000, progress: 100, speed: '已完成', status: 'completed', outputPath: 'data/downloads/user@icloud.com/alipay.ipa', errorMessage: '', createdAt: '2024-08-20 15:24:05' },
      { id: '3', appName: 'Infuse · 精彩影音播放器', bundleID: 'com.firecore.infuse', appId: 1136220934, version: '7.7.2', versionId: '863920191', fileSize: '118.5 MB', totalBytes: 124256256, currBytes: 124256256, progress: 100, speed: '已完成', status: 'completed', outputPath: 'data/downloads/user@icloud.com/infuse.ipa', errorMessage: '', createdAt: '2024-08-20 14:10:33' },
      { id: '4', appName: '网易云音乐', bundleID: 'com.netease.cloudmusic', appId: 590338362, version: '9.0.70', versionId: '864201991', fileSize: '215.3 MB', totalBytes: 225758413, currBytes: 0, progress: 0, speed: '等待中', status: 'pending', outputPath: '', errorMessage: '', createdAt: '2024-08-20 15:31:00' }
    ]
    purchasedApps.value = [
      { id: 414478124, bundleID: 'com.tencent.xin', name: '微信 (WeChat)', version: '8.0.50', price: 0, purchaseDate: '2024-01-15 10:20', displayPrice: '已购' },
      { id: 333206289, bundleID: 'com.alipay.iphoneclient', name: '支付宝', version: '10.5.88', price: 0, purchaseDate: '2024-02-08 14:12', displayPrice: '已购' },
      { id: 1136220934, bundleID: 'com.firecore.infuse', name: 'Infuse • 精彩影音播放器', version: '7.7.2', price: 0, purchaseDate: '2024-03-22 09:45', displayPrice: '已购' },
      { id: 916364737, bundleID: 'com.procreate.pocket', name: 'Procreate Pocket', version: '4.0.11', price: 0, purchaseDate: '2024-04-10 18:30', displayPrice: '已购' },
      { id: 1596487405, bundleID: 'com.taguirov.adam.WebDAV', name: 'WebDAV Manager', version: '2.1.0', price: 0, purchaseDate: '2024-05-19 16:30', displayPrice: '已购' },
      { id: 736536022, bundleID: 'tv.danmaku.bilianime', name: '哔哩哔哩 (Bilibili)', version: '7.82.0', price: 0, purchaseDate: '2024-06-01 21:05', displayPrice: '已购' }
    ]
    purchasedTotal.value = 6
    devices.value = [
      { udid: '00008130-001A49021E28001C', name: 'iPhone 15 Pro Max', productType: 'iPhone 15 Pro Max', productVersion: '17.5.1', connectionType: 'USB 3.0' }
    ]
    selectedDeviceUDID.value = '00008130-001A49021E28001C'
    selectedIPAPath.value = 'D:\\Dev\\AppleVault\\data\\downloads\\applevault.user@icloud.com\\WeChat_8.0.50.ipa'
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
    const idx = downloadTasks.value.findIndex(t => t.id === updatedTask.id)
    if (idx !== -1) {
      downloadTasks.value[idx] = updatedTask
    } else {
      downloadTasks.value.unshift(updatedTask)
    }
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
