import { ref, computed, h, type Ref } from 'vue'
import { backend as main } from '../../wailsjs/go/models'
import { Search, Purchase } from '../../wailsjs/go/backend/App'
import AppButton from '../ui/components/AppButton.vue'
import { useAppMessage } from '../ui/feedback'
import type { AppTheme } from '../ui/theme'

export function useSearch(options: {
  effectiveTheme: Ref<AppTheme>
  onSelectAppForVersions: (app: main.AppItem) => void
  onDownloadApp: (app: main.AppItem) => void
  onBusyChange?: (busy: boolean, text?: string) => void
}) {
  const message = useAppMessage()
  const isSearching = ref(false)

  const searchForm = ref({
    term: '',
    limit: 10,
    platform: 'iphone'
  })

  const searchResults = ref<main.AppItem[]>([])

  async function handlePurchaseApp(bundleId: string) {
    options.onBusyChange?.(true, `正在获取应用授权: ${bundleId}...`)
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
      options.onBusyChange?.(false)
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
          h(AppButton, { size: 'tiny', secondary: true, onClick: () => options.onSelectAppForVersions(row) }, () => '历史版本'),
          h(AppButton, { size: 'tiny', type: 'primary', onClick: () => options.onDownloadApp(row) }, () => '下载'),
          h(AppButton, { size: 'tiny', secondary: true, onClick: () => handlePurchaseApp(row.bundleID) }, () => '获取许可')
        ])
      }
    }
  ]

  const displayedSearchColumns = computed(() => {
    if (options.effectiveTheme.value !== 'handdrawn') return searchColumns

    const widths = [196, 163, 94, 87, 83, 217]
    return searchColumns.map((column, index) => ({
      ...column,
      minWidth: undefined,
      width: widths[index]
    }))
  })

  async function handleSearch() {
    if (!searchForm.value.term) {
      message.warning('请输入搜索关键词')
      return
    }

    isSearching.value = true
    options.onBusyChange?.(true, `正在搜索 "${searchForm.value.term}"...`)

    try {
      const res = await Search(searchForm.value.term, searchForm.value.limit, searchForm.value.platform)
      searchResults.value = res.apps || []
      options.onBusyChange?.(false, `搜索完成，找到 ${searchResults.value.length} 个应用。`)
    } catch (err: any) {
      message.error(`搜索失败: ${err}`)
      options.onBusyChange?.(false, '搜索失败')
    } finally {
      isSearching.value = false
      options.onBusyChange?.(false)
    }
  }

  return {
    isSearching,
    searchForm,
    searchResults,
    displayedSearchColumns,
    handleSearch,
    handlePurchaseApp
  }
}
