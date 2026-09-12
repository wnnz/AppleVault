<template>
  <div class="view-panel">
    <div class="view-header">
      <div>
        <div class="title-with-badge"><h2 class="view-title">设备直装</h2></div>
        <p class="view-desc">通过 USB 数据线将正版签署的 IPA 应用包一键安装至 iOS 设备</p>
      </div>
    </div>

    <div class="installer-grid">
      <AppCard class="clean-card flex-col">
        <div class="card-headline">
          <span class="headline-title">1. 选择安装包</span>
          <span v-if="selectedIPAPath" class="headline-badge">已就绪</span>
        </div>
        <div class="clean-drop-zone" :class="{ 'drop-active': !!selectedIPAPath }" @click="emit('selectIPA')">
          <div class="drop-primary-title">{{ selectedIPAPath ? selectedIPAFileName : '点击选择或拖拽 .ipa 文件到此处' }}</div>
          <div class="drop-secondary-path text-ellipsis" :title="selectedIPAPath">{{ selectedIPAPath || '支持标准 iOS 签名应用包格式' }}</div>
        </div>
        <div class="mt-3 flex-align-center">
          <AppInput :model-value="selectedIPAPath" readonly size="small" class="flex-1 mr-2" placeholder="尚未选择文件" />
          <AppButton secondary size="small" @click="emit('selectIPA')">浏览文件</AppButton>
        </div>
      </AppCard>

      <AppCard class="clean-card flex-col">
        <div class="card-headline">
          <span class="headline-title">2. 选择苹果设备</span>
          <AppButton secondary size="small" :loading="isLoadingDevices" :disabled="isInstallingIPA" @click="emit('refreshDevices')">刷新检测</AppButton>
        </div>
        <AppSelect
          v-model="selectedDeviceUDID"
          :options="deviceOptions"
          :loading="isLoadingDevices"
          :disabled="isInstallingIPA"
          placeholder="请选择已连接的 iOS 设备"
          size="small"
        />
        <div class="sub-alert-box mt-3"><span>请保持设备屏幕<b>常亮解锁</b>；若设备息屏休眠，USB 通信将中断并丢失连接。</span></div>
        <div v-if="selectedDevice" class="device-spec-box mt-3">
          <div class="spec-row"><span class="spec-k">设备名称</span><span class="spec-v font-bold">{{ selectedDevice.name }}</span></div>
          <div class="spec-row"><span class="spec-k">设备型号</span><span class="spec-v">{{ selectedDevice.productType || '-' }}</span></div>
          <div class="spec-row"><span class="spec-k">系统版本</span><span class="spec-v">{{ selectedDevice.productVersion || '-' }}</span></div>
          <div class="spec-row"><span class="spec-k">连接模式</span><span class="spec-v">{{ selectedDevice.connectionType || '-' }}</span></div>
          <div class="spec-row"><span class="spec-k">UDID</span><span class="spec-v text-ellipsis" :title="selectedDevice.udid">{{ selectedDevice.udid }}</span></div>
        </div>
        <div v-else-if="!isLoadingDevices && devices.length === 0" class="no-device-box mt-3">
          <div class="no-device-text">未检测到已连接的苹果设备</div>
          <div class="no-device-sub">请直连电脑 USB 接口、点亮屏幕、输入密码并信任此电脑</div>
        </div>
      </AppCard>
    </div>

    <AppCard class="clean-card mt-4 installer-action-banner">
      <div class="installer-action-info">
        <div class="action-banner-title">{{ !selectedIPAPath ? '请先选择待安装的 IPA 文件' : (!selectedDeviceUDID ? '请选择目标苹果设备' : '就绪，可以开始安装') }}</div>
        <div class="action-banner-desc">本工具下载的正版 IPA 需安装至登录了相同 Apple ID 的设备上，未签名包将无法被系统接受。</div>
      </div>
      <AppButton
        type="primary"
        size="large"
        :disabled="!selectedIPAPath || !selectedDeviceUDID || isLoadingDevices"
        :loading="isInstallingIPA"
        class="install-submit-btn"
        @click="emit('installIPA')"
      >
        {{ isInstallingIPA ? '正在安装中...' : '开始安装到设备' }}
      </AppButton>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import AppButton from '../../ui/components/AppButton.vue'
import AppCard from '../../ui/components/AppCard.vue'
import AppInput from '../../ui/components/AppInput.vue'
import AppSelect, { type AppSelectOption } from '../../ui/components/AppSelect.vue'

interface DeviceInfo {
  udid: string
  name: string
  productType: string
  productVersion: string
  connectionType: string
}

defineProps<{
  selectedIPAPath: string
  selectedIPAFileName: string
  devices: DeviceInfo[]
  selectedDevice: DeviceInfo | null | undefined
  deviceOptions: AppSelectOption[]
  isLoadingDevices: boolean
  isInstallingIPA: boolean
}>()

const selectedDeviceUDID = defineModel<string | null>('selectedDeviceUDID', { required: true })
const emit = defineEmits<{
  selectIPA: []
  refreshDevices: []
  installIPA: []
}>()
</script>
