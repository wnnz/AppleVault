import { ref } from 'vue'
import { backend as main } from '../../wailsjs/go/models'
import {
  GetSettings,
  SaveSettings,
  SelectDirectory,
  TestProxy,
  OpenInExplorer
} from '../../wailsjs/go/backend/App'
import { useAppMessage } from '../ui/feedback'

export const platformOptions = [
  { label: 'iPhone (iOS)', value: 'iphone' },
  { label: 'iPad (iPadOS)', value: 'ipad' },
  { label: 'Apple TV (tvOS)', value: 'appletv' },
  { label: 'VisionOS', value: 'visionos' }
]

export const limitOptions = [
  { label: '5 个', value: 5 },
  { label: '10 个', value: 10 },
  { label: '20 个', value: 20 },
  { label: '50 个', value: 50 }
]

export function useSettings(options?: {
  onStatusChange?: (status: string) => void
}) {
  const message = useAppMessage()
  const isTestingProxy = ref(false)

  const settings = ref<main.Settings>({
    keychainPassphrase: '123456',
    defaultDownloadDir: 'data/downloads/default',
    defaultPlatform: 'iphone',
    enableProxy: true,
    proxyUrl: 'http://127.0.0.1:10808',
    ipaToolPath: ''
  })

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
    options?.onStatusChange?.('正在测试网络代理连接...')
    try {
      const res = await TestProxy(settings.value.proxyUrl)
      if (res.success) {
        message.success(res.message)
        options?.onStatusChange?.('网络代理测试成功！')
      } else {
        message.error(res.message)
        options?.onStatusChange?.('网络代理测试失败')
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

  function handleOpenDefaultDownloadDir() {
    OpenInExplorer(settings.value.defaultDownloadDir || '')
  }

  return {
    settings,
    isTestingProxy,
    platformOptions,
    limitOptions,
    loadSettings,
    onProxyToggle,
    handleSelectDefaultDir,
    handleTestProxy,
    handleSaveSettings,
    handleOpenDefaultDownloadDir
  }
}
