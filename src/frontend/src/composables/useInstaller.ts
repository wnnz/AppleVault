import { ref, computed } from 'vue'
import { backend as main } from '../../wailsjs/go/models'
import { SelectIPA, ListDevices, InstallIPA } from '../../wailsjs/go/backend/App'
import { useAppMessage } from '../ui/feedback'

export function useInstaller(options?: {
  onBusyChange?: (busy: boolean, text?: string) => void
  onNavigateToInstaller?: () => void
}) {
  const message = useAppMessage()

  const selectedIPAPath = ref('')
  const devices = ref<main.DeviceInfo[]>([])
  const selectedDeviceUDID = ref<string | null>(null)
  const isLoadingDevices = ref(false)
  const isInstallingIPA = ref(false)

  const selectedIPAFileName = computed(() =>
    selectedIPAPath.value.split(/[\\/]/).pop() || selectedIPAPath.value
  )

  const selectedDevice = computed(() =>
    devices.value.find(device => device.udid === selectedDeviceUDID.value)
  )

  const deviceOptions = computed(() =>
    devices.value.map(device => {
      const modelShort = device.productType
      const osPrefix = device.productType.toLowerCase().includes('ipad') ? 'iPadOS' : 'iOS'
      const versionStr = device.productVersion ? `${osPrefix} ${device.productVersion}` : ''
      const details = (
        device.name === device.productType
          ? [versionStr, device.connectionType]
          : [modelShort, versionStr, device.connectionType]
      ).filter(Boolean).join(' · ')
      return {
        label: details ? `${device.name} (${details})` : device.name,
        value: device.udid
      }
    })
  )

  function useIPAPath(paths: string[]) {
    const ipaPath = paths.find(path => path.toLowerCase().endsWith('.ipa'))
    if (!ipaPath) {
      message.warning('请拖放 .ipa 格式的安装包')
      return
    }
    selectedIPAPath.value = ipaPath
    options?.onNavigateToInstaller?.()
    options?.onBusyChange?.(false, `已选定 IPA: ${ipaPath}`)
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
    options?.onBusyChange?.(true, '正在检测已连接的苹果设备...')
    try {
      const result = await ListDevices()
      devices.value = result || []
      if (!devices.value.some(device => device.udid === selectedDeviceUDID.value)) {
        selectedDeviceUDID.value = devices.value.length === 1 ? devices.value[0].udid : null
      }
      if (devices.value.length === 0) {
        options?.onBusyChange?.(false, '未发现可用苹果设备')
        message.warning('未发现设备，请确认设备已连接、解锁并信任此电脑')
      } else {
        options?.onBusyChange?.(false, `发现 ${devices.value.length} 台可用设备`)
      }
    } catch (err: any) {
      devices.value = []
      selectedDeviceUDID.value = null
      options?.onBusyChange?.(false, '设备检测失败')
      message.error(`设备检测失败: ${err}`)
    } finally {
      isLoadingDevices.value = false
      options?.onBusyChange?.(false)
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
    options?.onBusyChange?.(true, `正在安装 ${selectedIPAFileName.value}（请保持设备屏幕常亮勿息屏）...`)
    message.info('开始安装应用，请确保设备屏幕保持常亮解锁...')

    try {
      const result = await InstallIPA(selectedIPAPath.value, selectedDeviceUDID.value)
      if (result.success) {
        options?.onBusyChange?.(false, result.message)
        message.success(result.message)
      } else {
        options?.onBusyChange?.(false, 'IPA 安装失败')
        message.error(result.message || 'IPA 安装失败')
      }
    } catch (err: any) {
      options?.onBusyChange?.(false, 'IPA 安装失败')
      message.error(`IPA 安装失败: ${err}`)
    } finally {
      isInstallingIPA.value = false
      options?.onBusyChange?.(false)
    }
  }

  return {
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
  }
}
