<template>
  <div class="view-panel">
    <div class="view-header">
      <div>
        <div class="title-with-badge"><h2 class="view-title">已购应用</h2></div>
        <p class="view-desc">浏览与检索当前 Apple ID 名下已获得正版许可的历史应用库</p>
      </div>
      <div v-if="isLoading" class="purchased-loading-progress" role="status" aria-live="polite">
        <AppProgress v-if="loadTotal > 0" class="purchased-loading-bar" :percentage="loadProgress" :height="6" />
        <div v-else class="purchased-loading-bar purchased-loading-bar-indeterminate" aria-hidden="true"></div>
        <span v-if="loadTotal > 0" class="purchased-loading-count">{{ loadLoaded }}/{{ loadTotal }}</span>
        <span v-else class="purchased-loading-count loading-total-text">读取总数…</span>
      </div>
    </div>

    <AppCard class="clean-card mb-4 search-bar-card">
      <div class="search-input-group">
        <div class="search-input-wrapper">
          <AppSearchInput
            v-model="searchKeyword"
            size="small"
            placeholder="搜索已购应用名称、Bundle ID 或 App ID..."
            @keydown.esc="searchKeyword = ''"
          />
        </div>
        <div class="filter-controls">
          <AppButton type="primary" size="small" :loading="isLoading" @click="emit('refresh')">刷新列表</AppButton>
        </div>
      </div>
    </AppCard>

    <AppCard class="clean-card table-flex-card">
      <AppTable class="purchased-table" :columns="columns" :data="apps" :loading="isLoading" :scroll-x="720" size="small" flex-height style="height: 100%;">
        <template #empty>
          <div class="purchased-empty-box">
            <div v-if="searchKeyword" class="search-none-state">
              <div class="search-none-icon">🔍</div>
              <div class="search-none-title">未找到与「{{ searchKeyword }}」匹配的已购应用</div>
              <div class="search-none-desc">请尝试输入不同关键词，或点击下方按钮清空搜索</div>
              <AppButton size="small" secondary class="mt-2" @click="searchKeyword = ''">清空搜索</AppButton>
            </div>
            <div v-else class="search-none-state">
              <div class="search-none-title">暂无已购应用记录</div>
              <div class="search-none-desc">点击右上角「刷新列表」获取当前 Apple ID 历史正版应用</div>
            </div>
          </div>
        </template>
      </AppTable>
      <div class="purchased-bottom-pagination" aria-label="已购应用分页">
        <div class="purchased-search-badge purchased-footer-count" :class="{ 'has-filter': !!searchKeyword }">
          <span v-if="searchKeyword">找到 <b>{{ matchTotal }}</b> 款匹配应用（共 <b>{{ totalApps }}</b> 款）</span>
          <span v-else>共 <b>{{ total }}</b> 款已购应用</span>
        </div>
        <div class="page-size-selector">
          <span class="size-label">每页</span>
          <AppSelect
            :model-value="pageSize"
            size="small"
            style="width: 76px;"
            :options="pageSizeOptions"
            @update:model-value="emit('pageSizeChange', $event)"
          />
        </div>
        <AppButton secondary size="small" :disabled="page <= 1 || isLoading" @click="emit('previous')">上一页</AppButton>
        <span class="page-indicator">第 {{ page }} / {{ pageCount }} 页</span>
        <AppButton secondary size="small" :disabled="page >= pageCount || isLoading" @click="emit('next')">下一页</AppButton>
      </div>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import AppButton from '../../ui/components/AppButton.vue'
import AppCard from '../../ui/components/AppCard.vue'
import AppProgress from '../../ui/components/AppProgress.vue'
import AppSearchInput from '../../ui/components/AppSearchInput.vue'
import AppSelect, { type AppSelectOption, type SelectValue } from '../../ui/components/AppSelect.vue'
import AppTable from '../../ui/components/AppTable.vue'

defineProps<{
  isLoading: boolean
  loadTotal: number
  loadLoaded: number
  loadProgress: number
  columns: any[]
  apps: any[]
  matchTotal: number
  totalApps: number
  total: number
  page: number
  pageCount: number
  pageSize: number
  pageSizeOptions: AppSelectOption[]
}>()

const searchKeyword = defineModel<string>('searchKeyword', { required: true })
const emit = defineEmits<{
  refresh: []
  previous: []
  next: []
  pageSizeChange: [value: SelectValue]
}>()
</script>
