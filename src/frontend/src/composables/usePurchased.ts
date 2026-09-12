import { ref, computed, watch, h, type Ref } from 'vue'
import { backend as main } from '../../wailsjs/go/models'
import { ListPurchases } from '../../wailsjs/go/backend/App'
import AppButton from '../ui/components/AppButton.vue'
import { useAppMessage } from '../ui/feedback'
import type { AppTheme } from '../ui/theme'

export function formatReadableDate(raw?: string): string {
  if (!raw) return '-'
  return raw.replace(/T/, ' ').replace(/:\d{2}Z$/, '').replace(/Z$/, '')
}

export const purchasedPageSizeOptions = [
  { label: '20', value: 20 },
  { label: '50', value: 50 },
  { label: '100', value: 100 }
]

export function usePurchased(options: {
  effectiveTheme: Ref<AppTheme>
  onSelectAppForVersions: (app: main.AppItem) => void
  onDownloadApp: (app: main.AppItem) => void
  onBusyChange?: (busy: boolean, text?: string) => void
}) {
  const message = useAppMessage()

  const isPurchasedLoading = ref(false)
  const purchasedLoadLoaded = ref(0)
  const purchasedLoadTotal = ref(0)
  const purchasedPage = ref(1)
  const purchasedPageSize = ref(20)
  const purchasedTotal = ref(0)
  const purchasedApps = ref<main.AppItem[]>([])
  const purchasedSearchKeyword = ref('')

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
          h(AppButton, { size: 'tiny', secondary: true, onClick: () => options.onSelectAppForVersions(row) }, () => '历史版本'),
          h(AppButton, { size: 'tiny', type: 'primary', onClick: () => options.onDownloadApp(row) }, () => '下载')
        ])
      }
    }
  ]

  const displayedPurchasedColumns = computed(() => {
    if (options.effectiveTheme.value !== 'handdrawn') return purchasedColumns

    const widths = [227, 200, 104, 153, 156]
    return purchasedColumns.map((column, index) => ({
      ...column,
      minWidth: undefined,
      width: widths[index]
    }))
  })

  async function loadPurchases() {
    isPurchasedLoading.value = true
    options.onBusyChange?.(true, '正在获取已购应用总数...')
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
          options.onBusyChange?.(true, `正在加载已购应用 (${purchasedLoadLoaded.value}/${firstTotal})...`)
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
      options.onBusyChange?.(false, `已购应用加载完成 (共 ${purchasedApps.value.length} 款)。`)
    } catch (err: any) {
      message.error(`加载已购列表失败: ${err}`)
      options.onBusyChange?.(false, '加载失败')
    } finally {
      isPurchasedLoading.value = false
      options.onBusyChange?.(false)
    }
  }

  function onPurchasedPageSizeChange(val: string | number | null) {
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

  return {
    isPurchasedLoading,
    purchasedLoadLoaded,
    purchasedLoadTotal,
    purchasedLoadProgress,
    purchasedPage,
    purchasedPageSize,
    purchasedTotal,
    purchasedApps,
    purchasedSearchKeyword,
    filteredPurchasedApps,
    purchasedMatchTotal,
    purchasedPageCount,
    displayedPurchasedColumns,
    purchasedPageSizeOptions,
    loadPurchases,
    onPurchasedPageSizeChange,
    prevPurchasedPage,
    nextPurchasedPage
  }
}
