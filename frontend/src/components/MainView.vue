<template>
  <div class="main-container" :style="{ background: isDark ? '#18181c' : '#f5f7fa', color: isDark ? '#e0e0e0' : '#1f2225' }">
    <!-- 1. Top Header -->
    <header class="header-bar" :style="{ borderColor: isDark ? '#2d2d30' : '#e5e7eb', background: isDark ? '#1f1f23' : '#ffffff' }">
      <div class="header-left">
        <span class="logo-emoji">🍎</span>
        <span class="logo-text">IPATool GUI</span>
        <n-tag size="small" type="info" :bordered="false" round>Wails v2 + Go 1.26</n-tag>
        <span class="subtitle">App Store 官方正版 / 历史旧版 IPA 下载</span>
      </div>

      <div class="header-right">
        <!-- Proxy Quick Switch -->
        <div class="header-badge" :style="{ borderColor: isDark ? '#333' : '#e5e7eb', background: isDark ? '#28282c' : '#f9fafb' }">
          <n-switch v-model:value="settings.enableProxy" size="small" @update:value="onProxyToggle">
            <template #checked>代理开</template>
            <template #unchecked>代理关</template>
          </n-switch>
          <span class="badge-text" :title="settings.proxyUrl">{{ settings.proxyUrl }}</span>
        </div>

        <!-- Account Badge -->
        <div class="header-badge" :style="{ borderColor: isDark ? '#333' : '#e5e7eb', background: isDark ? '#28282c' : '#f9fafb' }">
          <span class="status-dot" :class="isLoggedIn ? 'dot-online' : 'dot-offline'"></span>
          <span class="badge-text font-bold">{{ account.name || '未登录' }}</span>
          <span v-if="account.email" class="badge-subtext">({{ account.email }})</span>
        </div>

        <n-button size="tiny" secondary @click="refreshAccount" :loading="isAccountLoading">
          刷新
        </n-button>

        <!-- Dark Mode Toggle -->
        <n-button size="small" circle quaternary @click="emit('toggleTheme')">
          <template #icon>
            <span>{{ isDark ? '☀️' : '🌙' }}</span>
          </template>
        </n-button>
      </div>
    </header>

    <!-- 2. Content Tabs -->
    <main class="content-body">
      <n-tabs v-model:value="activeTab" type="line" animated>
        <!-- TAB 1: 账号管理 -->
        <n-tab-pane name="account" tab="👤 账号管理">
          <div class="tab-pane-content">
            <n-grid :x-gap="16" :y-gap="16" cols="1 m:2" responsive="screen">
              <!-- Left: Current Account Info -->
              <n-grid-item>
                <n-card title="当前 Apple ID 状态" size="small" hoverable>
                  <n-space vertical size="large">
                    <div>
                      <div class="info-label">账号名称</div>
                      <div class="info-value font-bold">{{ account.name || '未登录' }}</div>
                    </div>
                    <div>
                      <div class="info-label">Apple ID 邮箱</div>
                      <div class="info-value">{{ account.email || '未登录' }}</div>
                    </div>
                    <n-alert type="info" :show-icon="true" size="small">
                      使用个人 Apple ID 官方凭据直接从苹果 App Store 官方服务器下载正版包，下载的文件包含您的正版签名，导入设备绝不闪退。
                    </n-alert>
                    <n-space>
                      <n-button type="error" secondary :disabled="!isLoggedIn" @click="handleRevoke" :loading="isRevoking">
                        退出登录 (Revoke)
                      </n-button>
                      <n-button secondary @click="handleClearKeychain" :loading="isClearing">
                        清空本地密钥缓存
                      </n-button>
                    </n-space>
                  </n-space>
                </n-card>
              </n-grid-item>

              <!-- Right: Login Form -->
              <n-grid-item>
                <n-card title="登录 Apple ID" size="small" hoverable>
                  <n-form label-placement="top" size="medium">
                    <n-form-item label="Apple ID 邮箱:">
                      <n-input v-model:value="loginForm.email" placeholder="例如: your_apple_id@icloud.com" />
                    </n-form-item>
                    <n-form-item label="Apple ID 密码:">
                      <n-input v-model:value="loginForm.password" type="password" show-password-on="click" placeholder="请输入密码" @keydown.enter="handleLogin" />
                    </n-form-item>
                    <n-alert type="warning" :show-icon="true" size="small" class="mb-4">
                      若账号开启了双重认证（2FA），点击登录后将自动弹出验证码输入窗口，无需在主界面提前输入。
                    </n-alert>
                    <n-button type="primary" block size="large" :loading="isLoggingIn" @click="handleLogin">
                      登 录 Apple ID
                    </n-button>
                  </n-form>
                </n-card>
              </n-grid-item>
            </n-grid>
          </div>
        </n-tab-pane>

        <!-- TAB 2: 应用搜索 -->
        <n-tab-pane name="search" tab="🔍 应用搜索">
          <div class="tab-pane-content">
            <n-card size="small" class="mb-3">
              <n-space align="center">
                <n-input v-model:value="searchForm.term" placeholder="输入应用名称（如: 支付宝、微信、TikTok）" style="width: 320px;" @keydown.enter="handleSearch" />
                <n-select v-model:value="searchForm.platform" :options="platformOptions" style="width: 130px;" />
                <n-select v-model:value="searchForm.limit" :options="limitOptions" style="width: 100px;" />
                <n-button type="primary" :loading="isSearching" @click="handleSearch">
                  搜 索
                </n-button>
              </n-space>
            </n-card>

            <n-card size="small" :bordered="false" content-style="padding: 0;">
              <n-data-table
                :columns="searchColumns"
                :data="searchResults"
                :loading="isSearching"
                :pagination="{ pageSize: 10 }"
                size="small"
                :max-height="400"
              />
            </n-card>
          </div>
        </n-tab-pane>

        <!-- TAB 3: 历史版本 -->
        <n-tab-pane name="versions" tab="📜 历史版本">
          <div class="tab-pane-content">
            <n-card size="small" class="mb-3">
              <n-space vertical size="medium">
                <n-space align="center">
                  <span class="form-label">Bundle ID:</span>
                  <n-input v-model:value="versionForm.bundleId" placeholder="例如: com.alipay.iphoneclient" style="width: 320px;" @keydown.enter="handleListVersions" />
                  <n-button type="primary" :loading="isListingVersions" @click="handleListVersions">
                    获取历史版本列表
                  </n-button>
                  <n-button secondary :disabled="versionItems.length === 0" :loading="isBatchQuerying" @click="handleBatchQuery">
                    批量查询前 30 个版本号
                  </n-button>
                </n-space>

                <n-space align="center">
                  <span class="form-label">筛选版本:</span>
                  <n-input v-model:value="versionForm.filter" placeholder="输入版本号(如 10.2.96)、体积或构建 ID 实时过滤" style="width: 320px;" />
                  <n-tag type="info" size="small" round>共 {{ filteredVersions.length }} / {{ versionItems.length }} 个版本</n-tag>
                </n-space>
              </n-space>
            </n-card>

            <n-card size="small" :bordered="false" content-style="padding: 0;">
              <n-data-table
                :columns="versionColumns"
                :data="filteredVersions"
                :loading="isListingVersions"
                :virtual-scroll="true"
                :max-height="420"
                size="small"
              />
            </n-card>
          </div>
        </n-tab-pane>

        <!-- TAB 4: 下载中心 -->
        <n-tab-pane name="download" tab="⬇️ 下载中心">
          <div class="tab-pane-content">
            <n-card size="small" style="max-width: 720px;" hoverable>
              <n-form label-placement="top" size="medium">
                <n-form-item label="Bundle Identifier (应用包名):">
                  <n-input v-model:value="downloadForm.bundleId" placeholder="例如: com.alipay.iphoneclient" />
                </n-form-item>

                <n-form-item label="App ID (可选):">
                  <n-input v-model:value="downloadForm.appId" placeholder="若已填 Bundle ID 可留空" />
                </n-form-item>

                <n-form-item label="历史版本构建 ID (External Version ID，留空下载最新版):">
                  <n-input v-model:value="downloadForm.versionId" placeholder="例如: 851864107 (支付宝 10.2.96)" />
                </n-form-item>

                <n-form-item label="目标平台:">
                  <n-select v-model:value="downloadForm.platform" :options="platformOptions" />
                </n-form-item>

                <n-form-item label="保存目录:">
                  <n-input-group>
                    <n-input v-model:value="downloadForm.outputPath" placeholder="选择保存目录" />
                    <n-button secondary @click="handleSelectDir">浏览...</n-button>
                  </n-input-group>
                </n-form-item>

                <n-form-item>
                  <n-checkbox v-model:checked="downloadForm.purchase">
                    若当前账号未购买该应用，自动获取免费购买凭证 (--purchase)
                  </n-checkbox>
                </n-form-item>

                <n-space size="large">
                  <n-button type="primary" size="large" :loading="isDownloading" @click="handleStartDownload">
                    🚀 开始下载 IPA 包
                  </n-button>
                  <n-button size="large" secondary @click="handleOpenOutputDir">
                    📁 打开保存目录
                  </n-button>
                </n-space>

                <div v-if="lastDownloadedPath" class="mt-4">
                  <n-alert type="success" title="最近下载成功：" size="small">
                    {{ lastDownloadedPath }}
                  </n-alert>
                </div>
              </n-form>
            </n-card>
          </div>
        </n-tab-pane>

        <!-- TAB 5: 已购应用 -->
        <n-tab-pane name="purchased" tab="📦 已购应用">
          <div class="tab-pane-content">
            <n-card size="small" class="mb-3">
              <n-space justify="space-between" align="center">
                <n-button type="primary" :loading="isPurchasedLoading" @click="loadPurchases">
                  刷新已购应用列表
                </n-button>
                <n-pagination
                  v-model:page="purchasedPage"
                  :page-size="20"
                  :item-count="purchasedTotal"
                  @update:page="loadPurchases"
                />
              </n-space>
            </n-card>

            <n-card size="small" :bordered="false" content-style="padding: 0;">
              <n-data-table
                :columns="purchasedColumns"
                :data="purchasedApps"
                :loading="isPurchasedLoading"
                size="small"
                :max-height="420"
              />
            </n-card>
          </div>
        </n-tab-pane>

        <!-- TAB 6: 设置 -->
        <n-tab-pane name="settings" tab="⚙️ 全局设置">
          <div class="tab-pane-content">
            <n-card size="small" style="max-width: 760px;" hoverable>
              <n-form label-placement="top" size="medium">
                <n-form-item label="内置 ipatool 引擎状态:">
                  <n-alert type="success" size="small" :show-icon="true">
                    <div>已自动加载程序同目录下的核心引擎</div>
                    <div class="text-xs mt-1 text-gray-500">{{ settings.ipaToolPath }}</div>
                  </n-alert>
                </n-form-item>

                <n-form-item label="网络代理配置 (HTTP / SOCKS5):">
                  <n-space vertical style="width: 100%;">
                    <n-checkbox v-model:checked="settings.enableProxy">
                      启用网络代理 (通过环境变量传递给 ipatool 引擎)
                    </n-checkbox>
                    <n-input-group>
                      <n-input v-model:value="settings.proxyUrl" :disabled="!settings.enableProxy" placeholder="例如: http://127.0.0.1:10808" />
                      <n-button secondary :disabled="!settings.enableProxy" :loading="isTestingProxy" @click="handleTestProxy">
                        测试代理连接
                      </n-button>
                    </n-input-group>
                    <div class="text-xs text-gray-400">
                      * 国内访问 Apple App Store 认证服务通常需要开启代理，支持 Clash / v2rayN 等常见本地代理端口。
                    </div>
                  </n-space>
                </n-form-item>

                <n-form-item label="本地密钥库解锁密码 (--keychain-passphrase):">
                  <n-input v-model:value="settings.keychainPassphrase" placeholder="设置本地保护密码（如 123456）" />
                  <div class="text-xs text-green-600 mt-1">
                    ✅ 核心保护：程序会自动以非交互模式注入该参数，彻底解决 Windows 终端卡死与密码校验失败问题。
                  </div>
                </n-form-item>

                <n-form-item label="默认 IPA 保存目录:">
                  <n-input-group>
                    <n-input v-model:value="settings.defaultDownloadDir" placeholder="默认保存目录" />
                    <n-button secondary @click="handleSelectDefaultDir">浏览...</n-button>
                  </n-input-group>
                </n-form-item>

                <n-form-item label="默认平台:">
                  <n-select v-model:value="settings.defaultPlatform" :options="platformOptions" />
                </n-form-item>

                <n-button type="primary" size="large" @click="handleSaveSettings">
                  💾 保存并应用设置
                </n-button>
              </n-form>
            </n-card>
          </div>
        </n-tab-pane>
      </n-tabs>
    </main>

    <!-- 3. Bottom Console Logs -->
    <footer class="console-footer" :style="{ borderColor: isDark ? '#2d2d30' : '#e5e7eb', background: isDark ? '#141416' : '#1e1e20' }">
      <div class="console-header" :style="{ borderColor: isDark ? '#222' : '#333' }">
        <div class="console-title">
          <n-spin v-if="isAnyOperationRunning" size="small" class="mr-2" />
          <span>控制台输出日志</span>
          <span class="text-xs text-gray-400 ml-2">({{ statusText }})</span>
        </div>
        <div class="console-actions">
          <n-button v-if="isAnyOperationRunning" size="tiny" type="error" secondary @click="handleCancel">
            取消操作
          </n-button>
          <n-button size="tiny" quaternary style="color: #ccc;" @click="clearLogs">
            清空日志
          </n-button>
        </div>
      </div>

      <div ref="logContainerRef" class="console-content">
        <div v-for="(log, idx) in logLines" :key="idx" class="log-line" :class="{ 'log-err': log.includes('[ERR]') }">
          {{ log }}
        </div>
      </div>
    </footer>

    <!-- 4. 2FA Modal Dialog -->
    <n-modal v-model:show="show2FAModal" preset="card" title="🔐 Apple ID 双重认证" style="width: 420px;" :mask-closable="false">
      <div class="modal-body">
        <p class="mb-4 text-sm text-gray-600">已向您的受信任 Apple 设备发送了验证码。请输入收到的 6 位验证码以完成登录：</p>
        <n-input
          ref="twoFAInputRef"
          v-model:value="twoFACode"
          placeholder="6 位验证码"
          maxlength="6"
          size="large"
          class="twofa-input"
          @keydown.enter="confirm2FA"
        />
        <div class="flex justify-end gap-3 mt-6">
          <n-button secondary @click="cancel2FA">取消</n-button>
          <n-button type="primary" :disabled="twoFACode.length !== 6" :loading="isLoggingIn" @click="confirm2FA">
            提交验证
          </n-button>
        </div>
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, h } from 'vue'
import { useMessage, useDialog, NButton, NTag, NSpace } from 'naive-ui'
import {
  GetAccountInfo,
  Login,
  Revoke,
  ClearKeychainCache,
  Search,
  ListVersions,
  GetVersionMetadata,
  Download,
  Purchase,
  ListPurchases,
  GetSettings,
  SaveSettings,
  SelectDirectory,
  TestProxy,
  OpenInExplorer,
  CancelRunningCommand
} from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { main } from '../../wailsjs/go/models'

const props = defineProps<{ isDark: boolean }>()
const emit = defineEmits<{ (e: 'toggleTheme'): void }>()

const message = useMessage()
const dialog = useDialog()

const activeTab = ref('versions')
const isAnyOperationRunning = ref(false)
const statusText = ref('就绪')

// Settings
const settings = ref<main.Settings>({
  keychainPassphrase: '123456',
  defaultDownloadDir: 'D:\\Downloads',
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

// Account State
const account = ref<main.AccountInfo>({ name: '', email: '', success: false })
const isLoggedIn = computed(() => account.value.success && !!account.value.email)
const isAccountLoading = ref(false)
const isRevoking = ref(false)
const isClearing = ref(false)
const isLoggingIn = ref(false)

const loginForm = ref({
  email: 'wzheng1996@live.com',
  password: ''
})

// 2FA Modal
const show2FAModal = ref(false)
const twoFACode = ref('')
const twoFAInputRef = ref()

// Search State
const isSearching = ref(false)
const searchForm = ref({
  term: '支付宝',
  limit: 10,
  platform: 'iphone'
})
const searchResults = ref<main.AppItem[]>([])

const searchColumns = [
  { title: '应用名称', key: 'name', width: 220, ellipsis: true },
  { title: 'Bundle ID', key: 'bundleID', width: 220, ellipsis: true },
  { title: 'App ID', key: 'id', width: 120 },
  { title: '最新版本', key: 'version', width: 100 },
  { title: '价格', key: 'displayPrice', width: 80 },
  {
    title: '操作',
    key: 'actions',
    width: 240,
    render(row: main.AppItem) {
      return h(NSpace, { size: 'small' }, () => [
        h(NButton, { size: 'tiny', type: 'primary', secondary: true, onClick: () => selectAppForVersions(row) }, () => '历史版本'),
        h(NButton, { size: 'tiny', secondary: true, onClick: () => selectAppForDownload(row) }, () => '直接下载'),
        h(NButton, { size: 'tiny', tertiary: true, onClick: () => handlePurchaseApp(row.bundleID) }, () => '获取许可')
      ])
    }
  }
]

// Versions State
const isListingVersions = ref(false)
const isBatchQuerying = ref(false)
const versionForm = ref({
  bundleId: 'com.alipay.iphoneclient',
  appId: 0,
  filter: ''
})

interface VersionItem {
  versionId: string
  displayVersion: string
  fileSize: string
  releaseDate: string
  isQuerying?: boolean
}

const versionItems = ref<VersionItem[]>([])

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
  { title: '构建 ID (External Version ID)', key: 'versionId', width: 240 },
  {
    title: '对应版本号',
    key: 'displayVersion',
    width: 140,
    render(row: VersionItem) {
      if (row.displayVersion === '未查询') {
        return h('span', { style: 'color: #999; font-style: italic;' }, '未查询')
      }
      return h('span', { style: 'font-weight: bold; color: #10b981;' }, row.displayVersion)
    }
  },
  { title: '文件体积 (大小)', key: 'fileSize', width: 140 },
  { title: '发布日期', key: 'releaseDate', width: 180 },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    render(row: VersionItem) {
      return h(NSpace, { size: 'small' }, () => [
        h(NButton, {
          size: 'tiny',
          secondary: true,
          loading: row.isQuerying,
          onClick: () => querySingleVersionMetadata(row)
        }, () => '查询详情/体积'),
        h(NButton, {
          size: 'tiny',
          type: 'primary',
          onClick: () => pickVersionForDownload(row.versionId)
        }, () => '一键下载此版')
      ])
    }
  }
]

// Download State
const isDownloading = ref(false)
const lastDownloadedPath = ref('')
const downloadForm = ref({
  bundleId: 'com.alipay.iphoneclient',
  appId: '',
  versionId: '851864107',
  platform: 'iphone',
  outputPath: 'D:\\Downloads',
  purchase: true
})

// Purchased State
const isPurchasedLoading = ref(false)
const purchasedPage = ref(1)
const purchasedTotal = ref(0)
const purchasedApps = ref<main.AppItem[]>([])

const purchasedColumns = [
  { title: '应用名称', key: 'name', width: 220, ellipsis: true },
  { title: 'Bundle ID', key: 'bundleID', width: 220, ellipsis: true },
  { title: 'App ID', key: 'id', width: 120 },
  { title: '获取时间', key: 'purchaseDate', width: 180 },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render(row: main.AppItem) {
      return h(NSpace, { size: 'small' }, () => [
        h(NButton, { size: 'tiny', type: 'primary', secondary: true, onClick: () => selectAppForVersions(row) }, () => '历史版本'),
        h(NButton, { size: 'tiny', secondary: true, onClick: () => selectAppForDownload(row) }, () => '下载')
      ])
    }
  }
]

// Settings State
const isTestingProxy = ref(false)

// Logs State
const logLines = ref<string[]>([])
const logContainerRef = ref<HTMLElement | null>(null)

function appendLog(line: string) {
  logLines.value.push(line)
  if (logLines.value.length > 500) {
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
    downloadForm.value.outputPath = s.defaultDownloadDir || 'D:\\Downloads'
    downloadForm.value.platform = s.defaultPlatform || 'iphone'
  } catch (err: any) {
    console.error(err)
  }
}

async function onProxyToggle(val: boolean) {
  settings.value.enableProxy = val
  await SaveSettings(settings.value)
  message.info(val ? '已开启代理' : '已关闭代理')
}

async function refreshAccount() {
  isAccountLoading.value = true
  statusText.value = '获取账号信息...'
  try {
    const res = await GetAccountInfo()
    account.value = res
    statusText.value = res.success ? `已登录: ${res.name}` : '未登录'
  } catch (err: any) {
    account.value = { name: '', email: '', success: false }
    statusText.value = '未登录'
  } finally {
    isAccountLoading.value = false
  }
}

async function handleLogin() {
  if (!loginForm.value.email || !loginForm.value.password) {
    message.warning('请输入 Apple ID 邮箱与密码')
    return
  }

  isLoggingIn.value = true
  isAnyOperationRunning.value = true
  statusText.value = '正在验证账号与密码...'

  try {
    const res = await Login(loginForm.value.email, loginForm.value.password, '')
    if (res.requires2FA) {
      twoFACode.value = ''
      show2FAModal.value = true
      statusText.value = '等待输入双重认证验证码...'
      nextTick(() => {
        twoFAInputRef.value?.focus()
      })
      return
    }

    if (res.success) {
      account.value = res.account
      message.success(`登录成功: ${res.account.name}`)
      loginForm.value.password = ''
      statusText.value = `登录成功: ${res.account.name}`
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
    message.warning('请输入 6 位数字验证码')
    return
  }

  isLoggingIn.value = true
  isAnyOperationRunning.value = true
  statusText.value = '正在提交 2FA 验证码...'

  try {
    const res = await Login(loginForm.value.email, loginForm.value.password, twoFACode.value)
    if (res.success) {
      show2FAModal.value = false
      account.value = res.account
      message.success(`登录成功: ${res.account.name}`)
      loginForm.value.password = ''
      twoFACode.value = ''
      statusText.value = `登录成功: ${res.account.name}`
    } else {
      message.error(`2FA 验证失败: ${res.errorMessage}`)
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
  statusText.value = '已取消 2FA 验证'
}

async function handleRevoke() {
  dialog.warning({
    title: '确认注销',
    content: '确定要注销当前 Apple ID 凭据并退出登录吗？',
    positiveText: '确定退出',
    negativeText: '取消',
    onPositiveClick: async () => {
      isRevoking.value = true
      isAnyOperationRunning.value = true
      try {
        const ok = await Revoke()
        if (ok) {
          account.value = { name: '', email: '', success: false }
          message.success('已注销登录凭据')
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
    title: '清空本地密钥缓存',
    content: '将删除 ~/.ipatool 目录，解决本地密钥库损坏或校验密码错误的问题。确定清空吗？',
    positiveText: '确定清空',
    negativeText: '取消',
    onPositiveClick: async () => {
      isClearing.value = true
      try {
        await ClearKeychainCache()
        message.success('本地密钥缓存已彻底清空')
        await refreshAccount()
      } catch (err: any) {
        message.error(`清空失败: ${err}`)
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
  statusText.value = `正在搜索: ${searchForm.value.term}...`

  try {
    const res = await Search(searchForm.value.term, searchForm.value.limit, searchForm.value.platform)
    searchResults.value = res.apps || []
    statusText.value = `搜索完成，找到 ${searchResults.value.length} 个应用`
  } catch (err: any) {
    message.error(`搜索失败: ${err}`)
    statusText.value = '搜索失败'
  } finally {
    isSearching.value = false
    isAnyOperationRunning.value = false
  }
}

function selectAppForVersions(app: main.AppItem) {
  versionForm.value.bundleId = app.bundleID
  versionForm.value.appId = app.id
  activeTab.value = 'versions'
  handleListVersions()
}

function selectAppForDownload(app: main.AppItem) {
  downloadForm.value.bundleId = app.bundleID
  downloadForm.value.appId = String(app.id)
  downloadForm.value.versionId = ''
  activeTab.value = 'download'
}

async function handlePurchaseApp(bundleId: string) {
  isAnyOperationRunning.value = true
  statusText.value = `正在获取应用授权: ${bundleId}...`
  try {
    const res = await Purchase(bundleId)
    if (res.alreadyOwned) {
      message.info('您已拥有该应用的授权凭据，无需重复获取。')
    } else {
      message.success('获取授权成功！')
    }
  } catch (err: any) {
    message.error(`获取授权失败: ${err}`)
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
  statusText.value = `正在获取 ${versionForm.value.bundleId} 历史版本...`

  try {
    const res = await ListVersions(versionForm.value.bundleId, versionForm.value.appId)
    versionItems.value = (res.externalVersionIdentifiers || []).map(id => ({
      versionId: id,
      displayVersion: '未查询',
      fileSize: '-',
      releaseDate: '-'
    }))
    statusText.value = `获取成功，共 ${versionItems.value.length} 个历史版本构建 ID`
  } catch (err: any) {
    message.error(`获取版本列表失败: ${err}`)
    statusText.value = '获取版本列表失败'
  } finally {
    isListingVersions.value = false
    isAnyOperationRunning.value = false
  }
}

async function querySingleVersionMetadata(row: VersionItem) {
  row.isQuerying = true
  isAnyOperationRunning.value = true
  statusText.value = `正在查询版本 ID ${row.versionId} 详情...`

  try {
    const res = await GetVersionMetadata(versionForm.value.bundleId, row.versionId, versionForm.value.appId)
    row.displayVersion = res.displayVersion
    row.fileSize = res.displayFileSize
    row.releaseDate = res.releaseDate ? new Date(res.releaseDate).toLocaleDateString() : '-'
    statusText.value = `版本 ${row.versionId} -> ${res.displayVersion} (${res.displayFileSize})`
  } catch (err: any) {
    message.error(`查询详情失败: ${err}`)
  } finally {
    row.isQuerying = false
    isAnyOperationRunning.value = false
  }
}

async function handleBatchQuery() {
  const count = Math.min(versionItems.value.length, 30)
  dialog.info({
    title: '批量查询确认',
    content: `即将批量查询前 ${count} 个历史版本的具体版本号与体积大小，耗时约 15~30 秒，是否继续？`,
    positiveText: '开始查询',
    negativeText: '取消',
    onPositiveClick: async () => {
      isBatchQuerying.value = true
      isAnyOperationRunning.value = true

      for (let i = 0; i < count; i++) {
        const item = versionItems.value[i]
        if (item.displayVersion !== '未查询') continue

        statusText.value = `批量查询进度 (${i + 1}/${count}): ${item.versionId}...`
        try {
          const res = await GetVersionMetadata(versionForm.value.bundleId, item.versionId, versionForm.value.appId)
          item.displayVersion = res.displayVersion
          item.fileSize = res.displayFileSize
          item.releaseDate = res.releaseDate ? new Date(res.releaseDate).toLocaleDateString() : '-'
        } catch {
          // Continue
        }
        await new Promise(r => setTimeout(r, 150))
      }

      isBatchQuerying.value = false
      isAnyOperationRunning.value = false
      statusText.value = '批量查询完成'
      message.success('前 30 个版本详情批量查询完毕')
    }
  })
}

function pickVersionForDownload(verId: string) {
  downloadForm.value.bundleId = versionForm.value.bundleId
  downloadForm.value.versionId = verId
  activeTab.value = 'download'
}

// Download
async function handleSelectDir() {
  try {
    const dir = await SelectDirectory('选择 IPA 文件保存目录', downloadForm.value.outputPath)
    if (dir) {
      downloadForm.value.outputPath = dir
    }
  } catch (err: any) {
    console.error(err)
  }
}

async function handleStartDownload() {
  if (!downloadForm.value.bundleId) {
    message.warning('请输入 Bundle ID')
    return
  }

  isDownloading.value = true
  isAnyOperationRunning.value = true
  const verDesc = downloadForm.value.versionId ? `版本 ID: ${downloadForm.value.versionId}` : '最新版'
  statusText.value = `正在下载 ${downloadForm.value.bundleId} (${verDesc})...`

  try {
    const res = await Download(
      downloadForm.value.bundleId,
      Number(downloadForm.value.appId) || 0,
      downloadForm.value.versionId,
      downloadForm.value.outputPath,
      downloadForm.value.platform,
      downloadForm.value.purchase
    )
    if (res.success) {
      lastDownloadedPath.value = res.output
      statusText.value = `下载成功: ${res.output}`
      dialog.success({
        title: '🎉 下载成功',
        content: `文件已保存至:\n${res.output}\n\n是否立即打开所在文件夹？`,
        positiveText: '打开所在目录',
        negativeText: '关闭',
        onPositiveClick: () => {
          OpenInExplorer(res.output)
        }
      })
    } else {
      message.error('下载未完成')
    }
  } catch (err: any) {
    message.error(`下载失败: ${err}`)
    statusText.value = '下载失败'
  } finally {
    isDownloading.value = false
    isAnyOperationRunning.value = false
  }
}

function handleOpenOutputDir() {
  if (lastDownloadedPath.value) {
    OpenInExplorer(lastDownloadedPath.value)
  } else {
    OpenInExplorer(downloadForm.value.outputPath)
  }
}

// Purchased
async function loadPurchases() {
  isPurchasedLoading.value = true
  isAnyOperationRunning.value = true
  statusText.value = `加载第 ${purchasedPage.value} 页已购应用...`

  try {
    const res = await ListPurchases(purchasedPage.value, 20)
    purchasedApps.value = res.apps || []
    purchasedTotal.value = res.totalCount || 0
    statusText.value = `已购应用加载完成 (共 ${res.totalCount} 个)`
  } catch (err: any) {
    message.error(`加载已购失败: ${err}`)
    statusText.value = '加载失败'
  } finally {
    isPurchasedLoading.value = false
    isAnyOperationRunning.value = false
  }
}

// Settings
async function handleSelectDefaultDir() {
  try {
    const dir = await SelectDirectory('选择默认下载目录', settings.value.defaultDownloadDir)
    if (dir) {
      settings.value.defaultDownloadDir = dir
      downloadForm.value.outputPath = dir
    }
  } catch (err: any) {
    console.error(err)
  }
}

async function handleTestProxy() {
  isTestingProxy.value = true
  try {
    const res = await TestProxy(settings.value.proxyUrl)
    if (res.success) {
      message.success(res.message)
    } else {
      message.error(res.message)
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
    downloadForm.value.outputPath = settings.value.defaultDownloadDir
    downloadForm.value.platform = settings.value.defaultPlatform
    message.success('全局设置保存成功')
  } catch (err: any) {
    message.error(`保存失败: ${err}`)
  }
}

function handleCancel() {
  CancelRunningCommand()
  isAnyOperationRunning.value = false
}

// Lifecycle
onMounted(() => {
  loadSettings()
  refreshAccount()

  EventsOn('log', (msg: string) => {
    appendLog(msg)
  })
})
</script>

<style scoped>
.main-container {
  display: flex;
  flex-direction: column;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  user-select: none;
}

.header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 18px;
  border-bottom: 1px solid;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.logo-emoji {
  font-size: 20px;
}

.logo-text {
  font-size: 16px;
  font-weight: bold;
}

.subtitle {
  font-size: 12px;
  color: #888;
  margin-left: 6px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px;
  border-radius: 14px;
  border: 1px solid;
  font-size: 12px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.dot-online {
  background-color: #10b981;
}

.dot-offline {
  background-color: #ef4444;
}

.badge-text {
  font-size: 12px;
}

.badge-subtext {
  font-size: 11px;
  color: #888;
}

.content-body {
  flex: 1;
  padding: 12px 18px;
  overflow-y: auto;
}

.tab-pane-content {
  padding-top: 6px;
}

.info-label {
  font-size: 12px;
  color: #888;
  margin-bottom: 2px;
}

.info-value {
  font-size: 14px;
}

.form-label {
  font-size: 13px;
  font-weight: 500;
}

.console-footer {
  height: 180px;
  display: flex;
  flex-direction: column;
  border-top: 1px solid;
  flex-shrink: 0;
}

.console-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 14px;
  border-bottom: 1px solid;
  background: rgba(0, 0, 0, 0.15);
}

.console-title {
  font-size: 12px;
  font-weight: 600;
  color: #aaa;
  display: flex;
  align-items: center;
}

.console-content {
  flex: 1;
  padding: 8px 14px;
  overflow-y: auto;
  font-family: Consolas, "Courier New", monospace;
  font-size: 11.5px;
  color: #ccc;
  line-height: 1.5;
  background: #18181b;
}

.log-line {
  word-break: break-all;
  white-space: pre-wrap;
}

.log-err {
  color: #f87171;
}

.twofa-input {
  text-align: center;
  font-size: 24px;
  font-weight: bold;
  letter-spacing: 6px;
}
</style>
