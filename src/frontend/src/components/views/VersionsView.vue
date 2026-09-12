<template>
  <div class="view-panel">
    <div class="view-header">
      <div>
        <div class="title-with-badge"><h2 class="view-title">历史版本</h2></div>
        <p class="view-desc">解析与枚举 App Store 所有历史构建版本，支持快速精准检索</p>
      </div>
      <div class="header-right-badges">
        <span class="count-pill">共 {{ filteredVersions.length }} / {{ versionItems.length }} 条记录</span>
      </div>
    </div>

    <AppCard class="clean-card mb-4">
      <div class="versions-toolbar">
        <div class="bundle-input-row">
          <AppInputGroup
            v-model="versionForm.bundleId"
            prefix="Bundle ID"
            size="small"
            placeholder="例如: com.alipay.iphoneclient"
            @keydown.enter="emit('listVersions')"
          />
          <AppButton
            type="primary"
            size="small"
            :loading="isListingVersions"
            :disabled="isListingVersions || isBatchQuerying || isTargetQuerying"
            @click="emit('listVersions')"
          >
            获取历史版本
          </AppButton>
          <AppButton
            :type="isBatchQuerying ? 'error' : 'default'"
            :secondary="!isBatchQuerying"
            size="small"
            :loading="isBatchQuerying"
            :disabled="versionItems.length === 0 || isListingVersions || isTargetQuerying"
            @click="emit('batchQuery')"
          >
            {{ isBatchQuerying ? '停止查询' : '批量解析前 30 项' }}
          </AppButton>
          <AppButton
            :type="isTargetQuerying ? 'error' : 'default'"
            :secondary="!isTargetQuerying"
            size="small"
            :loading="isTargetQuerying"
            :disabled="versionItems.length === 0 || isListingVersions || isBatchQuerying"
            @click="emit('targetQuery')"
          >
            {{ isTargetQuerying ? '停止查询' : '查找指定版本' }}
          </AppButton>
        </div>

        <div class="filter-search-row">
          <div class="filter-search-box">
            <AppInput
              v-model="versionForm.filter"
              size="small"
              placeholder="输入版本号 (如 10.2.80)、构建 ID 或体积进行实时筛选..."
            />
          </div>
        </div>
      </div>
    </AppCard>

    <AppCard class="clean-card table-flex-card">
      <AppTable
        :columns="columns"
        :data="filteredVersions"
        :loading="isListingVersions"
        :scroll-x="720"
        flex-height
        style="height: 100%;"
        size="small"
      >
        <template #empty>
          <div class="table-empty-box">
            <div class="empty-state-title">{{ versionForm.bundleId ? '未查询到版本信息' : '历史版本时光机' }}</div>
            <div class="empty-state-desc">{{ versionForm.bundleId ? '请检查 Bundle ID 是否正确，或当前账号是否具备权限' : '输入应用 Bundle ID 并点击「获取历史版本」，即可枚举所有构建历史' }}</div>
          </div>
        </template>
      </AppTable>
    </AppCard>

    <AppDialog v-model="showTargetVersionModal" title="查找指定版本" style="width: 440px; border-radius: 14px;">
      <div class="modal-dialog-inner">
        <p class="dialog-desc">
          请输入目标版本号（例如 <code>10.2.80</code> 或 <code>8.0.0</code>），程序将智能检索历史构建记录并快速定位匹配版本。
        </p>
        <AppInput
          ref="targetVersionInputRef"
          v-model="targetVersionInput"
          placeholder="例如: 10.2.80"
          autofocus
          @keydown.enter="emit('confirmTargetQuery')"
        />
        <div class="dialog-action-buttons mt-4">
          <AppButton secondary @click="showTargetVersionModal = false">取消</AppButton>
          <AppButton type="primary" :disabled="!targetVersionInput.trim()" @click="emit('confirmTargetQuery')">开始查询</AppButton>
        </div>
      </div>
    </AppDialog>
  </div>
</template>

<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import AppButton from '../../ui/components/AppButton.vue'
import AppCard from '../../ui/components/AppCard.vue'
import AppDialog from '../../ui/components/AppDialog.vue'
import AppInput from '../../ui/components/AppInput.vue'
import AppInputGroup from '../../ui/components/AppInputGroup.vue'
import AppTable from '../../ui/components/AppTable.vue'

interface VersionItem {
  versionId: string
  displayVersion: string
  fileSize: string
  releaseDate: string
  isQuerying?: boolean
}

defineProps<{
  versionForm: { bundleId: string; appId: number; appName: string; filter: string }
  versionItems: VersionItem[]
  filteredVersions: VersionItem[]
  columns: any[]
  isListingVersions: boolean
  isBatchQuerying: boolean
  isTargetQuerying: boolean
}>()

const showTargetVersionModal = defineModel<boolean>('showTargetVersionModal', { required: true })
const targetVersionInput = defineModel<string>('targetVersionInput', { required: true })
const targetVersionInputRef = ref<InstanceType<typeof AppInput> | null>(null)

const emit = defineEmits<{
  listVersions: []
  batchQuery: []
  targetQuery: []
  confirmTargetQuery: []
}>()

watch(showTargetVersionModal, async visible => {
  if (!visible) return
  await nextTick()
  targetVersionInputRef.value?.focus()
})
</script>
