<template>
  <div class="view-panel">
    <div class="view-header">
      <div>
        <div class="title-with-badge"><h2 class="view-title">设备直装</h2></div>
        <p class="view-desc">通过 USB 或 Wi-Fi 将正版签署的 IPA 应用包一键安装至 iOS 设备</p>
      </div>
    </div>

    <div class="installer-grid">
      <!-- 1. 选择安装包 -->
      <AppCard class="clean-card flex-col">
        <div class="card-headline">
          <span class="headline-title">1. 选择安装包</span>
          <span v-if="selectedIPAPath" class="headline-badge">已就绪</span>
        </div>
        <div class="clean-drop-zone" :class="{ 'drop-active': !!selectedIPAPath }" @click="handleSelectIPA">
          <div class="drop-primary-title">{{ selectedIPAPath ? selectedIPAFileName : '点击选择或拖拽 .ipa 文件到此处' }}</div>
          <div class="drop-secondary-path text-ellipsis" :title="selectedIPAPath">{{ selectedIPAPath || '支持标准 iOS 签名应用包格式' }}</div>
        </div>
        <div class="mt-3 flex-align-center">
          <AppInput :model-value="selectedIPAPath" readonly size="small" class="flex-1 mr-2" placeholder="尚未选择文件" />
          <AppButton secondary size="small" @click="handleSelectIPA">浏览文件</AppButton>
        </div>
        <div v-if="isInspectingIPA" class="ipa-inspection-state mt-3">正在校验 IPA 完整性与兼容性...</div>
        <div v-else-if="ipaInspection" class="ipa-inspection-box mt-3" :class="ipaInspection.compatible ? 'is-compatible' : 'is-incompatible'">
          <div class="ipa-inspection-title">
            <span>{{ ipaInspection.appName || selectedIPAFileName }}</span>
            <span class="headline-badge">{{ ipaInspection.compatible ? '检查通过' : '需要处理' }}</span>
          </div>
          <div class="spec-row"><span class="spec-k">Bundle ID</span><span class="spec-v text-ellipsis" :title="ipaInspection.bundleID">{{ ipaInspection.bundleID || '-' }}</span></div>
          <div class="spec-row"><span class="spec-k">应用版本</span><span class="spec-v">{{ ipaInspection.version || '-' }}<template v-if="ipaInspection.buildVersion">（Build {{ ipaInspection.buildVersion }}）</template></span></div>
          <div class="spec-row"><span class="spec-k">最低系统</span><span class="spec-v">{{ ipaInspection.minimumOSVersion || '未声明' }}</span></div>
          <div class="spec-row"><span class="spec-k">支持设备</span><span class="spec-v">{{ ipaInspection.supportedDeviceTypes?.join('、') || '未限制' }}</span></div>
          <div class="spec-row"><span class="spec-k">签名结构</span><span class="spec-v">{{ ipaInspection.signed ? '已检测到' : '未检测到' }}</span></div>
          <div class="spec-row"><span class="spec-k">文件大小</span><span class="spec-v">{{ ipaInspection.displayFileSize || '-' }}</span></div>
          <div class="ipa-compatibility-message">{{ ipaInspection.compatibilityMessage }}</div>
        </div>
        <div v-else-if="ipaInspectionError" class="ipa-inspection-state is-error mt-3">{{ ipaInspectionError }}</div>
      </AppCard>

      <!-- 2. 选择苹果设备 -->
      <AppCard class="clean-card flex-col">
        <div class="card-headline">
          <span class="headline-title">2. 选择苹果设备</span>
          <div class="device-card-actions">
            <AppButton
              secondary
              size="small"
              :loading="isPairingWiFi"
              :disabled="!selectedDevice || selectedDevice.connectionType === 'Wi-Fi' || isInstallingIPA"
              @click="handlePairWiFi"
            >
              Wi-Fi 配对
            </AppButton>
            <AppButton
              secondary
              size="small"
              :loading="isLoadingDevices"
              :disabled="isInstallingIPA || isPairingWiFi"
              @click="loadConnectedDevices"
            >
              刷新检测
            </AppButton>
          </div>
        </div>
        <AppSelect
          v-model="selectedDeviceUDID"
          :options="deviceOptions"
          :loading="isLoadingDevices"
          :disabled="isInstallingIPA"
          placeholder="请选择已连接的 iOS 设备"
          size="small"
        />
        <div class="sub-alert-box mt-3">
          <span>请保持设备屏幕<b>常亮解锁</b>；Wi-Fi 连接要求设备与电脑处于同一局域网。</span>
        </div>
        <div v-if="selectedDevice" class="device-spec-box mt-3">
          <div class="spec-row"><span class="spec-k">设备名称</span><span class="spec-v font-bold">{{ selectedDevice.name }}</span></div>
          <div class="spec-row"><span class="spec-k">设备型号</span><span class="spec-v">{{ selectedDevice.productType || '-' }}</span></div>
          <div class="spec-row"><span class="spec-k">系统版本</span><span class="spec-v">{{ selectedDevice.productVersion || '-' }}</span></div>
          <div class="spec-row"><span class="spec-k">连接模式</span><span class="spec-v">{{ selectedDevice.connectionType || '-' }}</span></div>
          <div class="spec-row"><span class="spec-k">UDID</span><span class="spec-v text-ellipsis" :title="selectedDevice.udid">{{ selectedDevice.udid }}</span></div>
        </div>
        <div v-else-if="!isLoadingDevices && devices.length === 0" class="no-device-box mt-3">
          <div class="no-device-text">未检测到已连接的苹果设备</div>
          <div class="no-device-sub">首次连接请使用 USB 完成信任；已配对设备可在同一 Wi-Fi 下直接刷新</div>
        </div>
      </AppCard>
    </div>

    <!-- 底部直装操作区 -->
    <AppCard class="clean-card mt-4 installer-action-banner">
      <div class="installer-action-info">
        <div class="action-banner-title">
          {{ !selectedIPAPath ? '请先选择待安装的 IPA 文件' : (!selectedDeviceUDID ? '请选择目标苹果设备' : (ipaInspection && !ipaInspection.compatible ? 'IPA 与所选设备不兼容' : '就绪，可以开始安装')) }}
        </div>
        <div class="action-banner-desc">本工具下载的正版 IPA 需安装至登录了相同 Apple ID 的设备上，未签名包将无法被系统接受。</div>
      </div>
      <AppButton
        type="primary"
        size="large"
        :disabled="!selectedIPAPath || !selectedDeviceUDID || isLoadingDevices || isInspectingIPA || !!(ipaInspection && !ipaInspection.compatible)"
        :loading="isInstallingIPA"
        class="install-submit-btn"
        @click="handleInstallIPA"
      >
        {{ isInstallingIPA ? '正在安装中...' : '开始安装到设备' }}
      </AppButton>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { backend as main } from '../../../wailsjs/go/models'
import { SelectIPA, ListDevices, InstallIPA, InspectIPA, PairDeviceForWiFi } from '../../../wailsjs/go/backend/App'
import AppButton from '../../ui/components/AppButton.vue'
import AppCard from '../../ui/components/AppCard.vue'
import AppInput from '../../ui/components/AppInput.vue'
import AppSelect from '../../ui/components/AppSelect.vue'
import { useAppDialog, useAppMessage } from '../../ui/feedback'

const emit = defineEmits<{
  (e: 'busyChange', busy: boolean, text?: string): void
}>()

const message = useAppMessage()
const dialog = useAppDialog()

const selectedIPAPath = ref('')
const devices = ref<main.DeviceInfo[]>([])
const selectedDeviceUDID = ref<string | null>(null)
const isLoadingDevices = ref(false)
const isInstallingIPA = ref(false)
const isPairingWiFi = ref(false)
const isInspectingIPA = ref(false)
const ipaInspection = ref<main.IPAInspectionResult | null>(null)
const ipaInspectionError = ref('')
let inspectionSequence = 0
const isDemoMode = new URLSearchParams(window.location.search).get('demo') === '1'

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

async function inspectSelectedIPA() {
  const ipaPath = selectedIPAPath.value
  const device = selectedDevice.value
  const sequence = ++inspectionSequence
  if (isDemoMode) return ipaInspection.value
  ipaInspection.value = null
  ipaInspectionError.value = ''
  if (!ipaPath) return null

  isInspectingIPA.value = true
  try {
    const result = await InspectIPA(ipaPath, device?.productType || '', device?.productVersion || '')
    if (sequence === inspectionSequence) ipaInspection.value = result
    return result
  } catch (err: any) {
    if (sequence === inspectionSequence) ipaInspectionError.value = `IPA 检查失败: ${err}`
    return null
  } finally {
    if (sequence === inspectionSequence) isInspectingIPA.value = false
  }
}

watch([selectedIPAPath, () => selectedDevice.value?.productType, () => selectedDevice.value?.productVersion], () => {
  void inspectSelectedIPA()
})

/**
 * 接收来自拖拽或外部选定的 IPA 文件路径
 */
function useIPAPath(paths: string[]) {
  const ipaPath = paths.find(path => path.toLowerCase().endsWith('.ipa'))
  if (!ipaPath) {
    message.warning('请拖放 .ipa 格式的安装包')
    return
  }
  selectedIPAPath.value = ipaPath
  emit('busyChange', false, `已选定 IPA: ${ipaPath}`)
}

/**
 * 打开系统文件选择框选取 IPA
 */
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

/**
 * 检测当前已连接的真机设备
 */
async function loadConnectedDevices() {
  isLoadingDevices.value = true
  emit('busyChange', true, '正在检测已连接的苹果设备...')
  try {
    const result = await ListDevices()
    devices.value = result || []
    if (!devices.value.some(device => device.udid === selectedDeviceUDID.value)) {
      selectedDeviceUDID.value = devices.value.length === 1 ? devices.value[0].udid : null
    }
    if (devices.value.length === 0) {
      emit('busyChange', false, '未发现可用苹果设备')
      message.warning('未发现设备，请确认设备已连接、解锁并信任此电脑')
    } else {
      emit('busyChange', false, `发现 ${devices.value.length} 台可用设备`)
    }
  } catch (err: any) {
    devices.value = []
    selectedDeviceUDID.value = null
    emit('busyChange', false, '设备检测失败')
    message.error(`设备检测失败: ${err}`)
  } finally {
    isLoadingDevices.value = false
  }
}

function handlePairWiFi() {
  if (!selectedDevice.value) {
    message.warning('请先选择通过 USB 连接的设备')
    return
  }
  dialog.info({
    title: '配置 Wi-Fi 连接',
    content: '请保持设备通过 USB 连接、解锁并信任此电脑。配对完成后，还需在 Apple Devices 或 iTunes 中开启“连接 Wi-Fi 时显示此设备”。',
    positiveText: '开始配对',
    negativeText: '取消',
    onPositiveClick: pairSelectedDeviceForWiFi
  })
}

async function pairSelectedDeviceForWiFi() {
  if (!selectedDeviceUDID.value) return
  isPairingWiFi.value = true
  emit('busyChange', true, '正在建立 Wi-Fi 设备配对...')
  try {
    await PairDeviceForWiFi(selectedDeviceUDID.value)
    message.success('配对完成，请在 Apple Devices/iTunes 开启 Wi-Fi 显示后拔线刷新')
    emit('busyChange', false, 'Wi-Fi 配对完成')
  } catch (err: any) {
    message.error(`Wi-Fi 配对失败: ${err}`)
    emit('busyChange', false, 'Wi-Fi 配对失败')
  } finally {
    isPairingWiFi.value = false
  }
}

/**
 * 执行设备安装操作
 */
async function handleInstallIPA() {
  if (!selectedIPAPath.value) {
    message.warning('请先选择 IPA 文件')
    return
  }
  if (!selectedDeviceUDID.value) {
    message.warning('请选择要安装的苹果设备')
    return
  }

  const inspection = await inspectSelectedIPA()
  if (!inspection) {
    message.error(ipaInspectionError.value || 'IPA 完整性检查失败')
    return
  }
  if (!inspection.compatible) {
    message.error(inspection.compatibilityMessage || 'IPA 与所选设备不兼容')
    return
  }

  isInstallingIPA.value = true
  emit('busyChange', true, `正在安装 ${selectedIPAFileName.value}（请保持设备屏幕常亮勿息屏）...`)
  message.info('开始安装应用，请确保设备屏幕保持常亮解锁...')

  try {
    const result = await InstallIPA(selectedIPAPath.value, selectedDeviceUDID.value)
    if (result.success) {
      emit('busyChange', false, result.message)
      message.success(result.message)
    } else {
      emit('busyChange', false, 'IPA 安装失败')
      message.error(result.message || 'IPA 安装失败')
    }
  } catch (err: any) {
    emit('busyChange', false, 'IPA 安装失败')
    message.error(`IPA 安装失败: ${err}`)
  } finally {
    isInstallingIPA.value = false
  }
}

defineExpose({
  selectedIPAPath,
  devices,
  selectedDeviceUDID,
  ipaInspection,
  useIPAPath,
  loadConnectedDevices
})
</script>
