import type { Ref } from 'vue'
import { backend as main } from '../../wailsjs/go/models'
import type { VersionItem } from './useVersions'

export function applyDemoData(params: {
  tabParam: string | null
  account: Ref<main.AccountInfo>
  statusText: Ref<string>
  searchForm: Ref<{ term: string; limit: number; platform: string }>
  settings: Ref<main.Settings>
  searchResults: Ref<main.AppItem[]>
  versionForm: Ref<{ bundleId: string; appId: number; appName: string; filter: string }>
  versionItems: Ref<VersionItem[]>
  downloadTasks: Ref<main.DownloadTask[]>
  purchasedApps: Ref<main.AppItem[]>
  purchasedTotal: Ref<number>
  devices: Ref<main.DeviceInfo[]>
  selectedDeviceUDID: Ref<string | null>
  selectedIPAPath: Ref<string>
}) {
  const {
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
  } = params

  account.value = { name: '果仓助手用户', email: 'applevault.user@icloud.com', success: true }

  const demoStatusByTab: Record<string, string> = {
    search: '搜索完成，找到 6 个应用。',
    versions: '共获取 8 个历史版本记录。',
    download: '当前有 1 个任务正在下载，速度 12.8 MB/s',
    purchased: '已购应用加载完成（共 6 款）。',
    installer: '已就绪，已检测到 iPhone 15 Pro Max',
    account: '已登录 果仓助手用户（applevault.user@icloud.com）',
    settings: '底层引擎就绪，配置已加载',
    about: '果仓助手 (AppleVault) v1.2.0'
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
}
