<template>
  <div class="view-panel">
    <div class="view-header">
      <div>
        <div class="title-with-badge"><h2 class="view-title">已购应用</h2></div>
        <p class="view-desc">浏览与检索当前 Apple ID 名下已获得正版许可的历史应用库</p>
      </div>
      <div v-if="isPurchasedLoading" class="purchased-loading-progress" role="status" aria-live="polite">
        <AppProgress v-if="purchasedLoadTotal > 0" class="purchased-loading-bar" :percentage="purchasedLoadProgress" :height="6" />
        <div v-else class="purchased-loading-bar purchased-loading-bar-indeterminate" aria-hidden="true"></div>
        <span v-if="purchasedLoadTotal > 0" class="purchased-loading-count">{{ purchasedLoadLoaded }}/{{ purchasedLoadTotal }}</span>
        <span v-else class="purchased-loading-count loading-total-text">读取总数…</span>
      </div>
    </div>

    <AppCard class="clean-card mb-4 search-bar-card">
      <div class="search-input-group">
        <div class="search-input-wrapper">
          <AppSearchInput
            v-model="purchasedSearchKeyword"
            size="small"
            placeholder="搜索已购应用名称、Bundle ID 或 App ID..."
            @keydown.esc="purchasedSearchKeyword = ''"
          />
        </div>
        <div class="filter-controls">
          <AppButton type="primary" size="small" :loading="isPurchasedLoading" @click="loadPurchases">刷新列表</AppButton>
        </div>
      </div>
    </AppCard>

    <AppCard class="clean-card table-flex-card">
      <AppTable
        class="purchased-table"
        :columns="displayedPurchasedColumns"
        :data="filteredPurchasedApps"
        :loading="isPurchasedLoading"
        :scroll-x="720"
        size="small"
        flex-height
        style="height: 100%;"
      >
        <template #empty>
          <div class="purchased-empty-box">
            <div v-if="purchasedSearchKeyword" class="search-none-state">
              <div class="search-none-icon">🔍</div>
              <div class="search-none-title">未找到与「{{ purchasedSearchKeyword }}」匹配的已购应用</div>
              <div class="search-none-desc">请尝试输入不同关键词，或点击下方按钮清空搜索</div>
              <AppButton size="small" secondary class="mt-2" @click="purchasedSearchKeyword = ''">清空搜索</AppButton>
            </div>
            <div v-else class="search-none-state">
              <div class="search-none-title">暂无已购应用记录</div>
              <div class="search-none-desc">点击右上角「刷新列表」获取当前 Apple ID 历史正版应用</div>
            </div>
          </div>
        </template>
      </AppTable>
      <div class="purchased-bottom-pagination" aria-label="已购应用分页">
        <div class="purchased-search-badge purchased-footer-count" :class="{ 'has-filter': !!purchasedSearchKeyword }">
          <span v-if="purchasedSearchKeyword">找到 <b>{{ purchasedMatchTotal }}</b> 款匹配应用（共 <b>{{ purchasedApps.length }}</b> 款）</span>
          <span v-else>共 <b>{{ purchasedTotal }}</b> 款已购应用</span>
        </div>
        <div class="page-size-selector">
          <span class="size-label">每页</span>
          <AppSelect
            :model-value="purchasedPageSize"
            size="small"
            style="width: 76px;"
            :options="purchasedPageSizeOptions"
            @update:model-value="onPurchasedPageSizeChange"
          />
        </div>
        <AppButton secondary size="small" :disabled="purchasedPage <= 1 || isPurchasedLoading" @click="prevPurchasedPage">上一页</AppButton>
        <span class="page-indicator">第 {{ purchasedPage }} / {{ purchasedPageCount }} 页</span>
        <AppButton secondary size="small" :disabled="purchasedPage >= purchasedPageCount || isPurchasedLoading" @click="nextPurchasedPage">下一页</AppButton>
      </div>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, h } from 'vue'
import { backend as main } from '../../../wailsjs/go/models'
import { ListPurchases } from '../../../wailsjs/go/backend/App'
import AppButton from '../../ui/components/AppButton.vue'
import AppCard from '../../ui/components/AppCard.vue'
import AppProgress from '../../ui/components/AppProgress.vue'
import AppSearchInput from '../../ui/components/AppSearchInput.vue'
import AppSelect, { type SelectValue } from '../../ui/components/AppSelect.vue'
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

const isPurchasedLoading = ref(false)
const purchasedLoadLoaded = ref(0)
const purchasedLoadTotal = ref(0)
const purchasedPage = ref(1)
const purchasedPageSize = ref(20)
const purchasedTotal = ref(0)
const purchasedApps = ref<main.AppItem[]>([])
const purchasedSearchKeyword = ref('')

const purchasedPageSizeOptions = [
  { label: '20', value: 20 },
  { label: '50', value: 50 },
  { label: '100', value: 100 }
]

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

function formatReadableDate(raw?: string): string {
  if (!raw) return '-'
  return raw.replace(/T/, ' ').replace(/:\d{2}Z$/, '').replace(/Z$/, '')
}

/**
 * 分页并发获取 Apple ID 名下的所有已购记录
 */
async function loadPurchases() {
  isPurchasedLoading.value = true
  emit('busyChange', true, '正在获取已购应用总数...')
  purchasedLoadLoaded.value = 0
  purchasedLoadTotal.value = 0

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
        emit('busyChange', true, `正在加载已购应用 (${purchasedLoadLoaded.value}/${firstTotal})...`)
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
    emit('busyChange', false, `已购应用加载完成 (共 ${purchasedApps.value.length} 款)。`)
  } catch (err: any) {
    message.error(`加载已购列表失败: ${err}`)
    emit('busyChange', false, '加载失败')
  } finally {
    isPurchasedLoading.value = false
  }
}

function onPurchasedPageSizeChange(val: SelectValue) {
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
        h(AppButton, { size: 'tiny', secondary: true, onClick: () => emit('selectAppForVersions', row) }, () => '历史版本'),
        h(AppButton, { size: 'tiny', type: 'primary', onClick: () => emit('downloadApp', row) }, () => '下载')
      ])
    }
  }
]

const displayedPurchasedColumns = computed(() => {
  if (props.effectiveTheme !== 'handdrawn') return purchasedColumns

  const widths = [227, 200, 104, 153, 156]
  return purchasedColumns.map((column, index) => ({
    ...column,
    minWidth: undefined,
    width: widths[index]
  }))
})

defineExpose({
  purchasedApps,
  purchasedTotal,
  loadPurchases
})
</script>
