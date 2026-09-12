<template>
  <div class="view-panel">
    <div class="view-header">
      <div>
        <div class="title-with-badge"><h2 class="view-title">应用搜索</h2></div>
        <p class="view-desc">在 Apple App Store 全球库中精准检索正版应用信息</p>
      </div>
    </div>

    <AppCard class="clean-card mb-4 search-bar-card">
      <div class="search-input-group">
        <div class="search-input-wrapper">
          <AppInput
            v-model="searchForm.term"
            size="small"
            placeholder="输入应用名称、关键字或开发商（如：微信、支付宝、TikTok）"
            @keydown.enter="handleSearch"
          />
        </div>
        <div class="filter-controls">
          <div class="filter-item">
            <span class="filter-label">平台</span>
            <AppSelect
              v-model="searchForm.platform"
              :options="platformOptions"
              size="small"
              style="width: 130px;"
            />
          </div>
          <div class="filter-item">
            <span class="filter-label">条数</span>
            <AppSelect
              v-model="searchForm.limit"
              :options="limitOptions"
              size="small"
              style="width: 90px;"
            />
          </div>
          <AppButton
            variant="primary"
            size="small"
            :loading="isSearching"
            @click="handleSearch"
          >
            搜索
          </AppButton>
        </div>
      </div>
    </AppCard>

    <AppCard class="clean-card table-flex-card">
      <AppTable
        :columns="displayedSearchColumns"
        :data="searchResults"
        :loading="isSearching"
        :pagination="{ pageSize: 10 }"
        :scroll-x="780"
        size="small"
        flex-height
        style="height: 100%;"
      >
        <template #empty>
          <div class="table-empty-box">
            <div class="empty-state-title">{{ searchForm.term ? '未找到匹配的应用' : '开启 App Store 探索之旅' }}</div>
            <div class="empty-state-desc">{{ searchForm.term ? '请尝试更换关键词，或切换不同国家/地区搜索' : '输入应用名称、开发商或拼音，随时开始检索正版应用' }}</div>
          </div>
        </template>
      </AppTable>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, h } from 'vue'
import { backend as main } from '../../../wailsjs/go/models'
import { Search, Purchase } from '../../../wailsjs/go/backend/App'
import AppButton from '../../ui/components/AppButton.vue'
import AppCard from '../../ui/components/AppCard.vue'
import AppInput from '../../ui/components/AppInput.vue'
import AppSelect from '../../ui/components/AppSelect.vue'
import AppTable from '../../ui/components/AppTable.vue'
import { useAppMessage } from '../../ui/feedback'
import type { AppTheme } from '../../ui/theme'

const props = defineProps<{
  effectiveTheme?: AppTheme
}>()

const emit = defineEmits<{
  (e: 'selectAppForVersions', app: main.AppItem): void
  (e: 'downloadApp', app: main.AppItem): void
  (e: 'busyChange', busy: boolean, text?: string): void
}>()

const message = useAppMessage()

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

const isSearching = ref(false)
const searchForm = ref({
  term: '',
  limit: 10,
  platform: 'iphone'
})
const searchResults = ref<main.AppItem[]>([])

/**
 * 检索应用列表
 */
async function handleSearch() {
  if (!searchForm.value.term) {
    message.warning('请输入搜索关键词')
    return
  }

  isSearching.value = true
  emit('busyChange', true, `正在搜索 "${searchForm.value.term}"...`)

  try {
    const res = await Search(searchForm.value.term, searchForm.value.limit, searchForm.value.platform)
    searchResults.value = res.apps || []
    emit('busyChange', false, `搜索完成，找到 ${searchResults.value.length} 个应用。`)
  } catch (err: any) {
    message.error(`搜索失败: ${err}`)
    emit('busyChange', false, '搜索失败')
  } finally {
    isSearching.value = false
  }
}

/**
 * 获取应用正版授权许可
 */
async function handlePurchaseApp(bundleId: string) {
  emit('busyChange', true, `正在获取应用授权: ${bundleId}...`)
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
    emit('busyChange', false)
  }
}

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
        h(AppButton, { size: 'tiny', secondary: true, onClick: () => emit('selectAppForVersions', row) }, () => '历史版本'),
        h(AppButton, { size: 'tiny', type: 'primary', onClick: () => emit('downloadApp', row) }, () => '下载'),
        h(AppButton, { size: 'tiny', secondary: true, onClick: () => handlePurchaseApp(row.bundleID) }, () => '获取许可')
      ])
    }
  }
]

const displayedSearchColumns = computed(() => {
  if (props.effectiveTheme !== 'handdrawn') return searchColumns

  const widths = [196, 163, 94, 87, 83, 217]
  return searchColumns.map((column, index) => ({
    ...column,
    minWidth: undefined,
    width: widths[index]
  }))
})

defineExpose({
  searchForm,
  searchResults,
  handleSearch
})
</script>
