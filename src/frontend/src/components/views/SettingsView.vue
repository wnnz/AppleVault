<template>
  <div class="view-panel">
    <div class="view-header">
      <div>
        <div class="title-with-badge"><h2 class="view-title">系统设置</h2></div>
        <p class="view-desc">全局参数、网络代理、下载路径与基础运行引擎配置</p>
      </div>
      <AppButton type="primary" size="large" class="settings-save-btn" @click="handleSaveSettings">
        保存并应用设置
      </AppButton>
    </div>

    <AppCard class="clean-card settings-stack">
      <!-- 底层引擎与环境 -->
      <div class="settings-group">
        <div class="settings-group-title">底层引擎与环境</div>
        <div class="settings-field-row">
          <div class="field-meta">
            <div class="field-title">CLI 引擎状态</div>
            <div class="field-desc">App Store 服务已内置；自动检测 tools/ 目录与系统环境变量中的 go-ios</div>
          </div>
          <div class="field-control">
            <div class="engine-badge-box">
              <span class="engine-dot"></span>
              <span class="engine-text">引擎就绪</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 网络与代理 -->
      <div class="settings-group">
        <div class="settings-group-title">网络与代理</div>
        <div class="settings-field-row">
          <div class="field-meta">
            <div class="field-title">启用网络代理</div>
            <div class="field-desc">App Store 接口认证时通过环境变量透传 HTTP / SOCKS5 代理</div>
          </div>
          <div class="field-control">
            <AppSwitch
              v-model="settings.enableProxy"
              @update:model-value="onProxyToggle"
            />
          </div>
        </div>
        <div class="settings-field-row">
          <div class="field-meta">
            <div class="field-title">代理服务器地址</div>
            <div class="field-desc">支持 Clash / v2rayN 等常见本地代理端口（如 http://127.0.0.1:10808）</div>
          </div>
          <div class="field-control">
            <div class="setting-input-action-group">
              <AppInput
                v-model="settings.proxyUrl"
                :disabled="!settings.enableProxy"
                size="small"
                class="flex-1"
                placeholder="http://127.0.0.1:10808"
              />
              <AppButton
                variant="secondary"
                size="small"
                class="setting-fixed-btn"
                :disabled="!settings.enableProxy"
                :loading="isTestingProxy"
                @click="handleTestProxy"
              >
                测试
              </AppButton>
            </div>
          </div>
        </div>
      </div>

      <!-- 存储与下载 -->
      <div class="settings-group">
        <div class="settings-group-title">存储与下载</div>
        <div class="settings-field-row">
          <div class="field-meta">
            <div class="field-title">IPA 存储路径</div>
            <div class="field-desc">按登录账号动态归档至 data/downloads/账号标识 目录</div>
          </div>
          <div class="field-control">
            <div class="setting-input-action-group">
              <AppInput
                :model-value="settings.defaultDownloadDir"
                readonly
                disabled
                size="small"
                class="flex-1"
              />
              <AppButton
                variant="secondary"
                size="small"
                class="setting-fixed-btn"
                @click="handleOpenDefaultDownloadDir"
              >
                打开
              </AppButton>
            </div>
          </div>
        </div>
        <div class="settings-field-row">
          <div class="field-meta">
            <div class="field-title">默认目标平台</div>
            <div class="field-desc">检索与下载默认针对的 Apple 硬件体系</div>
          </div>
          <div class="field-control">
            <AppSelect
              v-model="settings.defaultPlatform"
              :options="platformOptions"
              size="small"
              style="width: 100%;"
            />
          </div>
        </div>
      </div>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { backend as main } from '../../../wailsjs/go/models'
import {
  GetSettings,
  SaveSettings,
  TestProxy,
  OpenInExplorer
} from '../../../wailsjs/go/backend/App'
import AppButton from '../../ui/components/AppButton.vue'
import AppCard from '../../ui/components/AppCard.vue'
import AppInput from '../../ui/components/AppInput.vue'
import AppSelect from '../../ui/components/AppSelect.vue'
import AppSwitch from '../../ui/components/AppSwitch.vue'
import { useAppMessage } from '../../ui/feedback'

const emit = defineEmits<{
  (e: 'settingsChanged', settings: main.Settings): void
  (e: 'busyChange', busy: boolean, text?: string): void
}>()

const message = useAppMessage()
const isTestingProxy = ref(false)

const platformOptions = [
  { label: 'iPhone (iOS)', value: 'iphone' },
  { label: 'iPad (iPadOS)', value: 'ipad' },
  { label: 'Apple TV (tvOS)', value: 'appletv' },
  { label: 'VisionOS', value: 'visionos' }
]

const settings = ref<main.Settings>({
  keychainPassphrase: '',
  defaultDownloadDir: 'data/downloads/default',
  defaultPlatform: 'iphone',
  enableProxy: true,
  proxyUrl: 'http://127.0.0.1:10808',
  ipaToolPath: ''
})

/**
 * 从后端载入应用配置
 */
async function loadSettings() {
  try {
    const s = await GetSettings()
    settings.value = s
    emit('settingsChanged', s)
  } catch (err: any) {
    console.error('加载设置失败:', err)
  }
}

/**
 * 网络代理快速开关
 */
async function onProxyToggle(val: boolean) {
  settings.value.enableProxy = val
  await SaveSettings(settings.value)
  emit('settingsChanged', settings.value)
  message.info(val ? '已开启网络代理' : '已关闭网络代理')
}

/**
 * 测试网络代理连通性
 */
async function handleTestProxy() {
  isTestingProxy.value = true
  emit('busyChange', true, '正在测试网络代理连接...')
  try {
    const res = await TestProxy(settings.value.proxyUrl)
    if (res.success) {
      message.success(res.message)
      emit('busyChange', false, '网络代理测试成功！')
    } else {
      message.error(res.message)
      emit('busyChange', false, '网络代理测试失败')
    }
  } catch (err: any) {
    message.error(`测试异常: ${err}`)
  } finally {
    isTestingProxy.value = false
  }
}

/**
 * 保存全局设置
 */
async function handleSaveSettings() {
  try {
    await SaveSettings(settings.value)
    emit('settingsChanged', settings.value)
    message.success('设置已保存并生效！')
  } catch (err: any) {
    message.error(`保存失败: ${err}`)
  }
}

/**
 * 在系统文件管理器中打开默认下载目录
 */
function handleOpenDefaultDownloadDir() {
  OpenInExplorer(settings.value.defaultDownloadDir || '')
}

defineExpose({
  settings,
  loadSettings,
  onProxyToggle,
  handleOpenDefaultDownloadDir
})
</script>
