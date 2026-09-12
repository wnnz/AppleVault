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
            @keydown.enter="emit('search')"
          />
        </div>
        <div class="filter-controls">
          <div class="filter-item">
            <span class="filter-label">平台</span>
            <AppSelect v-model="searchForm.platform" :options="platformOptions" size="small" style="width: 130px;" />
          </div>
          <div class="filter-item">
            <span class="filter-label">条数</span>
            <AppSelect v-model="searchForm.limit" :options="limitOptions" size="small" style="width: 90px;" />
          </div>
          <AppButton variant="primary" size="small" :loading="isSearching" @click="emit('search')">搜索</AppButton>
        </div>
      </div>
    </AppCard>

    <AppCard class="clean-card table-flex-card">
      <AppTable
        :columns="columns"
        :data="results"
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
import AppButton from '../../ui/components/AppButton.vue'
import AppCard from '../../ui/components/AppCard.vue'
import AppInput from '../../ui/components/AppInput.vue'
import AppSelect, { type AppSelectOption } from '../../ui/components/AppSelect.vue'
import AppTable from '../../ui/components/AppTable.vue'

defineProps<{
  searchForm: { term: string; limit: number; platform: string }
  platformOptions: AppSelectOption[]
  limitOptions: AppSelectOption[]
  isSearching: boolean
  columns: any[]
  results: any[]
}>()

const emit = defineEmits<{ search: [] }>()
</script>
