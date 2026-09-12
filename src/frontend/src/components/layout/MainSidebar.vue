<template>
  <aside class="app-sidebar">
    <!-- Brand Header -->
    <div class="sidebar-brand">
      <div class="brand-icon-box">
        <img
          class="brand-logo-image"
          :src="standardLogo"
          alt="AppleVault"
        />
      </div>
      <div class="brand-text">
        <div class="brand-title-row">
          <div class="brand-name">果仓助手</div>
          <div class="brand-tag">v{{ appVersion }}</div>
        </div>
        <div class="brand-sub">AppleVault</div>
      </div>
    </div>

    <!-- Navigation List -->
    <nav class="sidebar-nav">
      <div class="nav-section-title">
        <span>核心功能</span>
      </div>
      <button
        class="nav-item"
        :class="{ active: activeTab === 'search' }"
        @click="emit('update:activeTab', 'search')"
      >
        <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="search" /></span>
        <span class="nav-label">应用搜索</span>
      </button>

      <button
        class="nav-item"
        :class="{ active: activeTab === 'versions' }"
        @click="emit('update:activeTab', 'versions')"
      >
        <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="history" /></span>
        <span class="nav-label">历史版本</span>
      </button>

      <button
        class="nav-item"
        :class="{ active: activeTab === 'download' }"
        @click="emit('update:activeTab', 'download')"
      >
        <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="download" /></span>
        <span class="nav-label">下载中心</span>
        <span v-if="activeTaskCount > 0" class="nav-badge">{{ activeTaskCount }}</span>
      </button>

      <button
        class="nav-item"
        :class="{ active: activeTab === 'purchased' }"
        @click="emit('update:activeTab', 'purchased')"
      >
        <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="purchased" /></span>
        <span class="nav-label">已购应用</span>
      </button>

      <button
        class="nav-item"
        :class="{ active: activeTab === 'installer' }"
        @click="emit('update:activeTab', 'installer')"
      >
        <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="device" /></span>
        <span class="nav-label">设备直装</span>
      </button>

      <div class="nav-section-title mt-4">
        <span>偏好与设置</span>
      </div>
      <button
        class="nav-item"
        :class="{ active: activeTab === 'account' }"
        @click="emit('update:activeTab', 'account')"
      >
        <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="account" /></span>
        <span class="nav-label">账号中心</span>
        <span v-if="isLoggedIn && account.region" class="sidebar-region-tag">{{ account.region }}</span>
        <span class="account-dot" :class="isLoggedIn ? 'dot-online' : 'dot-offline'"></span>
      </button>

      <button
        class="nav-item"
        :class="{ active: activeTab === 'settings' }"
        @click="emit('update:activeTab', 'settings')"
      >
        <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="settings" /></span>
        <span class="nav-label">系统设置</span>
      </button>

      <button
        class="nav-item"
        :class="{ active: activeTab === 'about' }"
        @click="emit('update:activeTab', 'about')"
      >
        <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="about" /></span>
        <span class="nav-label">关于软件</span>
      </button>
    </nav>

    <div v-if="effectiveTheme === 'handdrawn'" class="handdrawn-sidebar-floating-leaf" aria-hidden="true">
      <img :src="handdrawnSidebarFloatingLeaf" alt="" />
    </div>

    <div v-if="effectiveTheme === 'handdrawn'" class="handdrawn-sidebar-art" aria-hidden="true">
      <img :src="handdrawnSidebarArt" alt="" />
    </div>

    <!-- Sidebar Footer (Account & Quick Controls) -->
    <div class="sidebar-footer">
      <!-- Account Quick Card -->
      <div class="sidebar-account-card" @click="emit('update:activeTab', 'account')">
        <div class="user-avatar" :class="{ 'avatar-logged': isLoggedIn }">
          {{ isLoggedIn ? (account.name ? account.name.charAt(0).toUpperCase() : '') : '?' }}
        </div>
        <div class="user-info-text">
          <div class="user-name text-ellipsis">{{ account.name || '未登录 Apple ID' }}</div>
          <div class="user-email text-ellipsis">{{ account.email || '点击前往登录' }}</div>
        </div>
      </div>

      <!-- Quick Switch Bar -->
      <div class="sidebar-actions-bar">
        <!-- Proxy Switch -->
        <div class="quick-tool" title="网络代理">
          <AppSwitch
            :model-value="enableProxy"
            size="small"
            @update:model-value="emit('proxyToggle', $event)"
          />
          <span class="quick-tool-label">代理</span>
        </div>

        <!-- Theme Switcher (靠右显示) -->
        <div class="theme-switcher-wrapper">
          <button
            v-if="showThemeHint"
            type="button"
            class="theme-first-use-hint"
            @click="openThemePickerFromHint"
          >
            点击这里可以切换主题
          </button>
          <AppPopover
            v-model="showThemePopover"
            placement="top-end"
          >
            <template #trigger>
              <button
                class="icon-action-btn theme-action-btn"
                :class="{ active: showThemePopover }"
                title="选择外观主题"
                @click="dismissThemeHint"
              >
                <!-- 调色板/主题 图标 -->
                <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="13.5" cy="6.5" r=".5" fill="currentColor"></circle>
                  <circle cx="17.5" cy="10.5" r=".5" fill="currentColor"></circle>
                  <circle cx="8.5" cy="7.5" r=".5" fill="currentColor"></circle>
                  <circle cx="6.5" cy="12.5" r=".5" fill="currentColor"></circle>
                  <path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.648-.746 1.648-1.688 0-.437-.18-.835-.437-1.125-.29-.289-.438-.652-.438-1.125a1.64 1.64 0 0 1 1.668-1.668h1.996c3.051 0 5.563-2.512 5.563-5.563C22 6.5 17.5 2 12 2z"></path>
                </svg>
              </button>
            </template>
            <div
              class="theme-popover-menu"
              :class="[
                `theme-${effectiveTheme}`,
                isDarkTheme ? 'is-dark' : 'is-light'
              ]"
            >
              <div class="theme-popover-header">
                <div class="theme-popover-title">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 5px;">
                    <circle cx="13.5" cy="6.5" r=".5" fill="currentColor"></circle>
                    <circle cx="17.5" cy="10.5" r=".5" fill="currentColor"></circle>
                    <circle cx="8.5" cy="7.5" r=".5" fill="currentColor"></circle>
                    <circle cx="6.5" cy="12.5" r=".5" fill="currentColor"></circle>
                    <path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.648-.746 1.648-1.688 0-.437-.18-.835-.437-1.125-.29-.289-.438-.652-.438-1.125a1.64 1.64 0 0 1 1.668-1.668h1.996c3.051 0 5.563-2.512 5.563-5.563C22 6.5 17.5 2 12 2z"></path>
                  </svg>
                  选择外观主题
                </div>
              </div>
              
              <div class="theme-options-list">
                <div
                  v-for="item in themeList"
                  :key="item.key"
                  class="theme-select-item"
                  :class="{ active: effectiveTheme === item.key }"
                  @click="selectTheme(item.key)"
                >
                  <div class="theme-color-preview" :style="{ background: item.preview.bg, borderColor: item.preview.border }">
                    <div class="theme-color-card" :style="{ background: item.preview.card, borderColor: item.preview.border }">
                      <span class="theme-color-dot" :style="{ background: item.preview.accent }"></span>
                    </div>
                  </div>
                  <div class="theme-info-box">
                    <div class="theme-item-name">
                      <span>{{ item.name }}</span>
                      <span v-if="item.recommended" class="theme-thumb-badge" title="精选推荐">
                        <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor" aria-hidden="true">
                          <path d="M1 21h4V9H1v12zm22-11c0-1.1-.9-2-2-2h-6.31l.95-4.57.03-.32c0-.41-.17-.79-.44-1.06L14.17 1 7.59 7.59C7.22 7.95 7 8.45 7 9v10c0 1.1.9 2 2 2h9c.83 0 1.54-.5 1.84-1.22l3.02-7.05c.09-.23.14-.47.14-.73v-2z"></path>
                        </svg>
                      </span>
                    </div>
                    <div class="theme-item-desc">{{ item.description }}</div>
                  </div>
                  <div v-if="effectiveTheme === item.key" class="theme-check-icon">
                    <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                      <polyline points="20 6 9 17 4 12"></polyline>
                    </svg>
                  </div>
                </div>
              </div>
            </div>
          </AppPopover>
        </div>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, h, defineComponent, onMounted, onBeforeUnmount } from 'vue'
import { backend as main } from '../../../wailsjs/go/models'
import AppPopover from '../../ui/components/AppPopover.vue'
import AppSwitch from '../../ui/components/AppSwitch.vue'
import { type AppTheme, themeOptions } from '../../ui/theme'
import standardLogo from '../../assets/applevault-logo.webp'
import handdrawnSidebarArt from '../../ui/themes/handdrawn/assets/sketch-sidebar-art.svg'
import handdrawnSidebarFloatingLeaf from '../../ui/themes/handdrawn/assets/sketch-floating-leaf.svg'

const sketchNavPaths: Record<string, string[]> = {
  search: ['M10.8 4.2a6.6 6.6 0 1 0 0 13.2 6.6 6.6 0 0 0 0-13.2Z', 'm15.8 15.8 4 4'],
  history: ['M12 4.1a7.9 7.9 0 1 1-5.8 2.5', 'M4.2 4.6v4h4', 'M12 7.7v4.8l3.2 2'],
  download: ['M12 3.5v11', 'm7.8 11.2 4.2 4.1 4.2-4.1', 'M4.5 18v2h15v-2'],
  purchased: ['M5.5 8h13l-.5 12H6L5.5 8Z', 'M8.2 8V6.5a3.8 3.8 0 0 1 7.6 0V8'],
  device: ['M7.2 3.2h9.6c1 0 1.7.8 1.7 1.8v14c0 1-.7 1.8-1.7 1.8H7.2c-1 0-1.7-.8-1.7-1.8V5c0-1 .7-1.8 1.7-1.8Z', 'M10 6h4', 'M11 18h2'],
  account: ['M12 4.2a3.8 3.8 0 1 0 0 7.6 3.8 3.8 0 0 0 0-7.6Z', 'M5.2 20c.7-4 3.2-6.1 6.8-6.1s6.1 2.1 6.8 6.1'],
  settings: ['M12 8.4a3.6 3.6 0 1 0 0 7.2 3.6 3.6 0 0 0 0-7.2Z', 'M9.8 3.5 9.2 5a7.6 7.6 0 0 0-1.8 1l-1.5-.6-1.6 2.8 1.2 1a7.8 7.8 0 0 0-.1 2.1l-1.4.9.8 3.1 1.7-.1a7.5 7.5 0 0 0 1.5 1.4l-.2 1.7 3.1.9.9-1.4a7.3 7.3 0 0 0 2-.2l1.1 1.3 2.8-1.6-.6-1.5a7.6 7.6 0 0 0 1-1.8l1.6-.5-.1-3.2-1.6-.4a7.7 7.7 0 0 0-1.1-1.8l.5-1.6-2.8-1.5-1 1.3a7.7 7.7 0 0 0-2-.1l-.9-1.4Z'],
  about: ['M12 3.8a8.2 8.2 0 1 0 0 16.4 8.2 8.2 0 0 0 0-16.4Z', 'M12 10.5v5.8', 'M12 7.3h.01'],
}

const SketchNavIcon = defineComponent({
  name: 'SketchNavIcon',
  props: { name: { type: String, required: true } },
  setup(props) {
    return () => {
      const paths = sketchNavPaths[props.name] || sketchNavPaths.about
      const makePaths = (opacity: number, transform?: string) =>
        h('g', { opacity, transform }, paths.map((d) => h('path', { d })))
      return h('svg', {
        class: 'nav-icon-image sketch-nav-icon',
        viewBox: '0 0 24 24',
        fill: 'none',
        stroke: 'currentColor',
        'stroke-width': '1.75',
        'stroke-linecap': 'round',
        'stroke-linejoin': 'round',
      }, [makePaths(1), makePaths(0.28, 'translate(.35 .25)')])
    }
  },
})

const props = defineProps<{
  activeTab: string
  effectiveTheme: AppTheme
  isDarkTheme: boolean
  account: main.AccountInfo
  isLoggedIn: boolean
  enableProxy: boolean
  activeTaskCount: number
  isDemoMode?: boolean
  appVersion: string
}>()

const emit = defineEmits<{
  (e: 'update:activeTab', tab: string): void
  (e: 'update:currentTheme', theme: AppTheme): void
  (e: 'proxyToggle', val: boolean): void
}>()

const showThemePopover = ref(false)
const showThemeHint = ref(false)
let themeHintTimer: ReturnType<typeof setTimeout> | null = null
const themeList = themeOptions

function dismissThemeHint() {
  showThemeHint.value = false
  if (themeHintTimer) {
    clearTimeout(themeHintTimer)
    themeHintTimer = null
  }
}

function openThemePickerFromHint() {
  dismissThemeHint()
  showThemePopover.value = true
}

function selectTheme(themeKey: AppTheme) {
  emit('update:currentTheme', themeKey)
  showThemePopover.value = false
}

onMounted(() => {
  if (props.isDemoMode || localStorage.getItem('apple_vault_theme_hint_seen_v1') === 'true') return
  localStorage.setItem('apple_vault_theme_hint_seen_v1', 'true')
  showThemeHint.value = true
  themeHintTimer = setTimeout(dismissThemeHint, 5000)
})

onBeforeUnmount(dismissThemeHint)
</script>
