<template>
  <div
    class="app-layout"
    :class="[
      `theme-${effectiveTheme}`,
      isDarkTheme ? 'dark-mode' : 'light-mode',
      isHanddrawn ? 'handdrawn-mode' : 'minimal-mode'
    ]"
  >
    <!-- Left Modern Sidebar -->
    <aside class="app-sidebar">
      <!-- Brand Header -->
      <div class="sidebar-brand">
        <div class="brand-icon-box">
          <svg class="brand-apple-svg" viewBox="0 0 170 170" fill="currentColor">
            <path d="M150.37 130.25c-2.45 5.66-5.35 10.87-8.71 15.66-4.58 6.53-8.33 11.05-11.22 13.56-4.48 4.12-9.28 6.23-14.42 6.35-3.69 0-8.14-1.05-13.32-3.18-5.19-2.12-9.97-3.17-14.34-3.17-4.58 0-9.49 1.05-14.75 3.17-5.26 2.13-9.5 3.24-12.74 3.35-4.35.13-9.16-1.9-14.42-6.08-3.7-3.08-7.7-7.85-12.01-14.3-6.24-9.35-11.12-20.2-14.65-32.54-3.52-12.35-5.29-24.3-5.29-35.87 0-14.12 3.52-25.75 10.57-34.89 7.05-9.14 16.03-13.88 26.94-14.21 4.79 0 10.36 1.34 16.71 4.02 6.36 2.68 10.15 4.08 11.37 4.19 1.12-.11 5.02-1.57 11.7-4.38 6.68-2.82 12.35-4.08 17.02-3.78 12.79.89 23.01 5.66 30.65 14.31-11.29 6.81-16.79 16.32-16.5 28.53.33 9.61 4.2 17.58 11.62 23.9 7.42 6.32 16.31 9.94 26.68 10.86-2.12 6.54-4.53 13.06-7.24 19.56zm-29.35-104.9c-.11 4.14-1.55 8.35-4.32 12.63-2.77 4.28-6.42 7.74-10.96 10.38-3.02 1.63-6.21 2.72-9.56 3.27-.11-1.3-.11-2.4-.11-3.27 0-4.13 1.54-8.38 4.63-12.75 3.09-4.37 7.02-7.86 11.8-10.47 2.91-1.63 5.75-2.73 8.52-3.3 0 1.2.06 2.37 0 3.51z"/>
          </svg>
        </div>
        <div class="brand-text">
          <div class="brand-title-row">
            <div class="brand-name">果仓助手</div>
            <span v-if="isHanddrawn" class="brand-handdrawn-crown" aria-hidden="true">👑</span>
            <div class="brand-tag">v1.0</div>
          </div>
          <div class="brand-sub">AppleVault</div>
        </div>
      </div>

      <!-- Navigation List -->
      <nav class="sidebar-nav">
        <div class="nav-section-title">
          <span>核心功能</span>
          <span v-if="isHanddrawn" class="nav-title-sparkle" aria-hidden="true">✦</span>
        </div>
        <button
          class="nav-item"
          :class="{ active: activeTab === 'search' }"
          @click="activeTab = 'search'"
        >
          <span v-if="isHanddrawn" class="nav-doodle-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="10.5" cy="10.5" r="6" />
              <path d="M15 15l5.5 5.5" stroke-width="2.4" />
              <path d="M8 8.5a3 3 0 0 1 3-3" stroke-width="1.3" opacity="0.8" />
            </svg>
          </span>
          <span class="nav-label">应用搜索</span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'versions' }"
          @click="activeTab = 'versions'"
        >
          <span v-if="isHanddrawn" class="nav-doodle-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="13" r="7.5" />
              <polyline points="12 9 12 13 15 14.5" stroke-width="2.2" />
              <path d="M9 3.5h6" stroke-width="2" />
              <path d="M12 3.5v2" stroke-width="2" />
            </svg>
          </span>
          <span class="nav-label">历史版本</span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'download' }"
          @click="activeTab = 'download'"
        >
          <span v-if="isHanddrawn" class="nav-doodle-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M5 16.5c-2.2 0-3.8-1.7-3.8-3.8 0-1.8 1.4-3.4 3.2-3.7.6-3 3.2-5 6.4-5 3.3 0 6 2.2 6.5 5.2 2 .4 3.5 2 3.5 4.1 0 2.3-1.9 4.2-4.2 4.2H5z" />
              <path d="M12 11.5v6.5" stroke-width="2.2" />
              <path d="M9.5 15.5l2.5 2.5 2.5-2.5" stroke-width="2.2" />
            </svg>
          </span>
          <span class="nav-label">下载中心</span>
          <span v-if="activeTaskCount > 0" class="nav-badge">{{ activeTaskCount }}</span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'purchased' }"
          @click="activeTab = 'purchased'"
        >
          <span v-if="isHanddrawn" class="nav-doodle-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M6 8h12l1.5 11.5a1.5 1.5 0 0 1-1.5 1.5H6a1.5 1.5 0 0 1-1.5-1.5L6 8z" />
              <path d="M9 10V6a3 3 0 0 1 6 0v4" stroke-width="2" />
              <path d="M12 12.5v2.5" stroke-width="1.8" />
            </svg>
          </span>
          <span class="nav-label">已购应用</span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'installer' }"
          @click="activeTab = 'installer'"
        >
          <span v-if="isHanddrawn" class="nav-doodle-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="6.5" y="3.5" width="11" height="17" rx="3" />
              <circle cx="12" cy="17.5" r="1" fill="currentColor" />
              <line x1="10" y1="6.5" x2="14" y2="6.5" stroke-width="1.6" />
            </svg>
          </span>
          <span class="nav-label">设备直装</span>
        </button>

        <div class="nav-section-title mt-4">
          <span>偏好与设置</span>
          <span v-if="isHanddrawn" class="nav-title-sparkle" aria-hidden="true">✦</span>
        </div>
        <button
          class="nav-item"
          :class="{ active: activeTab === 'account' }"
          @click="activeTab = 'account'"
        >
          <span v-if="isHanddrawn" class="nav-doodle-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="8.5" r="4.2" />
              <path d="M4.5 19.5c0-3.3 3.3-5.5 7.5-5.5s7.5 2.2 7.5 5.5" />
            </svg>
          </span>
          <span class="nav-label">账号中心</span>
          <span class="account-dot" :class="isLoggedIn ? 'dot-online' : 'dot-offline'"></span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'settings' }"
          @click="activeTab = 'settings'"
        >
          <span v-if="isHanddrawn" class="nav-doodle-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="3" />
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
            </svg>
          </span>
          <span class="nav-label">系统设置</span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'about' }"
          @click="activeTab = 'about'"
        >
          <span v-if="isHanddrawn" class="nav-doodle-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="8.5" />
              <line x1="12" y1="11" x2="12" y2="16.5" stroke-width="2.2" />
              <circle cx="12" cy="7.5" r="1.2" fill="currentColor" />
            </svg>
          </span>
          <span class="nav-label">关于软件</span>
        </button>
      </nav>

      <!-- Sidebar Footer (Account & Quick Controls) -->
      <div class="sidebar-footer">
        <!-- 手绘模式小插图 (纯手绘插画贴纸，取代原文字框) -->
        <div v-if="isHanddrawn" class="handdrawn-sidebar-sticker" aria-hidden="true">
          <svg viewBox="0 0 140 38" width="140" height="38" fill="none" class="sidebar-sticker-svg">
            <ellipse cx="68" cy="24" rx="55" ry="9" class="doodle-ground-wash" />
            <g transform="translate(42, 6)">
              <path d="M12 7c-4-4-10-1-10 4.5 0 6.5 5 11.5 10 13.5 5-2 10-7 10-13.5C22 6 16 3 12 7z" class="doodle-apple-body" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round" />
              <path d="M12 7c-.5-3.5 1-5.5 3-7" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
              <path d="M13 3c2-2.5 5-2 6-.5.5 2-1 3.5-3.5 3.5-1.5 0-2.2-.5-2.5-3z" class="doodle-leaf" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round" />
              <circle cx="8" cy="12" r="1.2" fill="#ffffff" opacity="0.8" />
            </g>
            <path d="M82 14l1.2 2.8 2.8 1.2-2.8 1.2-1.2 2.8-1.2-2.8-2.8-1.2 2.8-1.2 1.2-2.8z" class="doodle-star" />
            <path d="M26 18l1 2 2 1-2 1-1 2-1-2-2-1 2-1 1-2z" class="doodle-star doodle-star-small" />
            <circle cx="98" cy="22" r="1.4" class="doodle-sparkle" />
            <circle cx="18" cy="12" r="1.2" class="doodle-sparkle" />
          </svg>
        </div>

        <!-- Account Quick Card -->
        <div class="sidebar-account-card" @click="activeTab = 'account'">
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
            <n-switch
              v-model:value="settings.enableProxy"
              size="small"
              @update:value="onProxyToggle"
            />
            <span class="quick-tool-label">代理</span>
          </div>

          <!-- Theme Switcher (靠右显示) -->
          <div class="theme-switcher-wrapper">
            <n-popover
              v-model:show="showThemePopover"
              trigger="click"
              placement="top-end"
              :show-arrow="false"
              raw
            >
              <template #trigger>
                <button
                  class="icon-action-btn theme-action-btn"
                  :class="{ active: showThemePopover }"
                  title="选择外观主题"
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
                  isDarkTheme ? 'is-dark' : 'is-light',
                  isHanddrawn ? 'is-handdrawn' : 'is-minimal'
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
                
                <!-- 简约风格 -->
                <div class="theme-group-label">简约风格</div>
                <div class="theme-options-list">
                  <div
                    v-for="item in themeList.filter(t => t.category === 'minimal')"
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
                      <div class="theme-item-name">{{ item.name }}</div>
                      <div class="theme-item-desc">{{ item.description }}</div>
                    </div>
                    <div v-if="effectiveTheme === item.key" class="theme-check-icon">
                      <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                        <polyline points="20 6 9 17 4 12"></polyline>
                      </svg>
                    </div>
                  </div>
                </div>

                <!-- 手绘水彩风格 -->
                <div class="theme-group-label" style="margin-top: 8px;">手绘风格</div>
                <div class="theme-options-list">
                  <div
                    v-for="item in themeList.filter(t => t.category === 'handdrawn')"
                    :key="item.key"
                    class="theme-select-item"
                    :class="{ active: effectiveTheme === item.key }"
                    @click="selectTheme(item.key)"
                  >
                    <div class="theme-color-preview handdrawn-preview" :style="{ background: item.preview.bg, borderColor: item.preview.border }">
                      <div class="theme-color-card" :style="{ background: item.preview.card, borderColor: item.preview.border }">
                        <span class="theme-color-dot" :style="{ background: item.preview.accent }"></span>
                      </div>
                    </div>
                    <div class="theme-info-box">
                      <div class="theme-item-name">{{ item.name }}</div>
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
            </n-popover>
          </div>
        </div>
      </div>
    </aside>

    <!-- Right Main Workspace -->
    <section class="app-main">
      <!-- 手绘模式右上角小插图 (可爱水彩云朵与星星) -->
      <div v-if="isHanddrawn" class="handdrawn-top-doodle" aria-hidden="true">
        <svg viewBox="0 0 110 46" width="110" height="46" fill="none" class="floating-doodle-svg">
          <path
            d="M24 35c-5 0-9-3.8-9-8.5 0-4.2 3.2-7.8 7.5-8.3 1.3-6.2 7-10.7 13.8-10.7 5.8 0 10.8 3.5 13 8.5 2.2-1.3 5-1.8 7.8-1 4.2 1.3 7 5 7.5 9.2 4.2.5 7.4 4 7.4 8.3 0 4.7-4 8.5-9 8.5H24z"
            class="doodle-cloud-body"
            stroke="currentColor"
            stroke-width="1.6"
            stroke-linejoin="round"
          />
          <path d="M29 24c1 1.2 2.5 1.8 4 1.8s3-.6 4-1.8" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" />
          <circle cx="27" cy="22" r="1" fill="currentColor" />
          <circle cx="39" cy="22" r="1" fill="currentColor" />
          <ellipse cx="25" cy="25" rx="2" ry="1.2" class="doodle-blush" />
          <ellipse cx="41" cy="25" rx="2" ry="1.2" class="doodle-blush" />
          <path d="M78 13l1.5 3.5 3.5 1.5-3.5 1.5-1.5 3.5-1.5-3.5-3.5-1.5 3.5-1.5 1.5-3.5z" class="doodle-star" />
          <path d="M96 23l1 2.2 2.2 1-2.2 1-1 2.2-1-2.2-2.2-1 2.2-1 1-2.2z" class="doodle-star doodle-star-small" />
          <circle cx="68" cy="9" r="1.2" class="doodle-sparkle" />
          <circle cx="88" cy="33" r="1" class="doodle-sparkle" />
        </svg>
      </div>

      <!-- Dynamic View Container -->
      <div class="view-content-wrapper">

        <!-- VIEW 1: 应用搜索 (search) -->
        <div v-show="activeTab === 'search'" class="view-panel">
          <div class="view-header">
            <div>
              <div class="title-with-badge">
                <h2 class="view-title">应用搜索</h2>
                <span v-if="isHanddrawn" class="handdrawn-view-badge" aria-hidden="true"># 随心探索 🔍</span>
              </div>
              <p class="view-desc">在 Apple App Store 全球库中精准检索正版应用信息</p>
            </div>
          </div>

          <div class="clean-card mb-4 search-bar-card">
            <div class="search-input-group">
              <div class="search-input-wrapper">
                <input
                  v-model="searchForm.term"
                  class="clean-input search-input height-aligned"
                  placeholder="输入应用名称、关键字或开发商（如：微信、支付宝、TikTok）"
                  @keydown.enter="handleSearch"
                />
              </div>

              <div class="filter-controls">
                <div class="filter-item">
                  <span class="filter-label">平台</span>
                  <n-select
                    v-model:value="searchForm.platform"
                    :options="platformOptions"
                    size="medium"
                    style="width: 130px;"
                  />
                </div>

                <div class="filter-item">
                  <span class="filter-label">条数</span>
                  <n-select
                    v-model:value="searchForm.limit"
                    :options="limitOptions"
                    size="medium"
                    style="width: 90px;"
                  />
                </div>

                <n-button
                  type="primary"
                  size="medium"
                  :loading="isSearching"
                  @click="handleSearch"
                  class="height-aligned-btn"
                >
                  搜索
                </n-button>
              </div>
            </div>
          </div>

          <div class="clean-card table-flex-card">
            <n-data-table
              :columns="searchColumns"
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
                  <div v-if="isHanddrawn" class="handdrawn-table-empty-doodle" aria-hidden="true">
                    <svg viewBox="0 0 92 76" width="76" height="62" fill="none" class="empty-doodle-svg">
                      <rect x="14" y="16" width="64" height="46" rx="8" class="doodle-paper" stroke="currentColor" stroke-width="1.8" />
                      <line x1="24" y1="28" x2="52" y2="28" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
                      <line x1="24" y1="38" x2="44" y2="38" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
                      <circle cx="58" cy="40" r="11" class="doodle-ground-wash" stroke="currentColor" stroke-width="2" />
                      <line x1="66" y1="48" x2="77" y2="59" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" />
                      <path d="M22 10l1.2 2.5 2.5 1.2-2.5 1.2-1.2 2.5-1.2-2.5-2.5-1.2 2.5-1.2 1.2-2.5z" class="doodle-star" />
                    </svg>
                  </div>
                  <div class="empty-state-title">{{ searchForm.term ? '未找到匹配的应用' : '开启 App Store 探索之旅' }}</div>
                  <div class="empty-state-desc">{{ searchForm.term ? '请尝试更换关键词，或切换不同国家/地区搜索' : '输入应用名称、开发商或拼音，随时开始检索正版应用' }}</div>
                </div>
              </template>
            </n-data-table>
          </div>
        </div>

        <!-- VIEW 2: 历史版本 (versions) -->
        <div v-show="activeTab === 'versions'" class="view-panel">
          <div class="view-header">
            <div>
              <div class="title-with-badge">
                <h2 class="view-title">历史版本</h2>
                <span v-if="isHanddrawn" class="handdrawn-view-badge" aria-hidden="true"># 时光轨迹 ⏳</span>
              </div>
              <p class="view-desc">解析与枚举 App Store 所有历史构建版本，支持快速精准检索</p>
            </div>
            <div class="header-right-badges">
              <span class="count-pill">共 {{ filteredVersions.length }} / {{ versionItems.length }} 条记录</span>
            </div>
          </div>

          <div class="clean-card mb-4">
            <div class="versions-toolbar">
              <div class="bundle-input-row">
                <div class="clean-input-prefix-box height-aligned">
                  <span class="prefix-label">Bundle ID</span>
                  <input
                    v-model="versionForm.bundleId"
                    class="clean-input flex-1 border-none"
                    placeholder="例如: com.alipay.iphoneclient"
                    @keydown.enter="handleListVersions"
                  />
                </div>

                <n-button
                  type="primary"
                  size="medium"
                  class="height-aligned-btn"
                  :loading="isListingVersions"
                  :disabled="isListingVersions || isBatchQuerying || isTargetQuerying"
                  @click="handleListVersions"
                >
                  获取历史版本
                </n-button>

                <n-button
                  :type="isBatchQuerying ? 'error' : 'default'"
                  :secondary="!isBatchQuerying"
                  size="medium"
                  class="height-aligned-btn"
                  :loading="isBatchQuerying"
                  :disabled="versionItems.length === 0 || isListingVersions || isTargetQuerying"
                  @click="handleBatchQueryClick"
                >
                  {{ isBatchQuerying ? '停止查询' : '批量解析前 30 项' }}
                </n-button>

                <n-button
                  :type="isTargetQuerying ? 'error' : 'default'"
                  :secondary="!isTargetQuerying"
                  size="medium"
                  class="height-aligned-btn"
                  :loading="isTargetQuerying"
                  :disabled="versionItems.length === 0 || isListingVersions || isBatchQuerying"
                  @click="handleTargetQueryClick"
                >
                  {{ isTargetQuerying ? '停止查询' : '查找指定版本' }}
                </n-button>
              </div>

              <div class="filter-search-row">
                <div class="filter-search-box">
                  <input
                    v-model="versionForm.filter"
                    class="clean-input filter-input height-aligned"
                    placeholder="输入版本号 (如 10.2.80)、构建 ID 或体积进行实时筛选..."
                  />
                </div>
              </div>
            </div>
          </div>

          <div class="clean-card table-flex-card">
            <n-data-table
              :columns="versionColumns"
              :data="filteredVersions"
              :loading="isListingVersions"
              :virtual-scroll="true"
              :scroll-x="720"
              flex-height
              style="height: 100%;"
              size="small"
            >
              <template #empty>
                <div class="table-empty-box">
                  <div v-if="isHanddrawn" class="handdrawn-table-empty-doodle" aria-hidden="true">
                    <svg viewBox="0 0 80 72" width="68" height="60" fill="none" class="empty-doodle-svg">
                      <path d="M24 16h32M24 56h32" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" />
                      <path d="M28 16c0 14 24 18 24 20s-24 6-24 20h24c0-14-24-18-24-20s24-6 24-20H28z" class="doodle-paper" stroke="currentColor" stroke-width="1.8" />
                      <ellipse cx="40" cy="51" rx="8" ry="3" class="doodle-ground-wash" />
                      <circle cx="40" cy="36" r="1.5" fill="currentColor" />
                      <path d="M62 20l1 2.2 2.2 1-2.2 1-1 2.2-1-2.2-2.2-1 2.2-1 1-2.2z" class="doodle-star" />
                    </svg>
                  </div>
                  <div class="empty-state-title">{{ versionForm.bundleId ? '未查询到版本信息' : '历史版本时光机' }}</div>
                  <div class="empty-state-desc">{{ versionForm.bundleId ? '请检查 Bundle ID 是否正确，或当前账号是否具备权限' : '输入应用 Bundle ID 并点击「获取历史版本」，即可枚举所有构建历史' }}</div>
                </div>
              </template>
            </n-data-table>
          </div>
        </div>

        <!-- VIEW 3: 下载中心 (download) -->
        <div v-show="activeTab === 'download'" class="view-panel">
          <div class="view-header">
            <div>
              <div class="title-with-badge">
                <h2 class="view-title">下载中心</h2>
                <span v-if="isHanddrawn" class="handdrawn-view-badge" aria-hidden="true"># 任务清单 📥</span>
              </div>
              <p class="view-desc">正在下载与已完成的 IPA 任务管理，自动归档至对应账号目录</p>
            </div>
            <div class="btn-group-row">
              <n-button
                secondary
                size="small"
                class="small-aligned-btn"
                :disabled="completedTaskCount === 0"
                @click="handleClearCompleted"
              >
                清空已完成
              </n-button>
              <n-button
                secondary
                size="small"
                class="small-aligned-btn"
                @click="handleOpenDefaultDownloadDir"
              >
                打开存储目录
              </n-button>
            </div>
          </div>

          <div class="tasks-container">
            <div v-if="downloadTasks.length === 0" class="empty-state-card">
              <!-- 手绘风格专属小插图 (可爱待办清单与铅笔) -->
              <div v-if="isHanddrawn" class="handdrawn-empty-doodle" aria-hidden="true">
                <svg viewBox="0 0 68 68" width="56" height="56" fill="none" class="empty-doodle-svg">
                  <rect x="18" y="15" width="32" height="42" rx="5" class="doodle-paper" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
                  <path d="M26 15v-3a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v3" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
                  <path d="M23 26l2 2 3-3" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" class="doodle-check" />
                  <line x1="32" y1="26" x2="43" y2="26" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
                  <path d="M23 35l2 2 3-3" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" class="doodle-check" />
                  <line x1="32" y1="35" x2="43" y2="35" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
                  <line x1="25" y1="44" x2="41" y2="44" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-dasharray="2 3" />
                  <g transform="translate(42, 28) rotate(22)">
                    <rect x="0" y="0" width="7" height="24" rx="2" class="doodle-pencil" stroke="currentColor" stroke-width="1.6" />
                    <path d="M0 24l3.5 6 3.5-6z" class="doodle-pencil-tip" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round" />
                  </g>
                  <path d="M11 20l1 2.2 2.2 1-2.2 1-1 2.2-1-2.2-2.2-1 2.2-1 1-2.2z" class="doodle-star" />
                </svg>
              </div>
              <div class="empty-title">暂无下载任务</div>
              <div class="empty-desc">前往「应用搜索」或「历史版本」点击下载即可在此实时追踪进度</div>
            </div>

            <div v-else class="task-grid">
              <div v-for="task in downloadTasks" :key="task.id" class="modern-task-card">
                <div class="task-card-header">
                  <div class="task-app-title-group">
                    <span class="task-name">{{ task.appName }}</span>
                    <span class="task-pill task-pill-ver">{{ task.version }}</span>
                    <span v-if="task.versionId" class="task-pill task-pill-build">Build {{ task.versionId }}</span>
                  </div>
                  <div class="task-actions-group">
                    <n-button
                      v-if="task.status === 'downloading'"
                      size="tiny"
                      type="error"
                      secondary
                      @click="handleCancelTask(task.id)"
                    >
                      取消
                    </n-button>
                    <n-button
                      v-if="task.status === 'completed'"
                      size="tiny"
                      type="primary"
                      @click="handleInstallFromTask(task.outputPath)"
                    >
                      安装到设备
                    </n-button>
                    <n-button
                      v-if="task.status === 'error' || task.status === 'canceled'"
                      size="tiny"
                      secondary
                      @click="handleRetryTask(task)"
                    >
                      重试
                    </n-button>
                    <n-button size="tiny" quaternary @click="handleDeleteTask(task.id)">
                      删除
                    </n-button>
                  </div>
                </div>

                <div class="task-meta-bundle">{{ task.bundleID }}</div>

                <div class="task-progress-section">
                  <n-progress
                    type="line"
                    :percentage="task.progress"
                    :status="getTaskProgressStatus(task.status)"
                    :show-indicator="false"
                    :height="6"
                    border-radius="3"
                  />
                </div>

                <div class="task-card-footer">
                  <div class="footer-status-tag">
                    <span class="status-indicator" :class="'indicator-' + task.status"></span>
                    <span class="status-text">{{ getTaskStatusText(task.status) }}</span>
                    <span v-if="task.status === 'downloading' && task.speed" class="task-speed">{{ task.speed }}</span>
                    <span v-if="task.status === 'downloading' && task.totalBytes > 0" class="task-bytes-info">
                      {{ formatTaskBytes(task.currBytes) }} / {{ formatTaskBytes(task.totalBytes) }}
                    </span>
                    <span v-else-if="task.fileSize && task.fileSize !== '-'" class="task-bytes-info">
                      体积: {{ task.fileSize }}
                    </span>
                    <span v-if="task.status === 'error'" class="task-error-text" :title="task.errorMessage">
                      {{ task.errorMessage }}
                    </span>
                  </div>

                  <div class="footer-right-info">
                    <span class="pct-text font-bold">{{ task.progress }}%</span>
                    <span class="time-text">{{ task.createdAt }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- VIEW 4: 已购应用 (purchased) -->
        <div v-show="activeTab === 'purchased'" class="view-panel">
          <div class="view-header">
            <div>
              <div class="title-with-badge">
                <h2 class="view-title">已购应用</h2>
                <span v-if="isHanddrawn" class="handdrawn-view-badge" aria-hidden="true"># 珍藏宝库 🛍️</span>
              </div>
              <p class="view-desc">浏览与检索当前 Apple ID 名下已获得正版许可的历史应用库</p>
            </div>
            <div class="purchased-header-actions">
              <n-button
                v-if="purchasedApps.length < purchasedTotal"
                secondary
                size="small"
                class="small-aligned-btn"
                :loading="isLoadingAllPurchases"
                @click="loadAllPurchases"
                title="拉取名下所有已购记录，以便完整检索所有应用"
              >
                加载全部已购 ({{ purchasedApps.length }}/{{ purchasedTotal }})
              </n-button>
              <n-button
                type="primary"
                size="small"
                class="small-aligned-btn ml-2"
                :loading="isPurchasedLoading"
                @click="loadPurchases"
              >
                刷新列表
              </n-button>
            </div>
          </div>

          <!-- 搜索与筛选工具栏 -->
          <div class="clean-card mb-3 purchased-search-toolbar">
            <div class="purchased-search-left">
              <!-- 搜索输入框 -->
              <div class="clean-input-box purchased-search-input-box">
                <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" class="search-icon">
                  <circle cx="11" cy="11" r="8"></circle>
                  <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                </svg>
                <input
                  v-model="purchasedSearchKeyword"
                  class="clean-input purchased-search-input"
                  placeholder="搜索已购应用名称、Bundle ID 或 App ID..."
                  @keydown.esc="purchasedSearchKeyword = ''"
                />
                <button
                  v-if="purchasedSearchKeyword"
                  class="search-clear-btn"
                  @click="purchasedSearchKeyword = ''"
                  title="清空搜索"
                >
                  <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round">
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                  </svg>
                </button>
              </div>

              <!-- 搜索匹配统计标签 -->
              <div class="purchased-search-badge" :class="{ 'has-filter': !!purchasedSearchKeyword }">
                <span v-if="purchasedSearchKeyword">
                  找到 <b>{{ filteredPurchasedApps.length }}</b> 款匹配应用 (共 {{ purchasedApps.length }} 款)
                </span>
                <span v-else>
                  当前显示 <b>{{ purchasedApps.length }}</b> 款 (账户共 {{ purchasedTotal }} 款)
                </span>
              </div>
            </div>

            <!-- 右侧分页控制 -->
            <div class="purchased-page-controls">
              <div class="page-size-selector">
                <span class="size-label">每页</span>
                <n-select
                  v-model:value="purchasedPageSize"
                  size="tiny"
                  style="width: 76px;"
                  :options="[
                    { label: '20', value: 20 },
                    { label: '50', value: 50 },
                    { label: '100', value: 100 }
                  ]"
                  @update:value="onPurchasedPageSizeChange"
                />
              </div>
              <n-button
                secondary
                size="small"
                class="small-aligned-btn"
                :disabled="purchasedPage <= 1 || isPurchasedLoading"
                @click="prevPurchasedPage"
              >
                上一页
              </n-button>
              <span class="page-indicator">第 {{ purchasedPage }} 页</span>
              <n-button
                secondary
                size="small"
                class="small-aligned-btn"
                :disabled="purchasedPage * purchasedPageSize >= purchasedTotal || isPurchasedLoading"
                @click="nextPurchasedPage"
              >
                下一页
              </n-button>
            </div>
          </div>

          <!-- 表格主体 -->
          <div class="clean-card table-flex-card">
            <n-data-table
              :columns="purchasedColumns"
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
                    <div v-if="isHanddrawn" class="handdrawn-table-empty-doodle" aria-hidden="true">
                      <svg viewBox="0 0 80 72" width="68" height="60" fill="none" class="empty-doodle-svg">
                        <circle cx="36" cy="34" r="16" class="doodle-paper" stroke="currentColor" stroke-width="1.8" />
                        <line x1="48" y1="46" x2="64" y2="62" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" />
                        <path d="M30 34h12" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
                      </svg>
                    </div>
                    <div v-else class="search-none-icon">🔍</div>
                    <div class="search-none-title">未找到与「{{ purchasedSearchKeyword }}」匹配的已购应用</div>
                    <div class="search-none-desc">请尝试输入不同关键词，或点击下方按钮清空搜索</div>
                    <n-button size="tiny" secondary class="mt-2" @click="purchasedSearchKeyword = ''">
                      清空搜索
                    </n-button>
                  </div>
                  <div v-else class="search-none-state">
                    <div v-if="isHanddrawn" class="handdrawn-table-empty-doodle" aria-hidden="true">
                      <svg viewBox="0 0 80 72" width="68" height="60" fill="none" class="empty-doodle-svg">
                        <path d="M22 26h36l3 32a3 3 0 0 1-3 3H22a3 3 0 0 1-3-3l3-32z" class="doodle-paper" stroke="currentColor" stroke-width="1.8" />
                        <path d="M32 26v-6a8 8 0 0 1 16 0v6" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
                        <path d="M36 38c2 3 6 3 8 0" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
                        <path d="M60 18l1.2 2.2 2.2 1.2-2.2 1.2-1.2 2.2-1.2-2.2-2.2-1.2 2.2-1.2 1.2-2.2z" class="doodle-star" />
                      </svg>
                    </div>
                    <div class="search-none-title">暂无已购应用记录</div>
                    <div class="search-none-desc">点击右上角「加载全部已购」或刷新获取当前 Apple ID 历史正版应用</div>
                  </div>
                </div>
              </template>
            </n-data-table>
          </div>
        </div>

        <!-- VIEW 5: 设备直装 (installer) -->
        <div v-show="activeTab === 'installer'" class="view-panel">
          <div class="view-header">
            <div>
              <div class="title-with-badge">
                <h2 class="view-title">设备直装</h2>
                <span v-if="isHanddrawn" class="handdrawn-view-badge" aria-hidden="true"># 极速快传 📱</span>
              </div>
              <p class="view-desc">通过 USB 数据线将正版签署的 IPA 应用包一键安装至 iOS 设备</p>
            </div>
          </div>

          <div class="installer-grid">
            <!-- Left: IPA File Selection -->
            <div class="clean-card flex-col">
              <div class="card-headline">
                <span class="headline-title">1. 选择安装包</span>
                <span v-if="selectedIPAPath" class="headline-badge">已就绪</span>
              </div>

              <div
                class="clean-drop-zone"
                :class="{ 'drop-active': !!selectedIPAPath }"
                @click="handleSelectIPA"
              >
                <!-- 手绘风格拖拽包裹小插图 -->
                <div v-if="isHanddrawn && !selectedIPAPath" class="handdrawn-drop-doodle" aria-hidden="true">
                  <svg viewBox="0 0 54 54" width="44" height="44" fill="none" class="drop-doodle-svg">
                    <path d="M11 22l16-8 16 8-16 8-16-8z" class="doodle-box-top" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" />
                    <path d="M11 22v15l16 9 16-9V22" class="doodle-box-side" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" />
                    <line x1="27" y1="30" x2="27" y2="46" stroke="currentColor" stroke-width="1.6" stroke-dasharray="2 3" />
                    <path d="M27 7l1 2.5 2.5 1-2.5 1-1 2.5-1-2.5-2.5-1 2.5-1 1-2.5z" class="doodle-star" />
                    <circle cx="40" cy="9" r="1.2" class="doodle-sparkle" />
                  </svg>
                </div>
                <div class="drop-primary-title">
                  {{ selectedIPAPath ? selectedIPAFileName : '点击选择或拖拽 .ipa 文件到此处' }}
                </div>
                <div class="drop-secondary-path text-ellipsis" :title="selectedIPAPath">
                  {{ selectedIPAPath || '支持标准 iOS 签名应用包格式' }}
                </div>
              </div>

              <div class="mt-3 flex-align-center">
                <input
                  :value="selectedIPAPath"
                  readonly
                  class="clean-input flex-1 mr-2 height-aligned"
                  placeholder="尚未选择文件"
                />
                <n-button secondary size="medium" class="height-aligned-btn" @click="handleSelectIPA">浏览文件</n-button>
              </div>
            </div>

            <!-- Right: Device Selection -->
            <div class="clean-card flex-col">
              <div class="card-headline">
                <span class="headline-title">2. 选择苹果设备</span>
                <n-button
                  secondary
                  size="tiny"
                  :loading="isLoadingDevices"
                  :disabled="isInstallingIPA"
                  @click="loadConnectedDevices"
                >
                  刷新检测
                </n-button>
              </div>

              <n-select
                v-model:value="selectedDeviceUDID"
                :options="deviceOptions"
                :loading="isLoadingDevices"
                :disabled="isInstallingIPA"
                placeholder="请选择已连接的 iOS 设备"
                size="large"
              />

              <div class="sub-alert-box mt-3">
                <span>请保持设备屏幕<b>常亮解锁</b>；若设备息屏休眠，USB 通信将中断并丢失连接。</span>
              </div>

              <!-- Device Specs Box -->
              <div v-if="selectedDevice" class="device-spec-box mt-3">
                <div class="spec-row">
                  <span class="spec-k">设备名称</span>
                  <span class="spec-v font-bold">{{ selectedDevice.name }}</span>
                </div>
                <div class="spec-row">
                  <span class="spec-k">设备型号</span>
                  <span class="spec-v">{{ selectedDevice.productType || '-' }}</span>
                </div>
                <div class="spec-row">
                  <span class="spec-k">系统版本</span>
                  <span class="spec-v">{{ selectedDevice.productVersion || '-' }}</span>
                </div>
                <div class="spec-row">
                  <span class="spec-k">连接模式</span>
                  <span class="spec-v">{{ selectedDevice.connectionType || '-' }}</span>
                </div>
                <div class="spec-row">
                  <span class="spec-k">UDID</span>
                  <span class="spec-v text-ellipsis" :title="selectedDevice.udid">{{ selectedDevice.udid }}</span>
                </div>
              </div>

              <div v-else-if="!isLoadingDevices && devices.length === 0" class="no-device-box mt-3">
                <div class="no-device-text">未检测到已连接的苹果设备</div>
                <div class="no-device-sub">请直连电脑 USB 接口、点亮屏幕、输入密码并信任此电脑</div>
              </div>
            </div>
          </div>

          <!-- Bottom Action Bar for Installer -->
          <div class="clean-card mt-4 installer-action-banner">
            <div class="installer-action-info">
              <div class="action-banner-title">
                {{ !selectedIPAPath ? '请先选择待安装的 IPA 文件' : (!selectedDeviceUDID ? '请选择目标苹果设备' : '就绪，可以开始安装') }}
              </div>
              <div class="action-banner-desc">
                本工具下载的正版 IPA 需安装至登录了相同 Apple ID 的设备上，未签名包将无法被系统接受。
              </div>
            </div>

            <n-button
              type="primary"
              size="large"
              :disabled="!selectedIPAPath || !selectedDeviceUDID || isLoadingDevices"
              :loading="isInstallingIPA"
              @click="handleInstallIPA"
              class="install-submit-btn"
            >
              {{ isInstallingIPA ? '正在安装中...' : '开始安装到设备' }}
            </n-button>
          </div>
        </div>

        <!-- VIEW 6: 账号中心 (account) -->
        <div v-show="activeTab === 'account'" class="view-panel">
          <div class="view-header">
            <div>
              <div class="title-with-badge">
                <h2 class="view-title">账号中心</h2>
                <span v-if="isHanddrawn" class="handdrawn-view-badge" aria-hidden="true"># 身份档案 👤</span>
              </div>
              <p class="view-desc">管理用于 App Store 正版授权通信与 IPA 下载的 Apple ID 身份凭证</p>
            </div>
          </div>

          <div class="two-columns-layout">
            <!-- Account Status Card -->
            <div class="clean-card">
              <div class="card-headline">
                <span class="headline-title">当前登录凭证</span>
                <span class="account-pill" :class="isLoggedIn ? 'pill-success' : 'pill-gray'">
                  {{ isLoggedIn ? '已授权' : '未登录' }}
                </span>
              </div>

              <div class="account-profile-box">
                <div class="large-avatar" :class="{ 'avatar-active': isLoggedIn }">
                  {{ isLoggedIn ? (account.name ? account.name.charAt(0).toUpperCase() : '') : '' }}
                </div>
                <div class="large-profile-info">
                  <div class="profile-name">{{ account.name || '尚未登录 Apple ID' }}</div>
                  <div class="profile-email">{{ account.email || '请在右侧输入账号密码完成登录' }}</div>
                </div>
              </div>

              <div class="sub-alert-box mt-4">
                <span>官方直接认证：所有凭据直接向 Apple 官方接口请求并保存在本地钥匙串，不经过任何第三方服务器。</span>
              </div>

              <div class="account-card-actions mt-4">
                <n-button
                  type="error"
                  secondary
                  :disabled="!isLoggedIn"
                  @click="handleRevoke"
                  :loading="isRevoking"
                  class="height-aligned-btn"
                >
                  退出登录
                </n-button>
                <n-button
                  secondary
                  @click="handleClearKeychain"
                  :loading="isClearing"
                  title="清理本地钥匙串缓存解决校验异常"
                  class="height-aligned-btn"
                >
                  清理钥匙串缓存
                </n-button>
                <n-button
                  secondary
                  @click="() => refreshAccount(false)"
                  :loading="isAccountLoading"
                  class="height-aligned-btn"
                >
                  刷新状态
                </n-button>
              </div>
            </div>

            <!-- Login Form Card -->
            <div class="clean-card">
              <div class="card-headline">
                <span class="headline-title">登录 Apple ID</span>
              </div>

              <div class="form-item-clean">
                <label class="clean-label">Apple ID 账户邮箱</label>
                <input
                  v-model="loginForm.email"
                  class="clean-input height-aligned"
                  placeholder="例如: your_apple_id@icloud.com"
                />
              </div>

              <div class="form-item-clean mt-3">
                <label class="clean-label">Apple ID 密码</label>
                <input
                  v-model="loginForm.password"
                  type="password"
                  class="clean-input height-aligned"
                  placeholder="请输入 Apple ID 密码"
                  @keydown.enter="handleLogin"
                />
              </div>

              <div class="sub-alert-box mt-3">
                <span>双重认证 (2FA)：若账号开启了 2FA，点击登录后将自动弹出 6 位验证码输入窗口。</span>
              </div>

              <div class="mt-4">
                <n-button
                  type="primary"
                  block
                  size="large"
                  :loading="isLoggingIn"
                  @click="handleLogin"
                  style="height: 40px;"
                >
                  登录
                </n-button>
              </div>
            </div>
          </div>
        </div>

        <!-- VIEW 7: 系统设置 (settings) -->
        <div v-show="activeTab === 'settings'" class="view-panel">
          <div class="view-header">
            <div>
              <div class="title-with-badge">
                <h2 class="view-title">系统设置</h2>
                <span v-if="isHanddrawn" class="handdrawn-view-badge" aria-hidden="true"># 个性工坊 ⚙️</span>
              </div>
              <p class="view-desc">全局参数、网络代理、下载路径与基础运行引擎配置</p>
            </div>
            <n-button
              type="primary"
              size="medium"
              class="height-aligned-btn"
              @click="handleSaveSettings"
            >
              保存并应用设置
            </n-button>
          </div>

          <div class="clean-card settings-stack">
            <!-- Section 1: Engine Status -->
            <div class="settings-group">
              <div class="settings-group-title">底层引擎与环境</div>
              <div class="settings-field-row">
                <div class="field-meta">
                  <div class="field-title">CLI 引擎状态</div>
                  <div class="field-desc">自动检测 tools/ 目录与系统环境变量中的 ipatool 和 go-ios</div>
                </div>
                <div class="field-control">
                  <div class="engine-badge-box">
                    <span class="engine-dot"></span>
                    <span class="engine-text">引擎就绪</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Section 2: Passphrase -->
            <div class="settings-group">
              <div class="settings-group-title">安全与钥匙串</div>
              <div class="settings-field-row">
                <div class="field-meta">
                  <div class="field-title">密钥库解锁密码 (--keychain-passphrase)</div>
                  <div class="field-desc">自动注入命令行参数，彻底杜绝 Windows 终端弹窗与死锁卡死</div>
                </div>
                <div class="field-control">
                  <input
                    v-model="settings.keychainPassphrase"
                    class="clean-input height-aligned w-full"
                    placeholder="输入密钥库密码（默认 123456）"
                  />
                </div>
              </div>
            </div>

            <!-- Section 3: Proxy -->
            <div class="settings-group">
              <div class="settings-group-title">网络与代理</div>
              <div class="settings-field-row">
                <div class="field-meta">
                  <div class="field-title">启用网络代理</div>
                  <div class="field-desc">App Store 接口认证时通过环境变量透传 HTTP / SOCKS5 代理</div>
                </div>
                <div class="field-control">
                  <n-switch v-model:value="settings.enableProxy" />
                </div>
              </div>

              <div class="settings-field-row">
                <div class="field-meta">
                  <div class="field-title">代理服务器地址</div>
                  <div class="field-desc">支持 Clash / v2rayN 等常见本地代理端口（如 http://127.0.0.1:10808）</div>
                </div>
                <div class="field-control">
                  <div class="setting-input-action-group">
                    <input
                      v-model="settings.proxyUrl"
                      :disabled="!settings.enableProxy"
                      class="clean-input flex-1 height-aligned"
                      placeholder="http://127.0.0.1:10808"
                    />
                    <n-button
                      secondary
                      size="medium"
                      class="height-aligned-btn setting-fixed-btn"
                      :disabled="!settings.enableProxy"
                      :loading="isTestingProxy"
                      @click="handleTestProxy"
                    >
                      测试
                    </n-button>
                  </div>
                </div>
              </div>
            </div>

            <!-- Section 4: Storage -->
            <div class="settings-group">
              <div class="settings-group-title">存储与下载</div>
              <div class="settings-field-row">
                <div class="field-meta">
                  <div class="field-title">IPA 存储路径</div>
                  <div class="field-desc">按登录账号动态归档至 data/downloads/账号标识 目录</div>
                </div>
                <div class="field-control">
                  <div class="setting-input-action-group">
                    <input
                      :value="settings.defaultDownloadDir"
                      readonly
                      disabled
                      class="clean-input flex-1 height-aligned"
                    />
                    <n-button
                      secondary
                      size="medium"
                      class="height-aligned-btn setting-fixed-btn"
                      @click="handleOpenDefaultDownloadDir"
                    >
                      打开
                    </n-button>
                  </div>
                </div>
              </div>

              <div class="settings-field-row">
                <div class="field-meta">
                  <div class="field-title">默认目标平台</div>
                  <div class="field-desc">检索与下载默认针对的 Apple 硬件体系</div>
                </div>
                <div class="field-control">
                  <n-select
                    v-model:value="settings.defaultPlatform"
                    :options="platformOptions"
                    size="medium"
                    style="width: 100%;"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- VIEW 8: 关于软件 (about) -->
        <div v-show="activeTab === 'about'" class="view-panel">
          <div class="view-header">
            <div>
              <div class="title-with-badge">
                <h2 class="view-title">关于软件</h2>
                <span v-if="isHanddrawn" class="handdrawn-view-badge" aria-hidden="true"># 灵感手账 🍎</span>
              </div>
              <p class="view-desc">果仓助手 (AppleVault) 软件信息、项目开源地址与使用说明</p>
            </div>
          </div>

          <div class="clean-card about-card-stack">
            <!-- App Banner -->
            <div class="about-hero">
              <div class="about-large-icon">
                <svg class="brand-apple-svg" style="width: 38px; height: 38px;" viewBox="0 0 170 170" fill="currentColor">
                  <path d="M150.37 130.25c-2.45 5.66-5.35 10.87-8.71 15.66-4.58 6.53-8.33 11.05-11.22 13.56-4.48 4.12-9.28 6.23-14.42 6.35-3.69 0-8.14-1.05-13.32-3.18-5.19-2.12-9.97-3.17-14.34-3.17-4.58 0-9.49 1.05-14.75 3.17-5.26 2.13-9.5 3.24-12.74 3.35-4.35.13-9.16-1.9-14.42-6.08-3.7-3.08-7.7-7.85-12.01-14.3-6.24-9.35-11.12-20.2-14.65-32.54-3.52-12.35-5.29-24.3-5.29-35.87 0-14.12 3.52-25.75 10.57-34.89 7.05-9.14 16.03-13.88 26.94-14.21 4.79 0 10.36 1.34 16.71 4.02 6.36 2.68 10.15 4.08 11.37 4.19 1.12-.11 5.02-1.57 11.7-4.38 6.68-2.82 12.35-4.08 17.02-3.78 12.79.89 23.01 5.66 30.65 14.31-11.29 6.81-16.79 16.32-16.5 28.53.33 9.61 4.2 17.58 11.62 23.9 7.42 6.32 16.31 9.94 26.68 10.86-2.12 6.54-4.53 13.06-7.24 19.56zm-29.35-104.9c-.11 4.14-1.55 8.35-4.32 12.63-2.77 4.28-6.42 7.74-10.96 10.38-3.02 1.63-6.21 2.72-9.56 3.27-.11-1.3-.11-2.4-.11-3.27 0-4.13 1.54-8.38 4.63-12.75 3.09-4.37 7.02-7.86 11.8-10.47 2.91-1.63 5.75-2.73 8.52-3.3 0 1.2.06 2.37 0 3.51z"/>
                </svg>
              </div>
              <div class="about-hero-text">
                <div class="about-app-title">果仓助手 (AppleVault)</div>
                <div class="about-version-line">
                  <span class="about-version-badge">版本 v1.0.0</span>
                  <span class="about-badge-sub">基于 Wails & Go 构建</span>
                </div>
                <p class="about-intro">
                  现代优雅的 Apple App Store 正版应用与历史版本下载管理工具，支持 iOS 设备一键直装。
                </p>
              </div>
            </div>

            <!-- GitHub Repo Section -->
            <div class="about-section-box">
              <div class="about-section-label">GitHub 官方开源仓库</div>
              <div class="repo-link-bar">
                <input
                  value="https://github.com/wnnz/AppleVault"
                  readonly
                  class="clean-input flex-1 height-aligned"
                />
                <n-button
                  type="primary"
                  size="medium"
                  class="height-aligned-btn"
                  @click="openGitHub"
                >
                  访问 GitHub
                </n-button>
                <n-button
                  secondary
                  size="medium"
                  class="height-aligned-btn"
                  @click="copyGitHubUrl"
                >
                  复制地址
                </n-button>
              </div>
            </div>

            <!-- Info Grid -->
            <div class="about-features-grid">
              <div class="about-feature-item">
                <div class="feature-title">官方正版授权</div>
                <div class="feature-desc">使用自己的 Apple ID 账户直接对接 App Store，下载包含个人官方凭据的正版 IPA，装机稳定不闪退。</div>
              </div>
              <div class="about-feature-item">
                <div class="feature-title">历史版本检索</div>
                <div class="feature-desc">支持输入任意 Bundle ID 查询完整历史构建记录与版本详情，可指定目标版本快速查找并一键下载。</div>
              </div>
              <div class="about-feature-item">
                <div class="feature-title">真机无缝安装</div>
                <div class="feature-desc">连接 iPhone / iPad 数据线即可自动识别设备规格与系统版本，将有效签名的应用包直接安装到设备。</div>
              </div>
              <div class="about-feature-item">
                <div class="feature-title">绿色纯净便携</div>
                <div class="feature-desc">无需安装器与后台常驻，所有配置与下载文件按登录账号自动隔离归档在本地 data 目录中。</div>
              </div>
            </div>

            <!-- Open Source Acknowledgement & Notice -->
            <div class="about-footer-notice">
              <span>本项目仅为 App Store 官方接口的图形化辅助工具，与 Apple Inc. 无官方隶属关系。感谢开源社区优秀项目 ipatool 与 go-ios 提供的底层能力支持。</span>
            </div>
          </div>
        </div>

      </div>

      <!-- Bottom Minimal Status Bar -->
      <div class="bottom-status-bar">
        <div class="status-indicator-section">
          <span class="status-pulse" :class="{ 'pulse-busy': isAnyOperationRunning }"></span>
          <span class="status-summary-text text-ellipsis">{{ statusText }}</span>
        </div>

        <div class="status-bar-tools">
          <n-button
            v-if="isAnyOperationRunning"
            size="tiny"
            type="error"
            secondary
            @click="handleCancel"
            class="mr-2"
          >
            终止操作
          </n-button>

          <button class="status-btn" @click="toggleLogs">
            {{ showLogs ? '收起日志' : '实时日志' }}
          </button>
        </div>
      </div>

      <!-- Collapsible Sleek Terminal Drawer -->
      <div v-show="showLogs" class="sleek-console-drawer">
        <div class="console-action-bar">
          <div class="console-title">实时执行日志</div>
          <div class="console-btns">
            <button class="console-bar-btn" @click="clearLogs">清空</button>
            <button class="console-bar-btn" @click="toggleLogs">收起</button>
          </div>
        </div>
        <div ref="logContainerRef" class="console-scroll-screen">
          <div v-if="logLines.length === 0" class="console-empty-tip">等待命令输出...</div>
          <div
            v-for="(log, idx) in logLines"
            :key="idx"
            class="console-row"
            :class="{ 'row-err': log.includes('[ERR]') }"
          >
            {{ log }}
          </div>
        </div>
      </div>
    </section>

    <!-- 2FA Modal Dialog -->
    <n-modal
      v-model:show="show2FAModal"
      preset="card"
      title="Apple ID 双重认证"
      style="width: 400px; border-radius: 14px;"
      :mask-closable="false"
    >
      <div class="modal-dialog-inner">
        <p class="dialog-desc">已向你的受信任 Apple 设备发送了验证码，请输入 6 位数字验证码：</p>
        <input
          ref="twoFAInputRef"
          v-model="twoFACode"
          class="clean-input text-center font-bold text-lg height-aligned"
          placeholder="6 位验证码"
          maxlength="6"
          autofocus
          @keydown.enter="confirm2FA"
          style="letter-spacing: 6px; height: 44px; font-size: 20px;"
        />
        <div class="dialog-action-buttons mt-4">
          <n-button secondary @click="cancel2FA" class="height-aligned-btn">取消</n-button>
          <n-button
            type="primary"
            :disabled="twoFACode.length !== 6"
            :loading="isLoggingIn"
            @click="confirm2FA"
            class="height-aligned-btn"
          >
            提交验证
          </n-button>
        </div>
      </div>
    </n-modal>

    <!-- Target Version Modal Dialog -->
    <n-modal
      v-model:show="showTargetVersionModal"
      preset="card"
      title="查找指定版本"
      style="width: 440px; border-radius: 14px;"
    >
      <div class="modal-dialog-inner">
        <p class="dialog-desc">
          请输入目标版本号（例如 <code>10.2.80</code> 或 <code>8.0.0</code>），程序将智能检索历史构建记录并快速定位匹配版本。
        </p>
        <input
          ref="targetVersionInputRef"
          v-model="targetVersionInput"
          class="clean-input height-aligned"
          placeholder="例如: 10.2.80"
          autofocus
          @keydown.enter="confirmStartTargetQuery"
        />
        <div class="dialog-action-buttons mt-4">
          <n-button secondary @click="showTargetVersionModal = false" class="height-aligned-btn">取消</n-button>
          <n-button
            type="primary"
            :disabled="!targetVersionInput.trim()"
            @click="confirmStartTargetQuery"
            class="height-aligned-btn"
          >
            开始查询
          </n-button>
        </div>
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, h } from 'vue'
import {
  useMessage,
  useDialog,
  NButton,
  NSelect,
  NSwitch,
  NModal,
  NDataTable,
  NProgress,
  NSpace
} from 'naive-ui'
import {
  AddDownloadTask,
  GetDownloadTasks,
  CancelDownloadTask,
  DeleteDownloadTask,
  ClearCompletedDownloadTasks,
  GetAccountInfo,
  Login,
  Revoke,
  ClearKeychainCache,
  Search,
  ListVersions,
  GetVersionMetadata,
  Purchase,
  ListPurchases,
  GetSettings,
  SaveSettings,
  SelectDirectory,
  TestProxy,
  OpenInExplorer,
  CancelRunningCommand,
  SelectIPA,
  ListDevices,
  InstallIPA
} from '../../wailsjs/go/main/App'
import { EventsOn, OnFileDrop, OnFileDropOff, BrowserOpenURL } from '../../wailsjs/runtime/runtime'
import { main } from '../../wailsjs/go/models'
import { AppTheme, ThemeOption } from '../types/theme'

const props = withDefaults(
  defineProps<{
    currentTheme?: AppTheme
    isDark?: boolean
  }>(),
  {
    currentTheme: 'minimal-light',
    isDark: false
  }
)

const emit = defineEmits<{
  (e: 'update:currentTheme', theme: AppTheme): void
  (e: 'toggleTheme'): void
}>()

const showThemePopover = ref(false)

const effectiveTheme = computed<AppTheme>(() => props.currentTheme || (props.isDark ? 'minimal-dark' : 'minimal-light'))
const isDarkTheme = computed(() => effectiveTheme.value.endsWith('-dark') || props.isDark)
const isHanddrawn = computed(() => effectiveTheme.value.startsWith('handdrawn'))

const themeList: ThemeOption[] = [
  {
    key: 'minimal-light',
    name: '简约 - 浅色',
    category: 'minimal',
    categoryName: '简约风格',
    description: '清爽现代、明亮通透',
    preview: {
      bg: '#f8fafc',
      card: '#ffffff',
      border: '#cbd5e1',
      accent: '#0071e3',
      text: '#1e293b'
    }
  },
  {
    key: 'minimal-dark',
    name: '简约 - 深色',
    category: 'minimal',
    categoryName: '简约风格',
    description: '沉浸暗色、夜间舒适',
    preview: {
      bg: '#0f172a',
      card: '#1e293b',
      border: '#334155',
      accent: '#0284c7',
      text: '#f1f5f9'
    }
  },
  {
    key: 'handdrawn-light',
    name: '手绘 - 浅色',
    category: 'handdrawn',
    categoryName: '手绘风格',
    description: '水彩插画、清新手账',
    preview: {
      bg: '#d3e8f8',
      card: '#ffffff',
      border: '#b9ddfb',
      accent: '#4092ea',
      text: '#1e3a5f'
    }
  },
  {
    key: 'handdrawn-dark',
    name: '手绘 - 深色',
    category: 'handdrawn',
    categoryName: '手绘风格',
    description: '夜空水彩、清澈静谧',
    preview: {
      bg: '#131c28',
      card: '#1b2636',
      border: '#3b82f6',
      accent: '#60a5fa',
      text: '#e2e8f0'
    }
  }
]

function selectTheme(themeKey: AppTheme) {
  emit('update:currentTheme', themeKey)
  showThemePopover.value = false
}

const message = useMessage()
const dialog = useDialog()

const activeTab = ref('search')
const isAnyOperationRunning = ref(false)
const statusText = ref('就绪')

// Settings
const settings = ref<main.Settings>({
  keychainPassphrase: '123456',
  defaultDownloadDir: 'data/downloads/default',
  defaultPlatform: 'iphone',
  enableProxy: true,
  proxyUrl: 'http://127.0.0.1:10808',
  ipaToolPath: ''
})

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

// Account State
const account = ref<main.AccountInfo>({ name: '', email: '', success: false })
const isLoggedIn = computed(() => account.value.success && !!account.value.email)
const isAccountLoading = ref(false)
const isRevoking = ref(false)
const isClearing = ref(false)
const isLoggingIn = ref(false)

const loginForm = ref({
  email: '',
  password: ''
})

// 2FA Modal
const show2FAModal = ref(false)
const twoFACode = ref('')
const twoFAInputRef = ref()

// Search State
const isSearching = ref(false)
const searchForm = ref({
  term: '',
  limit: 10,
  platform: 'iphone'
})
const searchResults = ref<main.AppItem[]>([])

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
      return h(NSpace, { size: 6, wrap: false }, () => [
        h(NButton, { size: 'tiny', secondary: true, onClick: () => selectAppForVersions(row) }, () => '历史版本'),
        h(NButton, { size: 'tiny', type: 'primary', onClick: () => downloadFromSearch(row) }, () => '下载'),
        h(NButton, { size: 'tiny', secondary: true, onClick: () => handlePurchaseApp(row.bundleID) }, () => '获取许可')
      ])
    }
  }
]

// Versions State
const isListingVersions = ref(false)
const isBatchQuerying = ref(false)
const shouldStopBatchQuery = ref(false)
const isTargetQuerying = ref(false)
const shouldStopTargetQuery = ref(false)
const showTargetVersionModal = ref(false)
const targetVersionInput = ref('')
const targetVersionInputRef = ref<any>(null)
const versionForm = ref({
  bundleId: '',
  appId: 0,
  appName: '',
  filter: ''
})

interface VersionItem {
  versionId: string
  displayVersion: string
  fileSize: string
  releaseDate: string
  isQuerying?: boolean
}

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

const versionItems = ref<VersionItem[]>([])

async function querySingleVersionMetadata(row: VersionItem) {
  row.isQuerying = true
  isAnyOperationRunning.value = true
  statusText.value = `正在查询版本 ID ${row.versionId} 的详情...`

  try {
    const res = await GetVersionMetadata(versionForm.value.bundleId, row.versionId, versionForm.value.appId)
    row.displayVersion = res.displayVersion
    row.fileSize = res.displayFileSize
    row.releaseDate = res.releaseDate ? res.releaseDate.replace(/T.*/, '') : '-'
    statusText.value = `构建 ID ${row.versionId} 对应版本: ${res.displayVersion} (${res.displayFileSize})`
  } catch (err: any) {
    message.error(`查询详情失败: ${err}`)
  } finally {
    row.isQuerying = false
    isAnyOperationRunning.value = false
  }
}

async function downloadFromVersions(row: VersionItem) {
  const appName = versionForm.value.appName || versionForm.value.bundleId
  const ver = row.displayVersion !== '未查询' ? row.displayVersion : (row.versionId ? `Build ${row.versionId}` : '最新版')
  try {
    await AddDownloadTask(
      appName,
      versionForm.value.bundleId,
      versionForm.value.appId,
      ver,
      row.versionId,
      row.fileSize
    )
    message.success(`已添加任务: ${appName} (${ver})`)
    activeTab.value = 'download'
  } catch (err: any) {
    message.error(`添加下载失败: ${err}`)
  }
}

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
  { title: '文件体积', key: 'fileSize', width: 110 },
  { title: '发布日期', key: 'releaseDate', width: 120 },
  {
    title: '操作',
    key: 'actions',
    width: 160,
    fixed: 'right' as const,
    render(row: VersionItem) {
      return h(NSpace, { size: 6, wrap: false }, () => [
        h(NButton, {
          size: 'tiny',
          secondary: true,
          disabled: isListingVersions.value || isBatchQuerying.value || isTargetQuerying.value,
          loading: row.isQuerying,
          onClick: () => querySingleVersionMetadata(row)
        }, () => '查详情'),
        h(NButton, {
          size: 'tiny',
          type: 'primary',
          onClick: () => downloadFromVersions(row)
        }, () => '下载')
      ])
    }
  }
]

// Download Tasks Manager State
const downloadTasks = ref<main.DownloadTask[]>([])
const activeTaskCount = computed(() =>
  downloadTasks.value.filter(t => t.status === 'downloading' || t.status === 'pending').length
)
const completedTaskCount = computed(() =>
  downloadTasks.value.filter(t => t.status === 'completed').length
)

// Purchased State
const isPurchasedLoading = ref(false)
const isLoadingAllPurchases = ref(false)
const purchasedPage = ref(1)
const purchasedPageSize = ref(20)
const purchasedTotal = ref(0)
const purchasedApps = ref<main.AppItem[]>([])
const purchasedSearchKeyword = ref('')

const filteredPurchasedApps = computed(() => {
  const kw = purchasedSearchKeyword.value.trim().toLowerCase()
  if (!kw) return purchasedApps.value
  return purchasedApps.value.filter(app => {
    const nameMatch = Boolean(app.name && app.name.toLowerCase().includes(kw))
    const bundleMatch = Boolean(app.bundleID && app.bundleID.toLowerCase().includes(kw))
    const idMatch = Boolean(app.id && String(app.id).includes(kw))
    return nameMatch || bundleMatch || idMatch
  })
})

// IPA Installer State
const selectedIPAPath = ref('')
const devices = ref<main.DeviceInfo[]>([])
const selectedDeviceUDID = ref<string | null>(null)
const isLoadingDevices = ref(false)
const isInstallingIPA = ref(false)
const selectedIPAFileName = computed(() => selectedIPAPath.value.split(/[\\/]/).pop() || selectedIPAPath.value)
const selectedDevice = computed(() => devices.value.find(device => device.udid === selectedDeviceUDID.value))
const deviceOptions = computed(() => devices.value.map(device => {
  const modelShort = device.productType
  const osPrefix = device.productType.toLowerCase().includes('ipad') ? 'iPadOS' : 'iOS'
  const versionStr = device.productVersion ? `${osPrefix} ${device.productVersion}` : ''
  const details = (device.name === device.productType
    ? [versionStr, device.connectionType]
    : [modelShort, versionStr, device.connectionType]
  ).filter(Boolean).join(' · ')
  return {
    label: details ? `${device.name} (${details})` : device.name,
    value: device.udid
  }
}))

function formatReadableDate(raw?: string): string {
  if (!raw) return '-'
  return raw.replace(/T/, ' ').replace(/:\d{2}Z$/, '').replace(/Z$/, '')
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
      return h(NSpace, { size: 6, wrap: false }, () => [
        h(NButton, { size: 'tiny', secondary: true, onClick: () => selectAppForVersions(row) }, () => '历史版本'),
        h(NButton, { size: 'tiny', type: 'primary', onClick: () => downloadFromPurchased(row) }, () => '下载')
      ])
    }
  }
]

// Settings State
const isTestingProxy = ref(false)

// Logs State
const showLogs = ref<boolean>(localStorage.getItem('apple_vault_show_logs') === 'true')
const logLines = ref<string[]>([])
const logContainerRef = ref<HTMLElement | null>(null)

function toggleLogs() {
  showLogs.value = !showLogs.value
  localStorage.setItem('apple_vault_show_logs', String(showLogs.value))
}

function appendLog(line: string) {
  logLines.value.push(line)
  if (logLines.value.length > 600) {
    logLines.value.shift()
  }
  nextTick(() => {
    if (logContainerRef.value) {
      logContainerRef.value.scrollTop = logContainerRef.value.scrollHeight
    }
  })
}

function clearLogs() {
  logLines.value = []
}

// --- Methods ---

async function loadSettings() {
  try {
    const s = await GetSettings()
    settings.value = s
  } catch (err: any) {
    console.error(err)
  }
}

async function onProxyToggle(val: boolean) {
  settings.value.enableProxy = val
  await SaveSettings(settings.value)
  message.info(val ? '已开启网络代理' : '已关闭网络代理')
}

async function refreshAccount(autoNavigate = false) {
  isAccountLoading.value = true
  statusText.value = '正在获取账号信息...'
  try {
    const res = await GetAccountInfo()
    account.value = res
    if (res.success && res.email) {
      statusText.value = `已登录: ${res.name} (${res.email})`
      await loadSettings()
      if (autoNavigate) {
        activeTab.value = 'search'
      }
    } else {
      statusText.value = '未检测到已登录的 Apple ID'
      activeTab.value = 'account'
    }
  } catch (err: any) {
    account.value = { name: '', email: '', success: false }
    statusText.value = '未登录'
    activeTab.value = 'account'
  } finally {
    isAccountLoading.value = false
  }
}

async function handleLogin() {
  if (!loginForm.value.email || !loginForm.value.password) {
    message.warning('请输入 Apple ID 邮箱和密码！')
    return
  }

  isLoggingIn.value = true
  isAnyOperationRunning.value = true
  statusText.value = '正在验证 Apple ID 账号与密码...'

  try {
    const res = await Login(loginForm.value.email, loginForm.value.password, '')
    if (res.requires2FA) {
      twoFACode.value = ''
      show2FAModal.value = true
      statusText.value = '等待输入双重认证验证码...'
      nextTick(() => {
        twoFAInputRef.value?.focus()
      })
      return
    }

    if (res.success) {
      account.value = res.account
      message.success(`登录成功: ${res.account.name}`)
      loginForm.value.password = ''
      statusText.value = `登录成功: ${res.account.name}`
      await loadSettings()
      activeTab.value = 'search'
    } else {
      message.error(`登录失败: ${res.errorMessage}`)
      statusText.value = `登录失败: ${res.errorMessage}`
    }
  } catch (err: any) {
    message.error(`执行异常: ${err}`)
  } finally {
    isLoggingIn.value = false
    isAnyOperationRunning.value = false
  }
}

async function confirm2FA() {
  if (twoFACode.value.length !== 6) {
    message.warning('请输入正确的 6 位数字验证码！')
    return
  }

  isLoggingIn.value = true
  isAnyOperationRunning.value = true
  statusText.value = '正在提交验证码并完成登录...'

  try {
    const res = await Login(loginForm.value.email, loginForm.value.password, twoFACode.value)
    if (res.success) {
      show2FAModal.value = false
      account.value = res.account
      message.success(`登录成功: ${res.account.name}`)
      loginForm.value.password = ''
      twoFACode.value = ''
      statusText.value = `登录成功: ${res.account.name}`
      await loadSettings()
      activeTab.value = 'search'
    } else {
      message.error(`验证失败: ${res.errorMessage}`)
      statusText.value = `验证失败: ${res.errorMessage}`
    }
  } catch (err: any) {
    message.error(`验证异常: ${err}`)
  } finally {
    isLoggingIn.value = false
    isAnyOperationRunning.value = false
  }
}

function cancel2FA() {
  show2FAModal.value = false
  twoFACode.value = ''
  statusText.value = '用户取消了验证。'
}

async function handleRevoke() {
  dialog.warning({
    title: '确认退出登录',
    content: '确定要退出当前 Apple ID 账号并清除本地授权凭证吗？',
    positiveText: '确定退出',
    negativeText: '取消',
    onPositiveClick: async () => {
      isRevoking.value = true
      isAnyOperationRunning.value = true
      try {
        const ok = await Revoke()
        if (ok) {
          account.value = { name: '', email: '', success: false }
          message.success('已成功注销登录凭据')
          statusText.value = '已退出登录'
          await loadSettings()
          activeTab.value = 'account'
        }
      } catch (err: any) {
        message.error(`注销失败: ${err}`)
      } finally {
        isRevoking.value = false
        isAnyOperationRunning.value = false
      }
    }
  })
}

async function handleClearKeychain() {
  dialog.warning({
    title: '清空本地密钥库缓存',
    content: '此操作将删除本地存储的 .ipatool 密钥数据并重置缓存。确定清理吗？',
    positiveText: '确定清理',
    negativeText: '取消',
    onPositiveClick: async () => {
      isClearing.value = true
      try {
        await ClearKeychainCache()
        message.success('本地密钥缓存已清除')
        await refreshAccount()
      } catch (err: any) {
        message.error(`清理失败: ${err}`)
      } finally {
        isClearing.value = false
      }
    }
  })
}

// Search
async function handleSearch() {
  if (!searchForm.value.term) {
    message.warning('请输入搜索关键词')
    return
  }

  isSearching.value = true
  isAnyOperationRunning.value = true
  statusText.value = `正在搜索 "${searchForm.value.term}"...`

  try {
    const res = await Search(searchForm.value.term, searchForm.value.limit, searchForm.value.platform)
    searchResults.value = res.apps || []
    statusText.value = `搜索完成，找到 ${searchResults.value.length} 个应用。`
  } catch (err: any) {
    message.error(`搜索失败: ${err}`)
    statusText.value = '搜索失败'
  } finally {
    isSearching.value = false
    isAnyOperationRunning.value = false
  }
}

function selectAppForVersions(app: main.AppItem) {
  versionForm.value.appName = app.name
  versionForm.value.bundleId = app.bundleID
  versionForm.value.appId = app.id
  versionForm.value.filter = ''
  targetVersionInput.value = ''
  activeTab.value = 'versions'
  handleListVersions()
}

async function downloadFromSearch(app: main.AppItem) {
  try {
    await AddDownloadTask(
      app.name,
      app.bundleID,
      app.id,
      app.version || '最新版',
      '',
      ''
    )
    message.success(`已添加任务「${app.name}」到下载中心`)
    activeTab.value = 'download'
  } catch (err: any) {
    message.error(`添加下载失败: ${err}`)
  }
}

async function downloadFromPurchased(app: main.AppItem) {
  try {
    await AddDownloadTask(
      app.name,
      app.bundleID,
      app.id,
      app.version || '最新版',
      '',
      ''
    )
    message.success(`已添加任务「${app.name}」到下载中心`)
    activeTab.value = 'download'
  } catch (err: any) {
    message.error(`添加下载失败: ${err}`)
  }
}

async function handlePurchaseApp(bundleId: string) {
  isAnyOperationRunning.value = true
  statusText.value = `正在获取应用授权: ${bundleId}...`
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
    isAnyOperationRunning.value = false
  }
}

// Versions
async function handleListVersions() {
  if (!versionForm.value.bundleId) {
    message.warning('请输入 Bundle ID')
    return
  }

  isListingVersions.value = true
  isAnyOperationRunning.value = true
  statusText.value = `正在查询 ${versionForm.value.bundleId} 的历史版本列表...`

  try {
    const res = await ListVersions(versionForm.value.bundleId, versionForm.value.appId)
    versionItems.value = (res.externalVersionIdentifiers || []).map(id => ({
      versionId: id,
      displayVersion: '未查询',
      fileSize: '-',
      releaseDate: '-'
    }))
    statusText.value = `共获取到 ${versionItems.value.length} 个历史版本记录。`
  } catch (err: any) {
    message.error(`查询历史版本失败: ${err}`)
    statusText.value = '查询历史版本失败'
  } finally {
    isListingVersions.value = false
    isAnyOperationRunning.value = false
  }
}

async function cancellableSleep(ms: number, stopRef: { value: boolean }): Promise<boolean> {
  const start = Date.now()
  while (Date.now() - start < ms) {
    if (stopRef.value) return false
    await new Promise(r => setTimeout(r, 20))
  }
  return !stopRef.value
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
    return item.displayVersion
  } catch (err) {
    console.warn(`查询构建 ID ${item.versionId} 失败:`, err)
    return ''
  } finally {
    item.isQuerying = false
  }
}

async function runBatchQuery(count: number) {
  isBatchQuerying.value = true
  shouldStopBatchQuery.value = false
  isAnyOperationRunning.value = true
  statusText.value = `正在批量查询前 ${count} 个历史版本号...`

  try {
    for (let i = 0; i < count; i++) {
      if (shouldStopBatchQuery.value) {
        statusText.value = '批量查询已手动停止。'
        message.info('批量查询已停止')
        break
      }

      const item = versionItems.value[i]
      if (item.displayVersion !== '未查询') continue

      statusText.value = `正在查询 (${i + 1}/${count}): ${item.versionId}...`
      await fetchItemMetadata(i)

      if (shouldStopBatchQuery.value) break
      if (!await cancellableSleep(120, shouldStopBatchQuery)) break
    }

    if (!shouldStopBatchQuery.value) {
      statusText.value = '批量查询完成。'
      message.success('批量查询完成')
    }
  } catch (err: any) {
    statusText.value = `批量查询失败: ${err}`
    message.error(`批量查询失败: ${err}`)
  } finally {
    isBatchQuerying.value = false
    shouldStopBatchQuery.value = false
    isAnyOperationRunning.value = false
  }
}

function handleBatchQueryClick() {
  if (isBatchQuerying.value) {
    stopBatchQuery()
    return
  }
  handleBatchQuery()
}

function stopBatchQuery() {
  shouldStopBatchQuery.value = true
  statusText.value = '正在停止批量查询...'
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
  nextTick(() => {
    targetVersionInputRef.value?.focus()
  })
}

function stopTargetQuery() {
  shouldStopTargetQuery.value = true
  statusText.value = '正在停止版本检索...'
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

async function runTargetQuery(targetVersion: string) {
  isTargetQuerying.value = true
  shouldStopTargetQuery.value = false
  isAnyOperationRunning.value = true
  statusText.value = `正在检索目标版本: ${targetVersion}...`

  const total = versionItems.value.length
  let found = false
  let foundItem: VersionItem | null = null

  try {
    let low = 0
    let high = total - 1

    statusText.value = `正在核对最新版本 (1/${total})...`
    const v0 = await fetchItemMetadata(0)
    if (shouldStopTargetQuery.value) {
      statusText.value = `已停止查询指定版本 (${targetVersion})。`
      return
    }

    if (v0 && isVersionMatch(v0, targetVersion)) {
      found = true
      foundItem = versionItems.value[0]
    } else if (v0 && compareVersions(v0, targetVersion) < 0) {
      statusText.value = `当前最新版本 (${v0}) 小于目标版本 ${targetVersion}，未找到该版本。`
      message.warning(`当前最新版本为 ${v0}，未找到更高版本 ${targetVersion}`)
      return
    } else {
      low = 1
    }

    if (!found && high >= low) {
      statusText.value = `正在核对早期版本 (${total}/${total})...`
      const vLast = await fetchItemMetadata(high)
      if (shouldStopTargetQuery.value) {
        statusText.value = `已停止查询指定版本 (${targetVersion})。`
        return
      }

      if (vLast && isVersionMatch(vLast, targetVersion)) {
        found = true
        foundItem = versionItems.value[high]
      } else if (vLast && compareVersions(vLast, targetVersion) > 0) {
        statusText.value = `当前早期版本 (${vLast}) 大于目标版本 ${targetVersion}，未找到该版本。`
        message.warning(`当前早期版本为 ${vLast}，未找到更低版本 ${targetVersion}`)
        return
      } else {
        high = high - 1
      }
    }

    if (!found) {
      while (low <= high && !shouldStopTargetQuery.value) {
        const mid = Math.floor((low + high) / 2)
        statusText.value = `正在核对版本 (构建 ID: ${versionItems.value[mid].versionId})...`
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
      statusText.value = `已停止查询指定版本 (${targetVersion})。`
      message.info('已停止查询指定版本')
    } else if (found && foundItem) {
      statusText.value = `已找到目标版本 ${foundItem.displayVersion} (构建 ID: ${foundItem.versionId})！`
      message.success(`已查询到指定版本 ${foundItem.displayVersion}！`)
      versionForm.value.filter = targetVersion
    } else {
      statusText.value = `检索完成，未在历史记录中找到指定版本: ${targetVersion}`
      message.warning(`未找到版本号: ${targetVersion}`)
    }
  } catch (err: any) {
    statusText.value = `查询指定版本失败: ${err}`
    message.error(`查询失败: ${err}`)
  } finally {
    isTargetQuerying.value = false
    shouldStopTargetQuery.value = false
    isAnyOperationRunning.value = false
  }
}

// Download Tasks Operations
async function loadDownloadTasks() {
  try {
    const list = await GetDownloadTasks()
    downloadTasks.value = list || []
  } catch (err: any) {
    console.error('加载任务失败:', err)
  }
}

function getTaskProgressStatus(status: string) {
  if (status === 'completed') return 'success'
  if (status === 'error') return 'error'
  return 'info'
}

function getTaskStatusText(status: string) {
  switch (status) {
    case 'pending': return '等待中'
    case 'downloading': return '下载中'
    case 'completed': return '已完成'
    case 'error': return '下载失败'
    case 'canceled': return '已取消'
    default: return status
  }
}

function formatTaskBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

async function handleClearCompleted() {
  try {
    await ClearCompletedDownloadTasks()
    await loadDownloadTasks()
    message.success('已清空所有已完成任务')
  } catch (err: any) {
    message.error(`操作失败: ${err}`)
  }
}

function handleOpenDefaultDownloadDir() {
  OpenInExplorer(settings.value.defaultDownloadDir || '')
}

async function handleCancelTask(id: string) {
  try {
    await CancelDownloadTask(id)
    message.info('正在取消任务...')
  } catch (err: any) {
    message.error(`取消失败: ${err}`)
  }
}

async function handleDeleteTask(id: string) {
  try {
    await DeleteDownloadTask(id)
    await loadDownloadTasks()
  } catch (err: any) {
    message.error(`删除失败: ${err}`)
  }
}

function handleInstallFromTask(outputPath: string) {
  if (!outputPath) {
    message.warning('未找到下载的 IPA 文件路径')
    return
  }
  selectedIPAPath.value = outputPath
  activeTab.value = 'installer'
  statusText.value = `已选定安装包: ${outputPath}`
  if (devices.value.length === 0 && !isLoadingDevices.value) {
    void loadConnectedDevices()
  }
}

async function handleRetryTask(task: main.DownloadTask) {
  try {
    await AddDownloadTask(
      task.appName,
      task.bundleID,
      task.appId,
      task.version,
      task.versionId,
      task.fileSize
    )
    message.success(`已重新添加任务「${task.appName}」`)
  } catch (err: any) {
    message.error(`重试失败: ${err}`)
  }
}

// IPA Installer
function useIPAPath(paths: string[]) {
  const ipaPath = paths.find(path => path.toLowerCase().endsWith('.ipa'))
  if (!ipaPath) {
    message.warning('请拖放 .ipa 格式的安装包')
    return
  }
  selectedIPAPath.value = ipaPath
  activeTab.value = 'installer'
  statusText.value = `已选定 IPA: ${ipaPath}`
}

async function handleSelectIPA() {
  try {
    const path = await SelectIPA()
    if (path) {
      useIPAPath([path])
    }
  } catch (err: any) {
    message.error(`选择 IPA 失败: ${err}`)
  }
}

async function loadConnectedDevices() {
  isLoadingDevices.value = true
  isAnyOperationRunning.value = true
  statusText.value = '正在检测已连接的苹果设备...'
  try {
    const result = await ListDevices()
    devices.value = result || []
    if (!devices.value.some(device => device.udid === selectedDeviceUDID.value)) {
      selectedDeviceUDID.value = devices.value.length === 1 ? devices.value[0].udid : null
    }
    if (devices.value.length === 0) {
      statusText.value = '未发现可用苹果设备'
      message.warning('未发现设备，请确认设备已连接、解锁并信任此电脑')
    } else {
      statusText.value = `发现 ${devices.value.length} 台可用设备`
    }
  } catch (err: any) {
    devices.value = []
    selectedDeviceUDID.value = null
    statusText.value = '设备检测失败'
    message.error(`设备检测失败: ${err}`)
  } finally {
    isLoadingDevices.value = false
    isAnyOperationRunning.value = false
  }
}

async function handleInstallIPA() {
  if (!selectedIPAPath.value) {
    message.warning('请先选择 IPA 文件')
    return
  }
  if (!selectedDeviceUDID.value) {
    message.warning('请选择要安装的苹果设备')
    return
  }

  isInstallingIPA.value = true
  isAnyOperationRunning.value = true
  statusText.value = `正在安装 ${selectedIPAFileName.value}（请保持设备屏幕常亮勿息屏）...`
  message.info('开始安装应用，请确保设备屏幕保持常亮解锁...')
  try {
    const result = await InstallIPA(selectedIPAPath.value, selectedDeviceUDID.value)
    if (result.success) {
      statusText.value = result.message
      message.success(result.message)
    } else {
      statusText.value = 'IPA 安装失败'
      message.error(result.message || 'IPA 安装失败')
    }
  } catch (err: any) {
    statusText.value = 'IPA 安装失败'
    message.error(`IPA 安装失败: ${err}`)
  } finally {
    isInstallingIPA.value = false
    isAnyOperationRunning.value = false
  }
}

// Purchased
async function loadPurchases() {
  isPurchasedLoading.value = true
  isAnyOperationRunning.value = true
  statusText.value = `正在加载第 ${purchasedPage.value} 页已购应用列表...`

  try {
    const res = await ListPurchases(purchasedPage.value, purchasedPageSize.value)
    purchasedApps.value = res.apps || []
    purchasedTotal.value = res.totalCount || 0
    statusText.value = `已购应用加载完成 (共 ${res.totalCount} 款)。`
  } catch (err: any) {
    message.error(`加载已购列表失败: ${err}`)
    statusText.value = '加载失败'
  } finally {
    isPurchasedLoading.value = false
    isAnyOperationRunning.value = false
  }
}

async function loadAllPurchases() {
  if (isLoadingAllPurchases.value || isPurchasedLoading.value) return
  isLoadingAllPurchases.value = true
  isAnyOperationRunning.value = true
  statusText.value = `正在拉取全部已购应用 (共 ${purchasedTotal.value} 款)...`

  try {
    const limit = Math.max(purchasedTotal.value || 200, 100)
    const res = await ListPurchases(1, limit)
    if (res.apps && res.apps.length > 0) {
      purchasedApps.value = res.apps
      purchasedTotal.value = res.totalCount || res.apps.length
      purchasedPage.value = 1
      statusText.value = `已成功载入全部 ${res.apps.length} 款已购应用。`
      message.success(`已成功加载全部 ${res.apps.length} 款已购应用`)
    }
  } catch (err: any) {
    message.error(`加载全部已购失败: ${err}`)
    statusText.value = '加载失败'
  } finally {
    isLoadingAllPurchases.value = false
    isAnyOperationRunning.value = false
  }
}

function onPurchasedPageSizeChange(val: number) {
  purchasedPageSize.value = val
  purchasedPage.value = 1
  loadPurchases()
}

function prevPurchasedPage() {
  if (purchasedPage.value > 1) {
    purchasedPage.value--
    loadPurchases()
  }
}

function nextPurchasedPage() {
  if (purchasedPage.value * purchasedPageSize.value < purchasedTotal.value) {
    purchasedPage.value++
    loadPurchases()
  }
}

// Settings
async function handleSelectDefaultDir() {
  try {
    const dir = await SelectDirectory('选择默认下载保存目录', settings.value.defaultDownloadDir)
    if (dir) {
      settings.value.defaultDownloadDir = dir
    }
  } catch (err: any) {
    console.error(err)
  }
}

async function handleTestProxy() {
  isTestingProxy.value = true
  statusText.value = '正在测试网络代理连接...'
  try {
    const res = await TestProxy(settings.value.proxyUrl)
    if (res.success) {
      message.success(res.message)
      statusText.value = '网络代理测试成功！'
    } else {
      message.error(res.message)
      statusText.value = '网络代理测试失败'
    }
  } catch (err: any) {
    message.error(`测试异常: ${err}`)
  } finally {
    isTestingProxy.value = false
  }
}

async function handleSaveSettings() {
  try {
    await SaveSettings(settings.value)
    message.success('设置已保存并生效！')
  } catch (err: any) {
    message.error(`保存失败: ${err}`)
  }
}

function openGitHub() {
  BrowserOpenURL('https://github.com/wnnz/AppleVault')
}

function copyGitHubUrl() {
  navigator.clipboard.writeText('https://github.com/wnnz/AppleVault')
  message.success('已复制仓库地址到剪贴板！')
}

function handleCancel() {
  shouldStopBatchQuery.value = true
  shouldStopTargetQuery.value = true
  CancelRunningCommand()
  isAnyOperationRunning.value = false
  statusText.value = '已终止当前操作'
}

// Lifecycle
onMounted(() => {
  const urlParams = new URLSearchParams(window.location.search)
  const tabParam = urlParams.get('tab')
  const isDemo = urlParams.get('demo') === '1'

  if (isDemo) {
    account.value = { name: 'AppleVault User', email: 'user@icloud.com', success: true }
    searchResults.value = [
      { id: 414478124, bundleID: 'com.tencent.xin', name: '微信 (WeChat)', version: '8.0.50', price: 0, displayPrice: '免费' },
      { id: 333206289, bundleID: 'com.alipay.iphoneclient', name: '支付宝 - 生活好 支付宝', version: '10.5.88', price: 0, displayPrice: '免费' },
      { id: 835599320, bundleID: 'com.zhiliaoapp.musically', name: 'TikTok - Videos, Music & LIVE', version: '35.8.0', price: 0, displayPrice: '免费' },
      { id: 590338362, bundleID: 'com.netease.cloudmusic', name: '网易云音乐', version: '9.0.70', price: 0, displayPrice: '免费' },
      { id: 736536022, bundleID: 'tv.danmaku.bilianime', name: '哔哩哔哩 (Bilibili)', version: '7.82.0', price: 0, displayPrice: '免费' }
    ]
    versionForm.value.bundleId = 'com.tencent.xin'
    versionForm.value.appName = '微信'
    versionItems.value = [
      { versionId: '868192301', displayVersion: '8.0.50', fileSize: '286.4 MB', releaseDate: '2024-08-15' },
      { versionId: '867204918', displayVersion: '8.0.49', fileSize: '284.1 MB', releaseDate: '2024-07-20' },
      { versionId: '865819021', displayVersion: '8.0.48', fileSize: '279.8 MB', releaseDate: '2024-06-12' },
      { versionId: '864201990', displayVersion: '8.0.47', fileSize: '275.2 MB', releaseDate: '2024-05-08' },
      { versionId: '862901124', displayVersion: '8.0.46', fileSize: '270.5 MB', releaseDate: '2024-04-01' }
    ]
    downloadTasks.value = [
      { id: '1', appName: '微信 (WeChat)', bundleID: 'com.tencent.xin', appId: 414478124, version: '8.0.50', versionId: '868192301', fileSize: '286.4 MB', totalBytes: 300312000, currBytes: 192200000, progress: 64, speed: '8.6 MB/s', status: 'downloading', outputPath: '', errorMessage: '', createdAt: '2024-08-20 15:30:12' },
      { id: '2', appName: '支付宝', bundleID: 'com.alipay.iphoneclient', appId: 333206289, version: '10.5.88', versionId: '865001129', fileSize: '142.0 MB', totalBytes: 148897000, currBytes: 148897000, progress: 100, speed: '已完成', status: 'completed', outputPath: 'data/downloads/user@icloud.com/alipay.ipa', errorMessage: '', createdAt: '2024-08-20 15:24:05' }
    ]
    purchasedApps.value = [
      { id: 414478124, bundleID: 'com.tencent.xin', name: '微信 (WeChat)', version: '8.0.50', price: 0, purchaseDate: '2024-01-15 10:20', displayPrice: '已购' },
      { id: 333206289, bundleID: 'com.alipay.iphoneclient', name: '支付宝', version: '10.5.88', price: 0, purchaseDate: '2024-02-08 14:12', displayPrice: '已购' },
      { id: 1136220934, bundleID: 'com.firecore.infuse', name: 'Infuse • 精彩影音播放器', version: '7.7.2', price: 0, purchaseDate: '2024-03-22 09:45', displayPrice: '已购' },
      { id: 1596487405, bundleID: 'com.taguirov.adam.WebDAV', name: 'WebDAV Manager', version: '2.1.0', price: 0, purchaseDate: '2024-05-19 16:30', displayPrice: '已购' }
    ]
    purchasedTotal.value = 4
    devices.value = [
      { udid: '00008130-001A49021E28001C', name: 'iPhone 15 Pro Max', productType: 'iPhone 15 Pro Max', productVersion: '17.5.1', connectionType: 'USB' }
    ]
    selectedDeviceUDID.value = '00008130-001A49021E28001C'
    selectedIPAPath.value = 'data\\downloads\\user@icloud.com\\WeChat_8.0.50.ipa'
  } else {
    loadSettings()
    refreshAccount(true)
    loadDownloadTasks()
  }

  if (tabParam) {
    activeTab.value = tabParam
  }

  OnFileDrop((_x: number, _y: number, paths: string[]) => {
    useIPAPath(paths)
  }, true)

  EventsOn('log', (msg: string) => {
    appendLog(msg)
  })

  EventsOn('download-task-updated', (updatedTask: main.DownloadTask) => {
    const idx = downloadTasks.value.findIndex(t => t.id === updatedTask.id)
    if (idx !== -1) {
      downloadTasks.value[idx] = updatedTask
    } else {
      downloadTasks.value.unshift(updatedTask)
    }
  })

  EventsOn('download-tasks-reload', () => {
    loadDownloadTasks()
  })
})

onBeforeUnmount(() => {
  OnFileDropOff()
})
</script>

<style scoped>
/* Base Structure & Layout */
.app-layout {
  display: flex;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background-color: #f8fafc;
  color: #1e293b;
  font-family: -apple-system, BlinkMacSystemFont, "SF Pro Display", "SF Pro Text", "Helvetica Neue", "PingFang SC", "Microsoft YaHei", sans-serif;
  user-select: none;
}

.dark-mode {
  background-color: #0f172a;
  color: #f1f5f9;
}

/* 1. Sidebar Styles */
.app-sidebar {
  width: 220px;
  min-width: 220px;
  background-color: #ffffff;
  border-right: 1px solid rgba(0, 0, 0, 0.06);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 16px 12px 12px 12px;
  box-sizing: border-box;
  z-index: 10;
}

.dark-mode .app-sidebar {
  background-color: #1e293b;
  border-right-color: rgba(255, 255, 255, 0.08);
}

.sidebar-brand {
  display: flex;
  align-items: center;
  padding: 2px 6px 14px 6px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.04);
}

.dark-mode .sidebar-brand {
  border-bottom-color: rgba(255, 255, 255, 0.05);
}

.brand-icon-box {
  width: 32px;
  height: 32px;
  border-radius: 9px;
  background: linear-gradient(135deg, #0071e3 0%, #409cff 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  margin-right: 10px;
  box-shadow: 0 4px 10px rgba(0, 113, 227, 0.25);
}

.brand-apple-svg {
  width: 18px;
  height: 18px;
}

.brand-text {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 1px;
}

.brand-title-row {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.brand-name {
  font-size: 15px;
  font-weight: 700;
  letter-spacing: -0.2px;
  color: #0f172a;
  line-height: 1.2;
}

.dark-mode .brand-name {
  color: #f8fafc;
}

.brand-sub {
  font-size: 11px;
  font-weight: 500;
  color: #64748b;
  letter-spacing: 0.2px;
  line-height: 1;
}

.dark-mode .brand-sub {
  color: #94a3b8;
}

.brand-tag {
  font-size: 10px;
  font-weight: 600;
  color: #0071e3;
  background: rgba(0, 113, 227, 0.08);
  padding: 1px 5px;
  border-radius: 4px;
}

.dark-mode .brand-tag {
  background: rgba(10, 132, 255, 0.2);
  color: #38bdf8;
}

/* Nav Menu */
.sidebar-nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding-top: 12px;
  overflow-y: auto;
}

.nav-section-title {
  font-size: 11px;
  font-weight: 600;
  color: #94a3b8;
  padding: 6px 10px 4px 10px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.nav-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 8px 12px;
  margin-bottom: 3px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: #475569;
  font-size: 13.5px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
  text-align: left;
}

.dark-mode .nav-item {
  color: #94a3b8;
}

.nav-item:hover {
  background-color: rgba(0, 0, 0, 0.04);
  color: #0f172a;
}

.dark-mode .nav-item:hover {
  background-color: rgba(255, 255, 255, 0.06);
  color: #f8fafc;
}

.nav-item.active {
  background-color: #0071e3;
  color: #ffffff;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(0, 113, 227, 0.25);
}

.dark-mode .nav-item.active {
  background-color: #0284c7;
  color: #ffffff;
}

.nav-label {
  flex: 1;
}

.nav-badge {
  background: #ff3b30;
  color: white;
  font-size: 10px;
  font-weight: 700;
  border-radius: 10px;
  padding: 1px 6px;
  min-width: 14px;
  text-align: center;
}

.account-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.dot-online {
  background-color: #34c759;
  box-shadow: 0 0 5px rgba(52, 199, 89, 0.5);
}

.dot-offline {
  background-color: #cbd5e1;
}

/* Sidebar Footer */
.sidebar-footer {
  padding-top: 10px;
  border-top: 1px solid rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.dark-mode .sidebar-footer {
  border-top-color: rgba(255, 255, 255, 0.06);
}

.sidebar-account-card {
  display: flex;
  align-items: center;
  padding: 8px 10px;
  border-radius: 8px;
  background-color: #f1f5f9;
  cursor: pointer;
  transition: background 0.15s ease;
}

.dark-mode .sidebar-account-card {
  background-color: #0f172a;
}

.sidebar-account-card:hover {
  background-color: #e2e8f0;
}

.dark-mode .sidebar-account-card:hover {
  background-color: #334155;
}

.user-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #cbd5e1;
  color: #475569;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 13px;
  margin-right: 8px;
}

.user-avatar.avatar-logged {
  background: linear-gradient(135deg, #0071e3 0%, #00c6ff 100%);
  color: #ffffff;
}

.user-info-text {
  flex: 1;
  overflow: hidden;
}

.user-name {
  font-size: 12px;
  font-weight: 600;
  color: #1e293b;
}

.dark-mode .user-name {
  color: #e2e8f0;
}

.user-email {
  font-size: 10.5px;
  color: #64748b;
}

.dark-mode .user-email {
  color: #94a3b8;
}

.sidebar-actions-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 4px;
}

.theme-switcher-wrapper {
  margin-left: auto;
  display: flex;
  align-items: center;
}

.quick-tool {
  display: flex;
  align-items: center;
  gap: 6px;
}

.quick-tool-label {
  font-size: 12px;
  color: #64748b;
  font-weight: 500;
}

.icon-action-btn {
  width: 28px;
  height: 28px;
  border-radius: 7px;
  border: 1px solid #e2e8f0;
  background: #ffffff;
  color: #64748b;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s ease;
  padding: 0;
  box-sizing: border-box;
}

.dark-mode .icon-action-btn {
  background: #0f172a;
  border-color: #334155;
  color: #94a3b8;
}

.icon-action-btn:hover {
  border-color: #0071e3;
  color: #0071e3;
  background: #f8fafc;
}

.dark-mode .icon-action-btn:hover {
  border-color: #0284c7;
  color: #38bdf8;
  background: #1e293b;
}

.icon-action-btn.active {
  background: rgba(0, 113, 227, 0.12);
  border-color: #0071e3;
  color: #0071e3;
}

.dark-mode .icon-action-btn.active {
  background: rgba(2, 132, 199, 0.2);
  border-color: #0284c7;
  color: #38bdf8;
}

/* 2. Main Workspace Layout */
.app-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  position: relative;
}

.view-content-wrapper {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 18px 20px 8px 20px;
  box-sizing: border-box;
}

.view-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.view-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.view-title {
  margin: 0 0 2px 0;
  font-size: 19px;
  font-weight: 700;
  letter-spacing: -0.4px;
  color: #0f172a;
}

.dark-mode .view-title {
  color: #f8fafc;
}

.view-desc {
  margin: 0;
  font-size: 12px;
  color: #64748b;
}

.dark-mode .view-desc {
  color: #94a3b8;
}

/* Clean Apple-style Card */
.clean-card {
  background: #ffffff;
  border: 1px solid rgba(0, 0, 0, 0.05);
  border-radius: 12px;
  padding: 14px 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.02), 0 2px 8px rgba(0, 0, 0, 0.02);
  box-sizing: border-box;
}

.dark-mode .clean-card {
  background: #1e293b;
  border-color: rgba(255, 255, 255, 0.06);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.table-flex-card {
  flex: 1;
  overflow: hidden;
  padding: 6px;
  display: flex;
  flex-direction: column;
}

/* Uniform Height Standards */
.height-aligned {
  height: 36px !important;
  box-sizing: border-box !important;
  line-height: 34px !important;
}

.height-aligned-btn {
  height: 36px !important;
  box-sizing: border-box !important;
  padding: 0 16px !important;
  font-size: 13px !important;
}

.small-aligned-btn {
  height: 30px !important;
  box-sizing: border-box !important;
  padding: 0 12px !important;
  font-size: 12px !important;
}

/* Clean Input Elements */
.clean-input {
  border: 1px solid #e2e8f0;
  background-color: #f8fafc;
  color: #0f172a;
  border-radius: 8px;
  padding: 0 12px;
  font-size: 13px;
  outline: none;
  transition: all 0.15s ease;
  box-sizing: border-box;
}

.dark-mode .clean-input {
  border-color: #334155;
  background-color: #0f172a;
  color: #f1f5f9;
}

.clean-input:focus {
  border-color: #0071e3;
  background-color: #ffffff;
  box-shadow: 0 0 0 2px rgba(0, 113, 227, 0.12);
}

.dark-mode .clean-input:focus {
  background-color: #1e293b;
  border-color: #0284c7;
  box-shadow: 0 0 0 2px rgba(2, 132, 199, 0.2);
}

/* Search View Specifics */
.search-input-group {
  display: flex;
  align-items: center;
  gap: 12px;
}

.search-input-wrapper {
  flex: 1;
  display: flex;
  align-items: center;
}

.search-input {
  width: 100%;
  border-radius: 8px;
}

.filter-controls {
  display: flex;
  align-items: center;
  gap: 10px;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 36px;
}

.filter-label {
  font-size: 12.5px;
  color: #64748b;
  font-weight: 500;
  white-space: nowrap;
}

/* Versions View Specifics */
.versions-toolbar {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.bundle-input-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.clean-input-prefix-box {
  display: flex;
  align-items: center;
  flex: 1;
  border: 1px solid #e2e8f0;
  background-color: #f8fafc;
  border-radius: 8px;
  padding: 0 8px;
}

.dark-mode .clean-input-prefix-box {
  border-color: #334155;
  background-color: #0f172a;
}

.prefix-label {
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
  padding-right: 8px;
  border-right: 1px solid #e2e8f0;
  white-space: nowrap;
}

.dark-mode .prefix-label {
  border-right-color: #334155;
  color: #94a3b8;
}

.border-none {
  border: none !important;
  background: transparent !important;
  box-shadow: none !important;
}

.filter-search-row {
  display: flex;
  align-items: center;
}

.filter-search-box {
  width: 100%;
}

.filter-input {
  width: 100%;
  font-size: 12.5px;
}

.count-pill {
  font-size: 12px;
  font-weight: 500;
  background: rgba(0, 113, 227, 0.08);
  color: #0071e3;
  padding: 3px 9px;
  border-radius: 20px;
}

.dark-mode .count-pill {
  background: rgba(2, 132, 199, 0.2);
  color: #38bdf8;
}

.tag-version-highlight {
  font-weight: 600;
  color: #107c41;
  background: rgba(16, 124, 65, 0.08);
  padding: 2px 6px;
  border-radius: 4px;
}

.dark-mode .tag-version-highlight {
  color: #4ade80;
  background: rgba(74, 222, 128, 0.12);
}

/* Download View Tasks */
.tasks-container {
  flex: 1;
  overflow-y: auto;
  padding-right: 4px;
}

.empty-state-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
  background: #ffffff;
  border-radius: 12px;
  border: 1px dashed #cbd5e1;
}

.dark-mode .empty-state-card {
  background: #1e293b;
  border-color: #334155;
}

.empty-title {
  font-size: 15px;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 6px;
}

.dark-mode .empty-title {
  color: #f1f5f9;
}

.empty-desc {
  font-size: 12.5px;
  color: #64748b;
}

.task-grid {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.modern-task-card {
  background: #ffffff;
  border-radius: 12px;
  border: 1px solid rgba(0, 0, 0, 0.05);
  padding: 12px 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.02);
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.dark-mode .modern-task-card {
  background: #1e293b;
  border-color: rgba(255, 255, 255, 0.06);
}

.task-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.task-app-title-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.task-name {
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
}

.dark-mode .task-name {
  color: #f8fafc;
}

.task-pill {
  font-size: 11px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 4px;
}

.task-pill-ver {
  background: rgba(52, 199, 89, 0.12);
  color: #28a745;
}

.task-pill-build {
  background: #f1f5f9;
  color: #64748b;
}

.dark-mode .task-pill-build {
  background: #334155;
  color: #94a3b8;
}

.task-actions-group {
  display: flex;
  gap: 6px;
}

.task-meta-bundle {
  font-size: 11px;
  color: #94a3b8;
  font-family: Menlo, Monaco, Consolas, monospace;
}

.task-progress-section {
  margin: 2px 0;
}

.task-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 11.5px;
}

.footer-status-tag {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-indicator {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.indicator-downloading {
  background-color: #0071e3;
  box-shadow: 0 0 5px rgba(0, 113, 227, 0.6);
}

.indicator-completed {
  background-color: #34c759;
}

.indicator-error {
  background-color: #ff3b30;
}

.indicator-canceled {
  background-color: #94a3b8;
}

.indicator-pending {
  background-color: #f59e0b;
}

.task-speed {
  font-weight: 600;
  color: #0071e3;
}

.dark-mode .task-speed {
  color: #38bdf8;
}

.task-bytes-info {
  color: #64748b;
}

.task-error-text {
  color: #ff3b30;
  max-width: 360px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.footer-right-info {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.pct-text {
  font-size: 12.5px;
  color: #0071e3;
}

.dark-mode .pct-text {
  color: #38bdf8;
}

.time-text {
  font-size: 11px;
  color: #94a3b8;
}

/* Purchased View Controls & Search Toolbar */
.purchased-search-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 10px 14px;
  flex-wrap: wrap;
}

.purchased-search-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 280px;
}

.purchased-search-input-box {
  display: flex;
  align-items: center;
  background: #f8fafc;
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 8px;
  padding: 0 10px;
  height: 32px;
  width: 330px;
  max-width: 100%;
  transition: all 0.2s ease;
}

.dark-mode .purchased-search-input-box {
  background: #0f172a;
  border-color: rgba(255, 255, 255, 0.1);
}

.purchased-search-input-box:focus-within {
  border-color: #0071e3;
  box-shadow: 0 0 0 2px rgba(0, 113, 227, 0.15);
  background: #ffffff;
}

.dark-mode .purchased-search-input-box:focus-within {
  border-color: #38bdf8;
  box-shadow: 0 0 0 2px rgba(56, 189, 248, 0.2);
  background: #1e293b;
}

.purchased-search-input {
  border: none;
  background: transparent;
  outline: none;
  font-size: 12.5px;
  padding: 0 8px;
  color: #1e293b;
  width: 100%;
}

.dark-mode .purchased-search-input {
  color: #f1f5f9;
}

.search-icon {
  color: #94a3b8;
  flex-shrink: 0;
}

.search-clear-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 2px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  border-radius: 50%;
  transition: all 0.15s;
}

.search-clear-btn:hover {
  color: #64748b;
  background: rgba(0, 0, 0, 0.06);
}

.dark-mode .search-clear-btn:hover {
  color: #cbd5e1;
  background: rgba(255, 255, 255, 0.1);
}

.purchased-search-badge {
  font-size: 12px;
  color: #64748b;
  user-select: none;
  white-space: nowrap;
}

.dark-mode .purchased-search-badge {
  color: #94a3b8;
}

.purchased-search-badge.has-filter {
  color: #0071e3;
}

.dark-mode .purchased-search-badge.has-filter {
  color: #38bdf8;
}

.purchased-page-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-size-selector {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-right: 4px;
}

.size-label {
  font-size: 12px;
  color: #64748b;
  white-space: nowrap;
}

.dark-mode .size-label {
  color: #94a3b8;
}

.page-indicator {
  font-size: 12px;
  color: #64748b;
  margin: 0 4px;
  white-space: nowrap;
}

.dark-mode .page-indicator {
  color: #94a3b8;
}

.purchased-header-actions {
  display: flex;
  align-items: center;
}

.purchased-empty-box {
  padding: 45px 20px;
  display: flex;
  justify-content: center;
  align-items: center;
  text-align: center;
}

.search-none-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  color: #64748b;
}

.dark-mode .search-none-state {
  color: #94a3b8;
}

.search-none-icon {
  font-size: 26px;
  margin-bottom: 4px;
  opacity: 0.85;
}

.search-none-title {
  font-size: 14px;
  font-weight: 600;
  color: #1e293b;
}

.dark-mode .search-none-title {
  color: #f1f5f9;
}

.search-none-desc {
  font-size: 12px;
  color: #94a3b8;
}

/* 手绘模式专属适配 */
.handdrawn-mode:not(.dark-mode) .purchased-search-input-box {
  background: #ffffff;
  border: 1px solid #cbe0f0;
  border-radius: 18px;
  box-shadow: 0 2px 8px rgba(162, 198, 224, 0.18);
}

.handdrawn-mode:not(.dark-mode) .purchased-search-input-box:focus-within {
  border-color: #5baaf5;
  box-shadow: 0 0 0 3px rgba(91, 170, 245, 0.22);
}

.handdrawn-mode.dark-mode .purchased-search-input-box {
  background: #141d2a;
  border: 1px solid rgba(91, 170, 245, 0.25);
  border-radius: 18px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.handdrawn-mode.dark-mode .purchased-search-input-box:focus-within {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.25);
}

.handdrawn-mode:not(.dark-mode) .purchased-search-badge.has-filter {
  color: #2563eb;
  font-weight: 600;
}

.handdrawn-mode.dark-mode .purchased-search-badge.has-filter {
  color: #60a5fa;
  font-weight: 600;
}

/* Installer View */
.installer-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
  flex: 1;
  overflow: hidden;
}

.flex-col {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.card-headline {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.headline-title {
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
}

.dark-mode .headline-title {
  color: #f8fafc;
}

.headline-badge {
  font-size: 11px;
  font-weight: 600;
  color: #34c759;
  background: rgba(52, 199, 89, 0.12);
  padding: 2px 7px;
  border-radius: 4px;
}

.clean-drop-zone {
  flex: 1;
  min-height: 130px;
  border: 2px dashed #cbd5e1;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  background: #f8fafc;
  transition: all 0.2s ease;
  padding: 16px;
  box-sizing: border-box;
}

.dark-mode .clean-drop-zone {
  border-color: #334155;
  background: #0f172a;
}

.clean-drop-zone:hover {
  border-color: #0071e3;
  background: rgba(0, 113, 227, 0.03);
}

.clean-drop-zone.drop-active {
  border-color: #34c759;
  background: rgba(52, 199, 89, 0.04);
}

.drop-primary-title {
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 4px;
  text-align: center;
}

.dark-mode .drop-primary-title {
  color: #f1f5f9;
}

.drop-secondary-path {
  font-size: 11px;
  color: #94a3b8;
  max-width: 90%;
  text-align: center;
}

.device-spec-box {
  background: #f8fafc;
  border-radius: 8px;
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.dark-mode .device-spec-box {
  background: #0f172a;
}

.spec-row {
  display: flex;
  justify-content: space-between;
  font-size: 11.5px;
}

.spec-k {
  color: #64748b;
}

.spec-v {
  color: #0f172a;
}

.dark-mode .spec-v {
  color: #f1f5f9;
}

.no-device-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 24px 10px;
  text-align: center;
}

.no-device-text {
  font-size: 12.5px;
  font-weight: 600;
  color: #475569;
}

.no-device-sub {
  font-size: 11px;
  color: #94a3b8;
  margin-top: 4px;
}

.sub-alert-box {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  background: rgba(0, 113, 227, 0.06);
  border-radius: 8px;
  padding: 8px 12px;
  font-size: 11.5px;
  color: #0369a1;
  line-height: 1.5;
}

.dark-mode .sub-alert-box {
  background: rgba(2, 132, 199, 0.15);
  color: #7dd3fc;
}

.installer-action-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  background: #ffffff;
}

.dark-mode .installer-action-banner {
  background: #1e293b;
}

.action-banner-title {
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
  margin-bottom: 2px;
}

.dark-mode .action-banner-title {
  color: #f8fafc;
}

.action-banner-desc {
  font-size: 11.5px;
  color: #64748b;
}

.install-submit-btn {
  height: 40px;
  padding: 0 24px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
}

/* Account Center View */
.two-columns-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.account-profile-box {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px 0 6px 0;
}

.large-avatar {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: #cbd5e1;
  color: #475569;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  font-weight: 700;
}

.large-avatar.avatar-active {
  background: linear-gradient(135deg, #0071e3 0%, #00c6ff 100%);
  color: #ffffff;
  box-shadow: 0 4px 12px rgba(0, 113, 227, 0.3);
}

.profile-name {
  font-size: 16px;
  font-weight: 700;
  color: #0f172a;
  margin-bottom: 2px;
}

.dark-mode .profile-name {
  color: #f8fafc;
}

.profile-email {
  font-size: 12.5px;
  color: #64748b;
}

.account-card-actions {
  display: flex;
  gap: 8px;
}

.form-item-clean {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.clean-label {
  font-size: 12px;
  font-weight: 600;
  color: #475569;
}

.dark-mode .clean-label {
  color: #94a3b8;
}

.account-pill {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 12px;
}

.pill-success {
  background: rgba(52, 199, 89, 0.12);
  color: #28a745;
}

.pill-gray {
  background: #f1f5f9;
  color: #64748b;
}

.dark-mode .pill-gray {
  background: #334155;
  color: #94a3b8;
}

/* Settings View */
.settings-stack {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-height: 100%;
  overflow-y: auto;
}

.settings-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-bottom: 14px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
}

.dark-mode .settings-group {
  border-bottom-color: rgba(255, 255, 255, 0.06);
}

.settings-group:last-child {
  border-bottom: none;
}

.settings-group-title {
  font-size: 12px;
  font-weight: 700;
  color: #0071e3;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.dark-mode .settings-group-title {
  color: #38bdf8;
}

.settings-field-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 44px;
  padding: 4px 0;
  gap: 20px;
}

.field-meta {
  flex: 1;
}

.field-control {
  width: 360px;
  min-width: 360px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

.setting-input-action-group {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 8px;
}

.setting-fixed-btn {
  width: 68px !important;
  min-width: 68px !important;
  padding: 0 !important;
  text-align: center;
}

.w-full {
  width: 100% !important;
}

.field-title {
  font-size: 13px;
  font-weight: 600;
  color: #0f172a;
}

.dark-mode .field-title {
  color: #f8fafc;
}

.field-desc {
  font-size: 11.5px;
  color: #64748b;
  margin-top: 1px;
}

.engine-badge-box {
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(52, 199, 89, 0.12);
  color: #28a745;
  padding: 3px 9px;
  border-radius: 20px;
  font-size: 11.5px;
  font-weight: 600;
}

.engine-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #34c759;
}

/* About View */
.about-card-stack {
  display: flex;
  flex-direction: column;
  gap: 18px;
  max-height: 100%;
  overflow-y: auto;
  padding: 20px 24px;
}

.about-hero {
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 20px 24px 22px 24px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
}

.dark-mode .about-hero {
  border-bottom-color: rgba(255, 255, 255, 0.06);
}

.about-large-icon {
  width: 68px;
  height: 68px;
  border-radius: 16px;
  background: linear-gradient(135deg, #0071e3 0%, #409cff 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  box-shadow: 0 6px 16px rgba(0, 113, 227, 0.25);
  flex-shrink: 0;
  margin-left: 6px;
}

.about-hero-text {
  flex: 1;
}

.about-app-title {
  font-size: 20px;
  font-weight: 700;
  color: #0f172a;
  letter-spacing: -0.4px;
}

.dark-mode .about-app-title {
  color: #f8fafc;
}

.about-version-line {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 4px 0 6px 0;
}

.about-version-badge {
  font-size: 11px;
  font-weight: 600;
  color: #0071e3;
  background: rgba(0, 113, 227, 0.08);
  padding: 2px 7px;
  border-radius: 6px;
}

.dark-mode .about-version-badge {
  background: rgba(2, 132, 199, 0.2);
  color: #38bdf8;
}

.about-badge-sub {
  font-size: 11.5px;
  color: #94a3b8;
}

.about-intro {
  margin: 0;
  font-size: 12.5px;
  color: #64748b;
  line-height: 1.5;
}

.dark-mode .about-intro {
  color: #94a3b8;
}

.about-section-box {
  background: #f8fafc;
  border-radius: 10px;
  padding: 12px 14px;
}

.dark-mode .about-section-box {
  background: #0f172a;
}

.about-section-label {
  font-size: 12px;
  font-weight: 600;
  color: #475569;
  margin-bottom: 8px;
}

.dark-mode .about-section-label {
  color: #94a3b8;
}

.repo-link-bar {
  display: flex;
  align-items: center;
  gap: 10px;
}

.about-features-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.about-feature-item {
  background: #f8fafc;
  border-radius: 10px;
  padding: 12px 14px;
}

.dark-mode .about-feature-item {
  background: #0f172a;
}

.feature-title {
  font-size: 13px;
  font-weight: 700;
  color: #0f172a;
  margin-bottom: 3px;
}

.dark-mode .feature-title {
  color: #f8fafc;
}

.feature-desc {
  font-size: 11.5px;
  color: #64748b;
  line-height: 1.5;
}

.dark-mode .feature-desc {
  color: #94a3b8;
}

.about-footer-notice {
  font-size: 11.5px;
  color: #94a3b8;
  line-height: 1.5;
  border-top: 1px solid rgba(0, 0, 0, 0.05);
  padding-top: 10px;
}

.dark-mode .about-footer-notice {
  border-top-color: rgba(255, 255, 255, 0.06);
}

/* Bottom Status Bar */
.bottom-status-bar {
  height: 30px;
  min-height: 30px;
  background-color: #ffffff;
  border-top: 1px solid rgba(0, 0, 0, 0.06);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  box-sizing: border-box;
  font-size: 11.5px;
  flex-shrink: 0;
}

.dark-mode .bottom-status-bar {
  background-color: #1e293b;
  border-top-color: rgba(255, 255, 255, 0.08);
}

.status-indicator-section {
  display: flex;
  align-items: center;
  gap: 8px;
  overflow: hidden;
  max-width: 70%;
}

.status-pulse {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: #34c759;
}

.status-pulse.pulse-busy {
  background-color: #0071e3;
  animation: pulse-ring 1.5s infinite;
}

@keyframes pulse-ring {
  0% {
    transform: scale(0.95);
    box-shadow: 0 0 0 0 rgba(0, 113, 227, 0.7);
  }
  70% {
    transform: scale(1);
    box-shadow: 0 0 0 5px rgba(0, 113, 227, 0);
  }
  100% {
    transform: scale(0.95);
    box-shadow: 0 0 0 0 rgba(0, 113, 227, 0);
  }
}

.status-summary-text {
  color: #64748b;
  font-size: 11px;
}

.dark-mode .status-summary-text {
  color: #94a3b8;
}

.status-bar-tools {
  display: flex;
  align-items: center;
}

.status-btn {
  background: transparent;
  border: none;
  font-size: 11px;
  color: #64748b;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
}

.status-btn:hover {
  background: #f1f5f9;
  color: #0f172a;
}

.dark-mode .status-btn:hover {
  background: #334155;
  color: #ffffff;
}

/* Console Terminal Drawer */
.sleek-console-drawer {
  height: 150px;
  background: #090d16;
  border-top: 1px solid #1e293b;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.console-action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 14px;
  background: #111827;
  border-bottom: 1px solid #1f2937;
}

.console-title {
  font-size: 11px;
  font-weight: 600;
  color: #94a3b8;
  letter-spacing: 0.5px;
}

.console-btns {
  display: flex;
  gap: 6px;
}

.console-bar-btn {
  background: transparent;
  border: none;
  color: #64748b;
  font-size: 11px;
  cursor: pointer;
  padding: 1px 6px;
  border-radius: 3px;
}

.console-bar-btn:hover {
  background: #1f2937;
  color: #f1f5f9;
}

.console-scroll-screen {
  flex: 1;
  overflow-y: auto;
  padding: 8px 14px;
  font-family: Menlo, Monaco, Consolas, monospace;
  font-size: 11.5px;
  line-height: 1.6;
}

.console-empty-tip {
  color: #475569;
  font-style: italic;
}

.console-row {
  color: #cbd5e1;
  word-break: break-all;
}

.console-row.row-err {
  color: #f87171;
}

/* Utilities */
.text-ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.flex-align-center {
  display: flex;
  align-items: center;
}

.flex-1 {
  flex: 1;
}

.font-bold {
  font-weight: 700;
}

.italic {
  font-style: italic;
}

.text-dim {
  color: #94a3b8;
}

.mt-2 { margin-top: 8px; }
.mt-3 { margin-top: 12px; }
.mt-4 { margin-top: 16px; }
.mb-2 { margin-bottom: 8px; }
.mb-3 { margin-bottom: 12px; }
.mb-4 { margin-bottom: 16px; }
.mr-2 { margin-right: 8px; }
.ml-2 { margin-left: 8px; }

/* ==========================================================================
   Handdrawn Theme Styles (清新水彩插画手账风格)
   ========================================================================== */

/* 手绘模式字体栈：温润柔和 */
.handdrawn-mode {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
}

/* 1. 手绘 - 浅色模式 (Watercolor Illustration Light) - 多层不均匀手绘水彩晕染 */
.handdrawn-mode:not(.dark-mode) {
  background-color: #f6fafc;
  background-image:
    /* 水彩天蓝晕染 - 右上方大水渍斑块 */
    radial-gradient(ellipse at 88% 12%, rgba(186, 230, 253, 0.65) 0%, rgba(224, 242, 254, 0.25) 36%, transparent 65%),
    /* 柔和薰衣草淡紫水彩晕染 - 右下角层次 */
    radial-gradient(ellipse at 85% 82%, rgba(237, 233, 254, 0.55) 0%, rgba(243, 232, 255, 0.18) 35%, transparent 60%),
    /* 薄荷浅绿水彩微斑 - 左下方点缀 */
    radial-gradient(ellipse at 18% 85%, rgba(204, 251, 241, 0.45) 0%, rgba(224, 242, 254, 0.15) 32%, transparent 58%),
    /* 浅蓝手绘水渍晕染 - 左上方边缘 */
    radial-gradient(ellipse at 16% 22%, rgba(191, 219, 254, 0.5) 0%, rgba(224, 242, 254, 0.15) 30%, transparent 55%),
    /* 中心温润画纸柔光 */
    radial-gradient(ellipse at 50% 45%, rgba(254, 249, 195, 0.28) 0%, transparent 50%),
    /* 极细微水彩纸质纹理微粒 */
    radial-gradient(rgba(125, 178, 224, 0.15) 1px, transparent 1px);
  background-size: 100% 100%, 100% 100%, 100% 100%, 100% 100%, 100% 100%, 20px 20px;
  background-attachment: fixed;
  color: #243b53;
}

/* 侧边栏：多层次不均匀手绘水彩水渍质感 */
.handdrawn-mode:not(.dark-mode) .app-sidebar {
  background: 
    radial-gradient(circle at 85% 10%, rgba(255, 255, 255, 0.75) 0%, transparent 50%),
    radial-gradient(circle at 15% 42%, rgba(186, 230, 253, 0.65) 0%, transparent 60%),
    radial-gradient(circle at 80% 86%, rgba(224, 231, 255, 0.45) 0%, transparent 55%),
    linear-gradient(175deg, #d3e8f8 0%, #dcedf9 45%, #e7f3fb 100%);
  border-right: 1px solid rgba(162, 203, 233, 0.65);
  box-shadow: 2px 0 16px rgba(162, 198, 224, 0.16);
}

.handdrawn-mode:not(.dark-mode) .sidebar-brand {
  border-bottom: 1px solid rgba(162, 203, 233, 0.6);
}

.handdrawn-mode:not(.dark-mode) .brand-icon-box {
  background: linear-gradient(135deg, #5baaf5 0%, #3b82f6 100%);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.28);
}

.handdrawn-mode:not(.dark-mode) .brand-name {
  color: #1c385c;
  font-weight: 700;
}

.handdrawn-mode:not(.dark-mode) .brand-sub {
  color: #527599;
  font-weight: 500;
}

.handdrawn-mode:not(.dark-mode) .brand-tag {
  border: 1px solid #bfdbfe;
  background: #e1f0fc;
  color: #2563eb;
  border-radius: 6px;
  font-weight: 600;
}

.handdrawn-mode:not(.dark-mode) .nav-section-title {
  color: #627d98;
  font-weight: 600;
}

/* 导航项：参考图风格，激活项为天蓝圆润卡片+柔光微阴影 */
.handdrawn-mode:not(.dark-mode) .nav-item {
  color: #3b5a7a;
  font-weight: 600;
  border: none;
  border-radius: 10px;
  transition: all 0.18s ease;
}

.handdrawn-mode:not(.dark-mode) .nav-item:hover {
  background: rgba(255, 255, 255, 0.65);
  color: #1e3a5f;
}

.handdrawn-mode:not(.dark-mode) .nav-item.active {
  background: linear-gradient(135deg, #5baaf5 0%, #3b82f6 100%);
  color: #ffffff;
  border: none;
  box-shadow: 0 4px 14px rgba(59, 130, 246, 0.35);
  border-radius: 10px;
  font-weight: 600;
}

.handdrawn-mode:not(.dark-mode) .sidebar-footer {
  border-top: 1px solid rgba(162, 203, 233, 0.6);
}

.handdrawn-mode:not(.dark-mode) .sidebar-account-card {
  background: rgba(255, 255, 255, 0.88);
  border: 1px solid rgba(186, 215, 237, 0.8);
  box-shadow: 0 4px 14px rgba(162, 198, 224, 0.25);
  border-radius: 12px;
}

.handdrawn-mode:not(.dark-mode) .sidebar-account-card:hover {
  background: #ffffff;
  box-shadow: 0 6px 18px rgba(147, 197, 235, 0.35);
  transform: translateY(-1px);
}

.handdrawn-mode:not(.dark-mode) .user-name {
  color: #1e3a5f;
  font-weight: 700;
}

.handdrawn-mode:not(.dark-mode) .user-email {
  color: #627d98;
}

.handdrawn-mode:not(.dark-mode) .quick-tool-label {
  color: #3b5a7a;
  font-weight: 600;
}

.handdrawn-mode:not(.dark-mode) .icon-action-btn {
  border: 1px solid #cbe0f0;
  background: #ffffff;
  color: #3b5a7a;
  box-shadow: 0 2px 8px rgba(162, 198, 224, 0.22);
  border-radius: 8px;
}

.handdrawn-mode:not(.dark-mode) .icon-action-btn:hover {
  background: #f0f7fe;
  color: #2563eb;
  border-color: #93c5fd;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.25);
  transform: translateY(-1px);
}

.handdrawn-mode:not(.dark-mode) .icon-action-btn:active {
  transform: translateY(1px);
  box-shadow: 0 1px 3px rgba(59, 130, 246, 0.2);
}

.handdrawn-mode:not(.dark-mode) .icon-action-btn.active {
  background: #e1f0fc;
  border-color: #60a5fa;
  color: #2563eb;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.25);
}

/* 标题与描述：清秀明亮 */
.handdrawn-mode:not(.dark-mode) .view-title {
  color: #1c385c;
  font-weight: 700;
  letter-spacing: -0.2px;
}

.handdrawn-mode:not(.dark-mode) .view-desc {
  color: #627d98;
}

/* 主工作区卡片：纯白纸感，柔和蓝光弥散阴影，细致水彩勾线 */
.handdrawn-mode:not(.dark-mode) .clean-card,
.handdrawn-mode:not(.dark-mode) .settings-group,
.handdrawn-mode:not(.dark-mode) .modern-task-card,
.handdrawn-mode:not(.dark-mode) .empty-state-card,
.handdrawn-mode:not(.dark-mode) .device-spec-box,
.handdrawn-mode:not(.dark-mode) .installer-action-banner,
.handdrawn-mode:not(.dark-mode) .about-section-box,
.handdrawn-mode:not(.dark-mode) .sub-alert-box {
  background: #ffffff;
  border: 1px solid #dcebf6;
  box-shadow: 0 6px 20px rgba(162, 198, 224, 0.2), 0 1px 3px rgba(162, 198, 224, 0.1);
  border-radius: 14px;
  transition: all 0.2s ease;
}

.handdrawn-mode:not(.dark-mode) .about-hero {
  background: rgba(240, 247, 255, 0.7);
  border: 1px solid #cbe0f0;
  border-radius: 16px;
  padding: 22px 28px;
  box-shadow: 0 4px 16px rgba(162, 198, 224, 0.2);
  gap: 24px;
}

.handdrawn-mode:not(.dark-mode) .about-large-icon {
  background: linear-gradient(135deg, #5baaf5 0%, #3b82f6 100%);
  box-shadow: 0 6px 18px rgba(59, 130, 246, 0.35);
  margin-left: 6px;
}

.handdrawn-mode:not(.dark-mode) .clean-card:hover,
.handdrawn-mode:not(.dark-mode) .modern-task-card:hover {
  box-shadow: 0 8px 24px rgba(147, 197, 235, 0.32);
}

/* 搜索框：参考图风格的药丸形胶囊圆角，纯白加细边 */
.handdrawn-mode:not(.dark-mode) .clean-input-box {
  background: #ffffff;
  border: 1px solid #cbe0f0;
  box-shadow: 0 2px 8px rgba(162, 198, 224, 0.18);
  border-radius: 20px;
}

.handdrawn-mode:not(.dark-mode) .clean-input-box:focus-within {
  border-color: #5baaf5;
  box-shadow: 0 0 0 3px rgba(91, 170, 245, 0.22);
}

.handdrawn-mode:not(.dark-mode) .clean-input {
  color: #1e3a5f;
}

.handdrawn-mode:not(.dark-mode) .clean-input-prefix-box {
  background: #eef6fc;
  border-right: 1px solid #cbe0f0;
  color: #2b496d;
  font-weight: 600;
  border-top-left-radius: 20px;
  border-bottom-left-radius: 20px;
}

.handdrawn-mode:not(.dark-mode) .prefix-label {
  color: #2b496d;
  font-weight: 600;
}

/* 标签胶囊：清秀马卡龙水彩色系 */
.handdrawn-mode:not(.dark-mode) .count-pill {
  border: 1px solid #bae6fd;
  background: #e0f2fe;
  color: #0284c7;
  font-weight: 600;
  border-radius: 6px;
}

.handdrawn-mode:not(.dark-mode) .tag-version-highlight {
  border: 1px solid #fde68a;
  background: #fef3c7;
  color: #b45309;
  font-weight: 600;
  border-radius: 6px;
}

.handdrawn-mode:not(.dark-mode) .task-pill-build {
  border: 1px solid #bbf7d0;
  background: #dcfce7;
  color: #15803d;
  font-weight: 600;
  border-radius: 6px;
}

.handdrawn-mode:not(.dark-mode) .pill-gray {
  border: 1px solid #e2e8f0;
  background: #f1f5f9;
  color: #475569;
  font-weight: 600;
  border-radius: 6px;
}

/* 拖拽区域 */
.handdrawn-mode:not(.dark-mode) .clean-drop-zone {
  background: #f6fafe;
  border: 2px dashed #93c5fd;
  border-radius: 16px;
  box-shadow: 0 4px 16px rgba(162, 198, 224, 0.15);
}

.handdrawn-mode:not(.dark-mode) .clean-drop-zone:hover {
  border-color: #3b82f6;
  background: #eef7ff;
}

/* 底部状态栏 */
.handdrawn-mode:not(.dark-mode) .bottom-status-bar {
  background: #e6f2fa;
  border-top: 1px solid rgba(162, 203, 233, 0.7);
  color: #334e68;
}

.handdrawn-mode:not(.dark-mode) .status-summary-text {
  color: #334e68;
  font-weight: 600;
}

.handdrawn-mode:not(.dark-mode) .status-btn {
  border: 1px solid #cbe0f0;
  background: #ffffff;
  color: #334e68;
  box-shadow: 0 2px 6px rgba(162, 198, 224, 0.18);
  border-radius: 6px;
}

.handdrawn-mode:not(.dark-mode) .status-btn:hover {
  background: #dbeafe;
  border-color: #93c5fd;
  color: #1d4ed8;
}

/* 2. 手绘 - 深色模式 (Midnight Watercolor / 午夜星空水彩) - 不均匀星云水彩晕染 */
.handdrawn-mode.dark-mode {
  background-color: #0e1624;
  background-image:
    /* 星空深蓝水彩晕染 - 左上斑块 */
    radial-gradient(ellipse at 16% 20%, rgba(30, 64, 175, 0.42) 0%, rgba(30, 58, 138, 0.18) 38%, transparent 65%),
    /* 星云紫水彩晕染 - 右上方大斑块 */
    radial-gradient(ellipse at 86% 16%, rgba(91, 33, 182, 0.35) 0%, rgba(67, 56, 202, 0.15) 36%, transparent 62%),
    /* 极光深青水彩微光 - 右下斑块 */
    radial-gradient(ellipse at 82% 84%, rgba(13, 148, 136, 0.3) 0%, rgba(15, 118, 110, 0.12) 35%, transparent 58%),
    /* 墨蓝深水晕染 - 左下斑块 */
    radial-gradient(ellipse at 22% 82%, rgba(29, 78, 216, 0.32) 0%, rgba(30, 58, 138, 0.12) 34%, transparent 55%),
    /* 星尘微粒点缀 */
    radial-gradient(rgba(147, 197, 253, 0.22) 1.2px, transparent 1.2px);
  background-size: 100% 100%, 100% 100%, 100% 100%, 100% 100%, 28px 28px;
  background-attachment: fixed;
  color: #e2e8f0;
}

.handdrawn-mode.dark-mode .app-sidebar {
  background: 
    radial-gradient(circle at 85% 12%, rgba(59, 130, 246, 0.2) 0%, transparent 60%),
    radial-gradient(circle at 15% 75%, rgba(124, 58, 237, 0.16) 0%, transparent 55%),
    linear-gradient(180deg, #0c131d 0%, #101926 50%, #142132 100%);
  border-right: 1px solid rgba(75, 115, 160, 0.28);
}

.handdrawn-mode.dark-mode .sidebar-brand {
  border-bottom: 1px solid rgba(75, 115, 160, 0.25);
}

.handdrawn-mode.dark-mode .brand-icon-box {
  background: linear-gradient(135deg, #2563eb 0%, #3b82f6 100%);
  box-shadow: 0 4px 14px rgba(37, 99, 235, 0.35);
}

.handdrawn-mode.dark-mode .brand-name {
  color: #f1f5f9;
  font-weight: 700;
}

.handdrawn-mode.dark-mode .brand-sub {
  color: #94a3b8;
  font-weight: 500;
}

.handdrawn-mode.dark-mode .brand-tag {
  border: 1px solid #1e3a8a;
  background: rgba(30, 58, 138, 0.4);
  color: #93c5fd;
  border-radius: 6px;
  font-weight: 600;
}

.handdrawn-mode.dark-mode .nav-section-title {
  color: #829ab1;
  font-weight: 600;
}

.handdrawn-mode.dark-mode .nav-item {
  color: #94a3b8;
  font-weight: 600;
  border: none;
  border-radius: 10px;
}

.handdrawn-mode.dark-mode .nav-item:hover {
  background: rgba(33, 49, 71, 0.7);
  color: #f1f5f9;
}

.handdrawn-mode.dark-mode .nav-item.active {
  background: linear-gradient(135deg, #2563eb 0%, #3b82f6 100%);
  color: #ffffff;
  border: none;
  box-shadow: 0 4px 14px rgba(37, 99, 235, 0.4);
  border-radius: 10px;
  font-weight: 600;
}

.handdrawn-mode.dark-mode .sidebar-footer {
  border-top: 1px solid rgba(75, 115, 160, 0.25);
}

.handdrawn-mode.dark-mode .sidebar-account-card {
  background: #1b2636;
  border: 1px solid rgba(91, 170, 245, 0.18);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
  border-radius: 12px;
}

.handdrawn-mode.dark-mode .sidebar-account-card:hover {
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.4);
  transform: translateY(-1px);
}

.handdrawn-mode.dark-mode .user-name {
  color: #f1f5f9;
  font-weight: 700;
}

.handdrawn-mode.dark-mode .user-email {
  color: #829ab1;
}

.handdrawn-mode.dark-mode .quick-tool-label {
  color: #cbd5e1;
  font-weight: 600;
}

.handdrawn-mode.dark-mode .icon-action-btn {
  border: 1px solid rgba(91, 170, 245, 0.22);
  background: #1b2636;
  color: #cbd5e1;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  border-radius: 8px;
}

.handdrawn-mode.dark-mode .icon-action-btn:hover {
  background: #233145;
  color: #60a5fa;
  border-color: #3b82f6;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
  transform: translateY(-1px);
}

.handdrawn-mode.dark-mode .icon-action-btn:active {
  transform: translateY(1px);
}

.handdrawn-mode.dark-mode .icon-action-btn.active {
  background: rgba(59, 130, 246, 0.25);
  border-color: #3b82f6;
  color: #93c5fd;
}

.handdrawn-mode.dark-mode .view-title {
  color: #f1f5f9;
  font-weight: 700;
}

.handdrawn-mode.dark-mode .view-desc {
  color: #829ab1;
}

.handdrawn-mode.dark-mode .clean-card,
.handdrawn-mode.dark-mode .settings-group,
.handdrawn-mode.dark-mode .modern-task-card,
.handdrawn-mode.dark-mode .empty-state-card,
.handdrawn-mode.dark-mode .device-spec-box,
.handdrawn-mode.dark-mode .installer-action-banner,
.handdrawn-mode.dark-mode .about-section-box,
.handdrawn-mode.dark-mode .sub-alert-box {
  background: #1b2636;
  border: 1px solid rgba(91, 170, 245, 0.18);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.35);
  border-radius: 14px;
}

.handdrawn-mode.dark-mode .about-hero {
  background: rgba(27, 38, 54, 0.7);
  border: 1px solid rgba(91, 170, 245, 0.22);
  border-radius: 16px;
  padding: 22px 28px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.28);
  gap: 24px;
}

.handdrawn-mode.dark-mode .about-large-icon {
  background: linear-gradient(135deg, #2563eb 0%, #3b82f6 100%);
  box-shadow: 0 6px 18px rgba(37, 99, 235, 0.4);
  margin-left: 6px;
}

.handdrawn-mode.dark-mode .clean-input-box {
  background: #141d2a;
  border: 1px solid rgba(91, 170, 245, 0.25);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  border-radius: 20px;
}

.handdrawn-mode.dark-mode .clean-input-box:focus-within {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.25);
}

.handdrawn-mode.dark-mode .clean-input {
  color: #f1f5f9;
}

.handdrawn-mode.dark-mode .clean-input-prefix-box {
  background: #101824;
  border-right: 1px solid rgba(91, 170, 245, 0.25);
  color: #93c5fd;
  font-weight: 600;
  border-top-left-radius: 20px;
  border-bottom-left-radius: 20px;
}

.handdrawn-mode.dark-mode .prefix-label {
  color: #93c5fd;
  font-weight: 600;
}

.handdrawn-mode.dark-mode .count-pill {
  border: 1px solid rgba(2, 132, 199, 0.4);
  background: rgba(2, 132, 199, 0.2);
  color: #7dd3fc;
  font-weight: 600;
  border-radius: 6px;
}

.handdrawn-mode.dark-mode .tag-version-highlight {
  border: 1px solid rgba(217, 119, 6, 0.4);
  background: rgba(217, 119, 6, 0.2);
  color: #fcd34d;
  font-weight: 600;
  border-radius: 6px;
}

.handdrawn-mode.dark-mode .task-pill-build {
  border: 1px solid rgba(22, 163, 74, 0.4);
  background: rgba(22, 163, 74, 0.2);
  color: #86efac;
  font-weight: 600;
  border-radius: 6px;
}

.handdrawn-mode.dark-mode .pill-gray {
  border: 1px solid rgba(148, 163, 184, 0.2);
  background: #243245;
  color: #cbd5e1;
  font-weight: 600;
  border-radius: 6px;
}

.handdrawn-mode.dark-mode .clean-drop-zone {
  background: #16202e;
  border: 2px dashed #3b82f6;
  border-radius: 16px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
}

.handdrawn-mode.dark-mode .clean-drop-zone:hover {
  border-color: #60a5fa;
  background: #1c293b;
}

.handdrawn-mode.dark-mode .bottom-status-bar {
  background: #0f1722;
  border-top: 1px solid rgba(75, 115, 160, 0.25);
  color: #cbd5e1;
}

.handdrawn-mode.dark-mode .status-summary-text {
  color: #cbd5e1;
  font-weight: 600;
}

.handdrawn-mode.dark-mode .status-btn {
  border: 1px solid rgba(91, 170, 245, 0.22);
  background: #1b2636;
  color: #cbd5e1;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.3);
  border-radius: 6px;
}

.handdrawn-mode.dark-mode .status-btn:hover {
  background: #233145;
  border-color: #3b82f6;
  color: #93c5fd;
}

/* 3. 手绘模式下的 Naive UI 深度穿透样式 */
.handdrawn-mode:not(.dark-mode) :deep(.n-button--primary-type) {
  background: linear-gradient(135deg, #5baaf5 0%, #3b82f6 100%) !important;
  border: none !important;
  color: #ffffff !important;
  border-radius: 9px !important;
  box-shadow: 0 3px 10px rgba(59, 130, 246, 0.3);
  font-weight: 600;
}

.handdrawn-mode:not(.dark-mode) :deep(.n-button--primary-type:hover) {
  background: linear-gradient(135deg, #6bb6fb 0%, #488ef7 100%) !important;
  box-shadow: 0 4px 14px rgba(59, 130, 246, 0.4);
  transform: translateY(-1px);
}

.handdrawn-mode:not(.dark-mode) :deep(.n-button:not(.n-button--primary-type)) {
  border: 1px solid #cfe2f2 !important;
  border-radius: 9px !important;
  box-shadow: 0 2px 6px rgba(162, 198, 224, 0.15);
  font-weight: 500;
}

.handdrawn-mode.dark-mode :deep(.n-button--primary-type) {
  background: linear-gradient(135deg, #2563eb 0%, #3b82f6 100%) !important;
  border: none !important;
  color: #ffffff !important;
  border-radius: 9px !important;
  box-shadow: 0 3px 10px rgba(37, 99, 235, 0.35);
  font-weight: 600;
}

.handdrawn-mode.dark-mode :deep(.n-button:not(.n-button--primary-type)) {
  border: 1px solid rgba(91, 170, 245, 0.22) !important;
  border-radius: 9px !important;
  font-weight: 500;
}

.handdrawn-mode:not(.dark-mode) :deep(.n-data-table) {
  border: 1px solid #dcebf6;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(162, 198, 224, 0.15);
  overflow: hidden;
}

.handdrawn-mode.dark-mode :deep(.n-data-table) {
  border: 1px solid rgba(75, 115, 160, 0.25);
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
  overflow: hidden;
}

.handdrawn-mode:not(.dark-mode) :deep(.n-select .n-base-selection) {
  border: 1px solid #cbe0f0 !important;
  box-shadow: 0 2px 6px rgba(162, 198, 224, 0.12);
  border-radius: 9px;
}

.handdrawn-mode.dark-mode :deep(.n-select .n-base-selection) {
  border: 1px solid rgba(91, 170, 245, 0.25) !important;
  border-radius: 9px;
}

/* ==========================================================================
   Handdrawn Cute Illustrations & Stickers (手绘精美小插图)
   ========================================================================== */

/* 1. 侧边栏底部手绘小插画贴纸 */
.handdrawn-sidebar-sticker {
  display: flex;
  justify-content: center;
  align-items: center;
  margin: 4px 0 10px 0;
  pointer-events: none;
  user-select: none;
}

.sidebar-sticker-svg {
  overflow: visible;
  filter: drop-shadow(0 3px 6px rgba(162, 198, 224, 0.25));
}

.handdrawn-mode:not(.dark-mode) .doodle-apple-body {
  fill: #ff6b6b;
  color: #c92a2a;
}

.handdrawn-mode:not(.dark-mode) .doodle-leaf {
  fill: #51cf66;
  color: #2b8a3e;
}

.handdrawn-mode:not(.dark-mode) .doodle-ground-wash {
  fill: rgba(186, 230, 253, 0.5);
}

.handdrawn-mode:not(.dark-mode) .doodle-star {
  fill: #fcc419;
  color: #e67700;
  animation: starTwinkle 3s ease-in-out infinite;
}

.handdrawn-mode:not(.dark-mode) .doodle-sparkle {
  fill: #74c0fc;
}

/* 深色模式下的侧边栏小插图 */
.handdrawn-mode.dark-mode .sidebar-sticker-svg {
  filter: drop-shadow(0 4px 10px rgba(0, 0, 0, 0.4));
}

.handdrawn-mode.dark-mode .doodle-apple-body {
  fill: #f87171;
  color: #ef4444;
}

.handdrawn-mode.dark-mode .doodle-leaf {
  fill: #4ade80;
  color: #22c55e;
}

.handdrawn-mode.dark-mode .doodle-ground-wash {
  fill: rgba(30, 58, 138, 0.45);
}

.handdrawn-mode.dark-mode .doodle-star {
  fill: #fde047;
  color: #eab308;
  animation: starTwinkle 3s ease-in-out infinite;
}

.handdrawn-mode.dark-mode .doodle-sparkle {
  fill: #93c5fd;
}

/* 2. 主区域右上角浮动治愈小插图 */
.handdrawn-top-doodle {
  position: absolute;
  top: 14px;
  right: 24px;
  pointer-events: none;
  user-select: none;
  z-index: 2;
  animation: floatCloud 4.5s ease-in-out infinite;
}

@keyframes floatCloud {
  0%, 100% {
    transform: translateY(0px) rotate(0deg);
  }
  50% {
    transform: translateY(-5px) rotate(1.2deg);
  }
}

@keyframes starTwinkle {
  0%, 100% {
    transform: scale(1);
    opacity: 0.95;
  }
  50% {
    transform: scale(1.18);
    opacity: 1;
  }
}

.handdrawn-mode:not(.dark-mode) .doodle-cloud-body {
  fill: rgba(255, 255, 255, 0.92);
  color: #60a5fa;
  filter: drop-shadow(0 4px 12px rgba(162, 198, 224, 0.3));
}

.handdrawn-mode:not(.dark-mode) .doodle-blush {
  fill: #ff8787;
  opacity: 0.65;
}

.handdrawn-mode.dark-mode .doodle-cloud-body {
  fill: rgba(30, 41, 59, 0.88);
  color: #93c5fd;
  filter: drop-shadow(0 4px 14px rgba(0, 0, 0, 0.4));
}

.handdrawn-mode.dark-mode .doodle-blush {
  fill: #f472b6;
  opacity: 0.7;
}

/* 3. 任务空状态手绘插图 */
.handdrawn-empty-doodle {
  display: flex;
  justify-content: center;
  margin-bottom: 12px;
}

.handdrawn-mode:not(.dark-mode) .doodle-paper {
  fill: #ffffff;
  color: #60a5fa;
  filter: drop-shadow(0 3px 8px rgba(162, 198, 224, 0.25));
}

.handdrawn-mode:not(.dark-mode) .doodle-check {
  color: #10b981;
}

.handdrawn-mode:not(.dark-mode) .doodle-pencil {
  fill: #fde047;
  color: #eab308;
}

.handdrawn-mode:not(.dark-mode) .doodle-pencil-tip {
  fill: #475569;
  color: #334155;
}

.handdrawn-mode.dark-mode .doodle-paper {
  fill: #1e293b;
  color: #93c5fd;
  filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.4));
}

.handdrawn-mode.dark-mode .doodle-check {
  color: #34d399;
}

.handdrawn-mode.dark-mode .doodle-pencil {
  fill: #facc15;
  color: #ca8a04;
}

.handdrawn-mode.dark-mode .doodle-pencil-tip {
  fill: #94a3b8;
  color: #64748b;
}

/* 4. IPA拖拽区域手绘插图 */
.handdrawn-drop-doodle {
  display: flex;
  justify-content: center;
  margin-bottom: 8px;
}

.handdrawn-mode:not(.dark-mode) .doodle-box-top {
  fill: #e0f2fe;
  color: #38bdf8;
}

.handdrawn-mode:not(.dark-mode) .doodle-box-side {
  fill: #bae6fd;
  color: #0284c7;
}

.handdrawn-mode.dark-mode .doodle-box-top {
  fill: #1e3a8a;
  color: #60a5fa;
}

.handdrawn-mode.dark-mode .doodle-box-side {
  fill: #172554;
  color: #3b82f6;
}

/* ==========================================================================
   Handdrawn Rich Elements: Washi Tape, Doodle Icons, Badges & Empty Illustrations
   ========================================================================== */

/* 1. 侧边栏导航手绘图标与动效 */
.nav-doodle-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  margin-right: 8px;
  flex-shrink: 0;
  transition: transform 0.22s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.nav-item:hover .nav-doodle-icon {
  transform: scale(1.18) rotate(5deg);
}

.nav-item.active .nav-doodle-icon {
  transform: scale(1.12) rotate(-3deg);
  animation: doodleWiggle 2.5s ease-in-out infinite alternate;
}

@keyframes doodleWiggle {
  0% { transform: scale(1.12) rotate(-3deg); }
  50% { transform: scale(1.15) rotate(3deg); }
  100% { transform: scale(1.12) rotate(-3deg); }
}

.nav-title-sparkle {
  font-size: 11px;
  margin-left: 5px;
  color: #f59e0b;
  display: inline-block;
  animation: starTwinkle 2.5s ease-in-out infinite;
}

.brand-handdrawn-crown {
  font-size: 13px;
  margin-left: 3px;
  display: inline-block;
  transform: rotate(12deg);
  filter: drop-shadow(0 2px 4px rgba(245, 158, 11, 0.3));
}

/* 2. 标题马克笔划线与手绘印章贴纸 */
.title-with-badge {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.handdrawn-view-badge {
  display: inline-flex;
  align-items: center;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 9px;
  border-radius: 8px;
  letter-spacing: 0.2px;
  user-select: none;
  transform: rotate(-1.5deg);
  transition: transform 0.2s ease;
}

.handdrawn-view-badge:hover {
  transform: rotate(1.2deg) scale(1.05);
}

/* 浅色模式手绘印章贴纸 */
.handdrawn-mode:not(.dark-mode) .handdrawn-view-badge {
  background: linear-gradient(135deg, rgba(254, 240, 138, 0.8) 0%, rgba(253, 230, 138, 0.6) 100%);
  color: #854d0e;
  border: 1.5px dashed rgba(202, 138, 4, 0.5);
  box-shadow: 0 2px 6px rgba(217, 119, 6, 0.15);
}

/* 深色模式手绘印章贴纸 (荧光粉笔黑板风) */
.handdrawn-mode.dark-mode .handdrawn-view-badge {
  background: rgba(30, 58, 138, 0.5);
  color: #93c5fd;
  border: 1.5px dashed rgba(96, 165, 250, 0.6);
  box-shadow: 0 0 10px rgba(59, 130, 246, 0.25);
}

/* 标题手绘马克笔划线底纹 */
.handdrawn-mode .view-title {
  position: relative;
  display: inline-block;
  z-index: 1;
}

.handdrawn-mode:not(.dark-mode) .view-title::after {
  content: "";
  position: absolute;
  left: -4px;
  bottom: 1px;
  width: 106%;
  height: 9px;
  background: rgba(253, 224, 71, 0.45);
  border-radius: 4px;
  z-index: -1;
  transform: rotate(-0.8deg);
}

.handdrawn-mode.dark-mode .view-title::after {
  content: "";
  position: absolute;
  left: -4px;
  bottom: 1px;
  width: 106%;
  height: 9px;
  background: rgba(56, 189, 248, 0.28);
  border-radius: 4px;
  z-index: -1;
  transform: rotate(-0.8deg);
}

/* 3. 卡片顶部和纸胶带 (Washi Tape) 手账贴纸效果 */
.handdrawn-mode:not(.dark-mode) .clean-card,
.handdrawn-mode:not(.dark-mode) .settings-group,
.handdrawn-mode:not(.dark-mode) .installer-action-banner,
.handdrawn-mode:not(.dark-mode) .about-section-box {
  position: relative;
}

.handdrawn-mode:not(.dark-mode) .clean-card::before,
.handdrawn-mode:not(.dark-mode) .settings-group::before,
.handdrawn-mode:not(.dark-mode) .installer-action-banner::before,
.handdrawn-mode:not(.dark-mode) .about-section-box::before {
  content: "";
  position: absolute;
  top: -7px;
  left: 50%;
  transform: translateX(-50%) rotate(-1deg);
  width: 68px;
  height: 14px;
  background: rgba(254, 240, 138, 0.65);
  border-left: 2px dashed rgba(234, 179, 8, 0.45);
  border-right: 2px dashed rgba(234, 179, 8, 0.45);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
  border-radius: 2px;
  pointer-events: none;
  z-index: 3;
}

/* 深色模式下的透明极光微发光胶带 */
.handdrawn-mode.dark-mode .clean-card,
.handdrawn-mode.dark-mode .settings-group,
.handdrawn-mode.dark-mode .installer-action-banner,
.handdrawn-mode.dark-mode .about-section-box {
  position: relative;
}

.handdrawn-mode.dark-mode .clean-card::before,
.handdrawn-mode.dark-mode .settings-group::before,
.handdrawn-mode.dark-mode .installer-action-banner::before,
.handdrawn-mode.dark-mode .about-section-box::before {
  content: "";
  position: absolute;
  top: -7px;
  left: 50%;
  transform: translateX(-50%) rotate(-1deg);
  width: 68px;
  height: 14px;
  background: rgba(56, 189, 248, 0.22);
  border-left: 2px dashed rgba(125, 211, 252, 0.5);
  border-right: 2px dashed rgba(125, 211, 252, 0.5);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
  border-radius: 2px;
  pointer-events: none;
  z-index: 3;
}

/* 4. 表格空状态手绘插画与排版 */
.table-empty-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 36px 20px;
  text-align: center;
}

.handdrawn-table-empty-doodle {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: 14px;
  filter: drop-shadow(0 3px 8px rgba(162, 198, 224, 0.25));
}

.handdrawn-mode.dark-mode .handdrawn-table-empty-doodle {
  filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.45));
}

.empty-state-title {
  font-size: 14px;
  font-weight: 700;
  margin-bottom: 6px;
  color: #1e3a5f;
}

.handdrawn-mode.dark-mode .empty-state-title {
  color: #e2e8f0;
}

.empty-state-desc {
  font-size: 12px;
  color: #627d98;
  max-width: 420px;
  line-height: 1.5;
}

.handdrawn-mode.dark-mode .empty-state-desc {
  color: #94a3b8;
}

/* 5. 手绘卡片微弧度手绘感边框与交互效果 */
.handdrawn-mode:not(.dark-mode) .nav-item.active {
  border-radius: 12px 14px 11px 13px;
}

.handdrawn-mode.dark-mode .nav-item.active {
  border-radius: 12px 14px 11px 13px;
}

.handdrawn-mode:not(.dark-mode) .clean-input-box {
  border-radius: 20px 22px 19px 21px;
}

.handdrawn-mode.dark-mode .clean-input-box {
  border-radius: 20px 22px 19px 21px;
}

/* 手绘按钮悬停微晃动有趣细节 */
.handdrawn-mode :deep(.n-button--primary-type):hover {
  transform: translateY(-1px) rotate(0.4deg);
}

.handdrawn-mode :deep(.n-button--primary-type):active {
  transform: translateY(1px) rotate(-0.3deg);
}
</style>
