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
            @keydown.enter="handleListVersions"
          />
          <AppButton
            type="primary"
            size="small"
            :loading="isListingVersions"
            :disabled="isListingVersions || isBatchQuerying || isTargetQuerying"
            @click="handleListVersions"
          >
            获取历史版本
          </AppButton>
          <AppButton
            :type="isBatchQuerying ? 'error' : 'default'"
            :secondary="!isBatchQuerying"
            size="small"
            :loading="isBatchQuerying"
            :disabled="versionItems.length === 0 || isListingVersions || isTargetQuerying"
            @click="handleBatchQueryClick"
          >
            {{ isBatchQuerying ? '停止查询' : '批量解析前 30 项' }}
          </AppButton>
          <AppButton
            :type="isTargetQuerying ? 'error' : 'default'"
            :secondary="!isTargetQuerying"
            size="small"
            :loading="isTargetQuerying"
            :disabled="versionItems.length === 0 || isListingVersions || isBatchQuerying"
            @click="handleTargetQueryClick"
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
        :columns="displayedVersionColumns"
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

    <!-- 查找指定版本弹窗 -->
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
          @keydown.enter="confirmStartTargetQuery"
        />
        <div class="dialog-action-buttons mt-4">
          <AppButton secondary @click="showTargetVersionModal = false">取消</AppButton>
          <AppButton type="primary" :disabled="!targetVersionInput.trim()" @click="confirmStartTargetQuery">开始查询</AppButton>
        </div>
      </div>
    </AppDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, h } from 'vue'
import { backend as main } from '../../../wailsjs/go/models'
import { ListVersions, GetVersionMetadata } from '../../../wailsjs/go/backend/App'
import AppButton from '../../ui/components/AppButton.vue'
import AppCard from '../../ui/components/AppCard.vue'
import AppDialog from '../../ui/components/AppDialog.vue'
import AppInput from '../../ui/components/AppInput.vue'
import AppInputGroup from '../../ui/components/AppInputGroup.vue'
import AppTable from '../../ui/components/AppTable.vue'
import { useAppDialog, useAppMessage } from '../../ui/feedback'
import type { AppTheme } from '../../ui/theme'

export interface VersionItem {
  versionId: string
  displayVersion: string
  fileSize: string
  releaseDate: string
  isQuerying?: boolean
}

const props = defineProps<{
  effectiveTheme?: AppTheme
}>()

const emit = defineEmits<{
  (e: 'downloadVersion', row: VersionItem, form: { appName: string; bundleId: string; appId: number }): void
  (e: 'busyChange', busy: boolean, text?: string): void
}>()

const message = useAppMessage()
const dialog = useAppDialog()

const isListingVersions = ref(false)
const isBatchQuerying = ref(false)
const shouldStopBatchQuery = ref(false)
const isTargetQuerying = ref(false)
const shouldStopTargetQuery = ref(false)
const showTargetVersionModal = ref(false)
const targetVersionInput = ref('')

const versionForm = ref({
  bundleId: '',
  appId: 0,
  appName: '',
  filter: ''
})

const versionItems = ref<VersionItem[]>([])

const filteredVersions = computed(() => {
  const f = versionForm.value.filter.trim().toLowerCase()
  if (!f) return versionItems.value
  return versionItems.value.filter(item =>
    item.versionId.toLowerCase().includes(f) ||
    item.displayVersion.toLowerCase().includes(f) ||
    item.fileSize.toLowerCase().includes(f) ||
    item.releaseDate.toLowerCase().includes(f)
  )
})

function parseVersionParts(v: string): number[] {
  if (!v || v === '未查询') return []
  const clean = v.trim().toLowerCase().replace(/^v/, '')
  const match = clean.match(/^(\d+(?:\.\d+)*)/)
  if (!match) return []
  return match[1].split('.').map(num => parseInt(num, 10))
}

function compareVersions(v1: string, v2: string): number {
  const p1 = parseVersionParts(v1)
  const p2 = parseVersionParts(v2)
  if (p1.length === 0 || p2.length === 0) {
    const clean1 = v1.trim().toLowerCase().replace(/^v/, '')
    const clean2 = v2.trim().toLowerCase().replace(/^v/, '')
    if (clean1 === clean2) return 0
    return clean1 > clean2 ? 1 : -1
  }
  const maxLen = Math.max(p1.length, p2.length)
  for (let i = 0; i < maxLen; i++) {
    const num1 = p1[i] || 0
    const num2 = p2[i] || 0
    if (num1 > num2) return 1
    if (num1 < num2) return -1
  }
  return 0
}

function isVersionMatch(actual: string, target: string): boolean {
  if (!actual || actual === '未查询') return false
  const a = actual.trim().toLowerCase().replace(/^v/, '')
  const t = target.trim().toLowerCase().replace(/^v/, '')
  if (a === t) return true
  return compareVersions(actual, target) === 0
}

async function cancellableSleep(ms: number, stopRef: { value: boolean }): Promise<boolean> {
  const start = Date.now()
  while (Date.now() - start < ms) {
    if (stopRef.value) return false
    await new Promise(r => setTimeout(r, 20))
  }
  return !stopRef.value
}

/**
 * 查询单项构建版本的详细版本号与文件体积
 */
async function querySingleVersionMetadata(row: VersionItem) {
  row.isQuerying = true
  emit('busyChange', true, `正在查询版本 ID ${row.versionId} 的详情...`)

  try {
    const res = await GetVersionMetadata(versionForm.value.bundleId, row.versionId, versionForm.value.appId)
    row.displayVersion = res.displayVersion
    row.fileSize = res.displayFileSize
    row.releaseDate = res.releaseDate ? res.releaseDate.replace(/T.*/, '') : '-'
    versionItems.value = [...versionItems.value]
    emit('busyChange', false, `构建 ID ${row.versionId} 对应版本: ${res.displayVersion} (${res.displayFileSize})`)
  } catch (err: any) {
    message.error(`查询详情失败: ${err}`)
  } finally {
    row.isQuerying = false
    emit('busyChange', false)
  }
}

/**
 * 枚举所有历史版本构建 ID
 */
async function handleListVersions() {
  if (!versionForm.value.bundleId) {
    message.warning('请输入 Bundle ID')
    return
  }

  isListingVersions.value = true
  emit('busyChange', true, `正在查询 ${versionForm.value.bundleId} 的历史版本列表...`)

  try {
    const res = await ListVersions(versionForm.value.bundleId, versionForm.value.appId)
    versionItems.value = (res.externalVersionIdentifiers || []).map(id => ({
      versionId: id,
      displayVersion: '未查询',
      fileSize: '-',
      releaseDate: '-'
    }))
    emit('busyChange', false, `共获取到 ${versionItems.value.length} 个历史版本记录。`)
  } catch (err: any) {
    message.error(`查询历史版本失败: ${err}`)
    emit('busyChange', false, '查询历史版本失败')
  } finally {
    isListingVersions.value = false
  }
}

async function fetchItemMetadata(index: number): Promise<string> {
  if (index < 0 || index >= versionItems.value.length) return ''
  const item = versionItems.value[index]
  if (item.displayVersion && item.displayVersion !== '未查询') {
    return item.displayVersion
  }
  item.isQuerying = true
  try {
    const res = await GetVersionMetadata(versionForm.value.bundleId, item.versionId, versionForm.value.appId)
    item.displayVersion = res.displayVersion
    item.fileSize = res.displayFileSize
    item.releaseDate = res.releaseDate ? res.releaseDate.replace(/T.*/, '') : '-'
    versionItems.value = [...versionItems.value]
    return item.displayVersion
  } catch (err) {
    console.warn(`查询构建 ID ${item.versionId} 失败:`, err)
    return ''
  } finally {
    item.isQuerying = false
  }
}

/**
 * 批量解析前 N 个版本的真实版本号
 */
async function runBatchQuery(count: number) {
  isBatchQuerying.value = true
  shouldStopBatchQuery.value = false
  emit('busyChange', true, `正在批量查询前 ${count} 个历史版本号...`)

  try {
    for (let i = 0; i < count; i++) {
      if (shouldStopBatchQuery.value) {
        emit('busyChange', false, '批量查询已手动停止。')
        message.info('批量查询已停止')
        break
      }

      const item = versionItems.value[i]
      if (item.displayVersion !== '未查询') continue

      emit('busyChange', true, `正在查询 (${i + 1}/${count}): ${item.versionId}...`)
      await fetchItemMetadata(i)

      if (shouldStopBatchQuery.value) break
      if (!await cancellableSleep(120, shouldStopBatchQuery)) break
    }

    if (!shouldStopBatchQuery.value) {
      emit('busyChange', false, '批量查询完成。')
      message.success('批量查询完成')
    }
  } catch (err: any) {
    emit('busyChange', false, `批量查询失败: ${err}`)
    message.error(`批量查询失败: ${err}`)
  } finally {
    isBatchQuerying.value = false
    shouldStopBatchQuery.value = false
  }
}

function stopBatchQuery() {
  shouldStopBatchQuery.value = true
  emit('busyChange', true, '正在停止批量查询...')
}

function handleBatchQuery() {
  const count = Math.min(versionItems.value.length, 30)
  dialog.info({
    title: '批量解析确认',
    content: `即将批量解析前 ${count} 个版本的详细版本号与体积，是否继续？`,
    positiveText: '开始解析',
    negativeText: '取消',
    onPositiveClick: () => {
      void runBatchQuery(count)
    }
  })
}

function handleBatchQueryClick() {
  if (isBatchQuerying.value) {
    stopBatchQuery()
    return
  }
  handleBatchQuery()
}

function stopTargetQuery() {
  shouldStopTargetQuery.value = true
  emit('busyChange', true, '正在停止版本检索...')
}

function handleTargetQueryClick() {
  if (isTargetQuerying.value) {
    stopTargetQuery()
    return
  }
  if (versionItems.value.length === 0) {
    message.warning('请先点击「获取历史版本」')
    return
  }
  showTargetVersionModal.value = true
}

function confirmStartTargetQuery() {
  const target = targetVersionInput.value.trim()
  if (!target) {
    message.warning('请输入目标版本号')
    return
  }
  showTargetVersionModal.value = false
  void runTargetQuery(target)
}

/**
 * 通过二分法快速定位匹配指定版本号的构建 ID
 */
async function runTargetQuery(targetVersion: string) {
  isTargetQuerying.value = true
  shouldStopTargetQuery.value = false
  emit('busyChange', true, `正在检索目标版本: ${targetVersion}...`)

  const total = versionItems.value.length
  let found = false
  let foundItem: VersionItem | null = null

  try {
    let low = 0
    let high = total - 1

    emit('busyChange', true, `正在核对最新版本 (1/${total})...`)
    const v0 = await fetchItemMetadata(0)
    if (shouldStopTargetQuery.value) {
      emit('busyChange', false, `已停止查询指定版本 (${targetVersion})。`)
      return
    }

    if (v0 && isVersionMatch(v0, targetVersion)) {
      found = true
      foundItem = versionItems.value[0]
    } else if (v0 && compareVersions(v0, targetVersion) < 0) {
      emit('busyChange', false, `当前最新版本 (${v0}) 小于目标版本 ${targetVersion}，未找到该版本。`)
      message.warning(`当前最新版本为 ${v0}，未找到更高版本 ${targetVersion}`)
      return
    } else {
      low = 1
    }

    if (!found && high >= low) {
      emit('busyChange', true, `正在核对早期版本 (${total}/${total})...`)
      const vLast = await fetchItemMetadata(high)
      if (shouldStopTargetQuery.value) {
        emit('busyChange', false, `已停止查询指定版本 (${targetVersion})。`)
        return
      }

      if (vLast && isVersionMatch(vLast, targetVersion)) {
        found = true
        foundItem = versionItems.value[high]
      } else if (vLast && compareVersions(vLast, targetVersion) > 0) {
        emit('busyChange', false, `当前早期版本 (${vLast}) 大于目标版本 ${targetVersion}，未找到该版本。`)
        message.warning(`当前早期版本为 ${vLast}，未找到更低版本 ${targetVersion}`)
        return
      } else {
        high = high - 1
      }
    }

    if (!found) {
      while (low <= high && !shouldStopTargetQuery.value) {
        const mid = Math.floor((low + high) / 2)
        emit('busyChange', true, `正在核对版本 (构建 ID: ${versionItems.value[mid].versionId})...`)
        const vMid = await fetchItemMetadata(mid)

        if (shouldStopTargetQuery.value) break
        if (!await cancellableSleep(120, shouldStopTargetQuery)) break

        if (vMid && isVersionMatch(vMid, targetVersion)) {
          found = true
          foundItem = versionItems.value[mid]
          break
        }

        if (vMid) {
          if (compareVersions(vMid, targetVersion) < 0) {
            high = mid - 1
          } else {
            low = mid + 1
          }
        } else {
          low++
        }
      }
    }

    if (shouldStopTargetQuery.value) {
      emit('busyChange', false, `已停止查询指定版本 (${targetVersion})。`)
      message.info('已停止查询指定版本')
    } else if (found && foundItem) {
      emit('busyChange', false, `已找到目标版本 ${foundItem.displayVersion} (构建 ID: ${foundItem.versionId})！`)
      message.success(`已查询到指定版本 ${foundItem.displayVersion}！`)
      versionForm.value.filter = targetVersion
    } else {
      emit('busyChange', false, `检索完成，未在历史记录中找到指定版本: ${targetVersion}`)
      message.warning(`未找到版本号: ${targetVersion}`)
    }
  } catch (err: any) {
    emit('busyChange', false, `查询指定版本失败: ${err}`)
    message.error(`查询失败: ${err}`)
  } finally {
    isTargetQuerying.value = false
    shouldStopTargetQuery.value = false
  }
}

/**
 * 供搜索或已购页面选中某应用后跳转并开始枚举
 */
function selectAppForVersions(app: main.AppItem) {
  versionForm.value.appName = app.name
  versionForm.value.bundleId = app.bundleID
  versionForm.value.appId = app.id
  versionForm.value.filter = ''
  targetVersionInput.value = ''
  void handleListVersions()
}

/**
 * 终止所有异步查询
 */
function stopAllQueries() {
  shouldStopBatchQuery.value = true
  shouldStopTargetQuery.value = true
}

const versionColumns = [
  { title: '构建 ID', key: 'versionId', minWidth: 160 },
  {
    title: '版本号',
    key: 'displayVersion',
    width: 120,
    render(row: VersionItem) {
      if (row.displayVersion === '未查询') {
        return h('span', { class: 'text-dim italic' }, '未查询')
      }
      return h('span', { class: 'tag-version-highlight' }, row.displayVersion)
    }
  },
  {
    title: '文件体积',
    key: 'fileSize',
    width: 110,
    render(row: VersionItem) {
      return h('span', row.fileSize && row.fileSize !== '-' ? row.fileSize : '-')
    }
  },
  {
    title: '发布日期',
    key: 'releaseDate',
    width: 120,
    render(row: VersionItem) {
      return h('span', row.releaseDate && row.releaseDate !== '-' ? row.releaseDate : '-')
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 160,
    fixed: 'right' as const,
    render(row: VersionItem) {
      return h('div', { class: 'app-button-group' }, [
        h(AppButton, {
          size: 'tiny',
          secondary: true,
          disabled: isListingVersions.value || isBatchQuerying.value || isTargetQuerying.value,
          loading: row.isQuerying,
          onClick: () => querySingleVersionMetadata(row)
        }, () => '查详情'),
        h(AppButton, {
          size: 'tiny',
          type: 'primary',
          onClick: () => emit('downloadVersion', row, versionForm.value)
        }, () => '下载')
      ])
    }
  }
]

const displayedVersionColumns = computed(() => {
  if (props.effectiveTheme !== 'handdrawn') return versionColumns

  const widths = [314, 139, 101, 129, 155]
  return versionColumns.map((column, index) => ({
    ...column,
    minWidth: undefined,
    width: widths[index]
  }))
})

defineExpose({
  versionForm,
  versionItems,
  selectAppForVersions,
  stopAllQueries
})
</script>
