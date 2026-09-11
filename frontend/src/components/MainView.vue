<template>
  <div
    class="app-layout"
    :class="[
      `theme-${effectiveTheme}`,
      isDarkTheme ? 'dark-mode' : 'light-mode',
      'minimal-mode',
      `active-tab-${activeTab}`
    ]"
  >
    <!-- Left Modern Sidebar -->
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
            <div class="brand-tag">v1.1</div>
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
          @click="activeTab = 'search'"
        >
          <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="search" /></span>
          <span class="nav-label">应用搜索</span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'versions' }"
          @click="activeTab = 'versions'"
        >
          <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="history" /></span>
          <span class="nav-label">历史版本</span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'download' }"
          @click="activeTab = 'download'"
        >
          <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="download" /></span>
          <span class="nav-label">下载中心</span>
          <span v-if="activeTaskCount > 0" class="nav-badge">{{ activeTaskCount }}</span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'purchased' }"
          @click="activeTab = 'purchased'"
        >
          <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="purchased" /></span>
          <span class="nav-label">已购应用</span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'installer' }"
          @click="activeTab = 'installer'"
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
          @click="activeTab = 'account'"
        >
          <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="account" /></span>
          <span class="nav-label">账号中心</span>
          <span class="account-dot" :class="isLoggedIn ? 'dot-online' : 'dot-offline'"></span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'settings' }"
          @click="activeTab = 'settings'"
        >
          <span class="nav-icon" aria-hidden="true"><SketchNavIcon name="settings" /></span>
          <span class="nav-label">系统设置</span>
        </button>

        <button
          class="nav-item"
          :class="{ active: activeTab === 'about' }"
          @click="activeTab = 'about'"
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
            <AppSwitch
              v-model="settings.enableProxy"
              size="small"
              @update:model-value="onProxyToggle"
            />
            <span class="quick-tool-label">代理</span>
          </div>

          <!-- Theme Switcher (靠右显示) -->
          <div class="theme-switcher-wrapper">
            <AppPopover
              v-model="showThemePopover"
              placement="top-end"
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
            </AppPopover>
          </div>
        </div>
      </div>
    </aside>

    <!-- Right Main Workspace -->
    <section class="app-main">

      <!-- Dynamic View Container -->
      <div class="view-content-wrapper">
        <!-- VIEW 1: 应用搜索 (search) -->
        <div v-show="activeTab === 'search'" class="view-panel">
          <div class="view-header">
            <div>
              <div class="title-with-badge">
                <h2 class="view-title">应用搜索</h2>
              </div>
              <p class="view-desc">在 Apple App Store 全球库中精准检索正版应用信息</p>
            </div>
          </div>

          <AppCard class="clean-card mb-4 search-bar-card">
            <div class="search-input-group">
              <div class="search-input-wrapper">
                <AppInput
                  v-model="searchForm.term"
                  class="clean-input search-input height-aligned"
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
                    class="height-aligned-select"
                    style="width: 130px;"
                  />
                </div>

                <div class="filter-item">
                  <span class="filter-label">条数</span>
                  <AppSelect
                    v-model="searchForm.limit"
                    :options="limitOptions"
                    size="small"
                    class="height-aligned-select"
                    style="width: 90px;"
                  />
                </div>

                <AppButton
                  variant="primary"
                  size="small"
                  :loading="isSearching"
                  @click="handleSearch"
                  class="height-aligned-btn"
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

        <!-- VIEW 2: 历史版本 (versions) -->
        <div v-show="activeTab === 'versions'" class="view-panel">
          <div class="view-header">
            <div>
              <div class="title-with-badge">
                <h2 class="view-title">历史版本</h2>
              </div>
              <p class="view-desc">解析与枚举 App Store 所有历史构建版本，支持快速精准检索</p>
            </div>
            <div class="header-right-badges">
              <span class="count-pill">共 {{ filteredVersions.length }} / {{ versionItems.length }} 条记录</span>
            </div>
          </div>

          <AppCard class="clean-card mb-4">
            <div class="versions-toolbar">
              <div class="bundle-input-row">
                <div class="clean-input-prefix-box height-aligned">
                  <span class="prefix-label">Bundle ID</span>
                  <AppInput
                    v-model="versionForm.bundleId"
                    class="clean-input flex-1 border-none"
                    placeholder="例如: com.alipay.iphoneclient"
                    @keydown.enter="handleListVersions"
                  />
                </div>

                <AppButton
                  type="primary"
                  size="small"
                  class="height-aligned-btn"
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
                  class="height-aligned-btn"
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
                  class="height-aligned-btn"
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
                    class="clean-input filter-input height-aligned"
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
        </div>

        <!-- VIEW 3: 下载中心 (download) -->
        <div v-show="activeTab === 'download'" class="view-panel">
          <div class="view-header">
            <div>
              <div class="title-with-badge">
                <h2 class="view-title">下载中心</h2>
              </div>
              <p class="view-desc">正在下载与已完成的 IPA 任务管理，自动归档至对应账号目录</p>
            </div>
            <div class="btn-group-row">
              <AppButton
                variant="secondary"
                size="small"
                class="small-aligned-btn"
                :disabled="completedTaskCount === 0"
                @click="handleClearCompleted"
              >
                清空已完成
              </AppButton>
              <AppButton
                variant="secondary"
                size="small"
                class="small-aligned-btn"
                @click="handleOpenDefaultDownloadDir"
              >
                打开存储目录
              </AppButton>
            </div>
          </div>

          <div class="tasks-container">
            <div v-if="downloadTasks.length === 0" class="empty-state-card">
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
                    <AppButton
                      v-if="task.status === 'downloading'"
                      size="tiny"
                      type="error"
                      secondary
                      @click="handleCancelTask(task.id)"
                    >
                      取消
                    </AppButton>
                    <AppButton
                      v-if="task.status === 'completed'"
                      size="tiny"
                      type="primary"
                      @click="handleInstallFromTask(task.outputPath)"
                    >
                      安装到设备
                    </AppButton>
                    <AppButton
                      v-if="task.status === 'error' || task.status === 'canceled'"
                      size="tiny"
                      secondary
                      @click="handleRetryTask(task)"
                    >
                      重试
                    </AppButton>
                    <AppButton size="tiny" quaternary @click="handleDeleteTask(task.id)">
                      删除
                    </AppButton>
                  </div>
                </div>

                <div class="task-meta-bundle">{{ task.bundleID }}</div>

                <div class="task-progress-section">
                  <AppProgress
                    :percentage="task.progress"
                    :status="getTaskProgressStatus(task.status)"
                    :class="'task-progress-' + task.status"
                    :height="6"
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
          <div class="view-header purchased-view-header">
            <div class="purchased-header-copy">
              <div class="title-with-badge">
                <h2 class="view-title">已购应用</h2>
              </div>
              <p class="view-desc">浏览与检索当前 Apple ID 名下已获得正版许可的历史应用库</p>
              <div v-if="isPurchasedLoading" class="purchased-loading-progress" role="status" aria-live="polite">
                <AppProgress
                  v-if="purchasedLoadTotal > 0"
                  class="purchased-loading-bar"
                  :percentage="purchasedLoadProgress"
                  :height="6"
                />
                <div v-else class="purchased-loading-bar purchased-loading-bar-indeterminate" aria-hidden="true"></div>
                <span v-if="purchasedLoadTotal > 0" class="purchased-loading-count">
                  {{ purchasedLoadLoaded }}/{{ purchasedLoadTotal }}
                </span>
                <span v-else class="purchased-loading-count loading-total-text">读取总数…</span>
              </div>
            </div>
            <div class="purchased-header-actions">
              <div class="purchased-search-input-box purchased-header-search">
                <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" class="search-icon">
                  <circle cx="11" cy="11" r="8"></circle>
                  <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                </svg>
                <AppInput
                  v-model="purchasedSearchKeyword"
                  class="purchased-search-input"
                  placeholder="搜索名称、Bundle ID 或 App ID..."
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
              <AppButton
                type="primary"
                size="small"
                class="small-aligned-btn"
                :loading="isPurchasedLoading"
                @click="loadPurchases"
              >
                刷新列表
              </AppButton>
            </div>
          </div>

          <!-- 表格主体 -->
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
                    <AppButton size="tiny" secondary class="mt-2" @click="purchasedSearchKeyword = ''">
                      清空搜索
                    </AppButton>
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
                <span v-if="purchasedSearchKeyword">
                  找到 <b>{{ purchasedMatchTotal }}</b> 款匹配应用（共 {{ purchasedApps.length }} 款）
                </span>
                <span v-else>
                  共 <b>{{ purchasedTotal }}</b> 款已购应用
                </span>
              </div>
              <div class="page-size-selector">
                <span class="size-label">每页</span>
                <AppSelect
                  v-model="purchasedPageSize"
                  size="small"
                  class="height-aligned-select"
                  style="width: 76px;"
                  :options="[
                    { label: '20', value: 20 },
                    { label: '50', value: 50 },
                    { label: '100', value: 100 }
                  ]"
                  @update:model-value="onPurchasedPageSizeChange"
                />
              </div>
              <AppButton
                secondary
                size="small"
                class="small-aligned-btn"
                :disabled="purchasedPage <= 1 || isPurchasedLoading"
                @click="prevPurchasedPage"
              >
                上一页
              </AppButton>
              <span class="page-indicator">第 {{ purchasedPage }} / {{ purchasedPageCount }} 页</span>
              <AppButton
                secondary
                size="small"
                class="small-aligned-btn"
                :disabled="purchasedPage >= purchasedPageCount || isPurchasedLoading"
                @click="nextPurchasedPage"
              >
                下一页
              </AppButton>
            </div>
          </AppCard>
        </div>

        <!-- VIEW 5: 设备直装 (installer) -->
        <div v-show="activeTab === 'installer'" class="view-panel">
          <div class="view-header">
            <div>
              <div class="title-with-badge">
                <h2 class="view-title">设备直装</h2>
              </div>
              <p class="view-desc">通过 USB 数据线将正版签署的 IPA 应用包一键安装至 iOS 设备</p>
            </div>
          </div>

          <div class="installer-grid">
            <!-- Left: IPA File Selection -->
            <AppCard class="clean-card flex-col">
              <div class="card-headline">
                <span class="headline-title">1. 选择安装包</span>
                <span v-if="selectedIPAPath" class="headline-badge">已就绪</span>
              </div>

              <div
                class="clean-drop-zone"
                :class="{ 'drop-active': !!selectedIPAPath }"
                @click="handleSelectIPA"
              >
                <div class="drop-primary-title">
                  {{ selectedIPAPath ? selectedIPAFileName : '点击选择或拖拽 .ipa 文件到此处' }}
                </div>
                <div class="drop-secondary-path text-ellipsis" :title="selectedIPAPath">
                  {{ selectedIPAPath || '支持标准 iOS 签名应用包格式' }}
                </div>
              </div>

              <div class="mt-3 flex-align-center">
                <AppInput
                  :model-value="selectedIPAPath"
                  readonly
                  class="clean-input flex-1 mr-2 height-aligned"
                  placeholder="尚未选择文件"
                />
                <AppButton secondary size="small" class="height-aligned-btn" @click="handleSelectIPA">浏览文件</AppButton>
              </div>
            </AppCard>

            <!-- Right: Device Selection -->
            <AppCard class="clean-card flex-col">
              <div class="card-headline">
                <span class="headline-title">2. 选择苹果设备</span>
                <AppButton
                  secondary
                  size="small"
                  class="small-aligned-btn"
                  :loading="isLoadingDevices"
                  :disabled="isInstallingIPA"
                  @click="loadConnectedDevices"
                >
                  刷新检测
                </AppButton>
              </div>

              <AppSelect
                v-model="selectedDeviceUDID"
                :options="deviceOptions"
                :loading="isLoadingDevices"
                :disabled="isInstallingIPA"
                placeholder="请选择已连接的 iOS 设备"
                size="small"
                class="height-aligned-select"
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
            </AppCard>
          </div>

          <!-- Bottom Action Bar for Installer -->
          <AppCard class="clean-card mt-4 installer-action-banner">
            <div class="installer-action-info">
              <div class="action-banner-title">
                {{ !selectedIPAPath ? '请先选择待安装的 IPA 文件' : (!selectedDeviceUDID ? '请选择目标苹果设备' : '就绪，可以开始安装') }}
              </div>
              <div class="action-banner-desc">
                本工具下载的正版 IPA 需安装至登录了相同 Apple ID 的设备上，未签名包将无法被系统接受。
              </div>
            </div>

            <AppButton
              type="primary"
              size="large"
              :disabled="!selectedIPAPath || !selectedDeviceUDID || isLoadingDevices"
              :loading="isInstallingIPA"
              @click="handleInstallIPA"
              class="install-submit-btn"
            >
              {{ isInstallingIPA ? '正在安装中...' : '开始安装到设备' }}
            </AppButton>
          </AppCard>
        </div>

        <!-- VIEW 6: 账号中心 (account) -->
        <div v-show="activeTab === 'account'" class="view-panel">
          <div class="view-header">
            <div>
              <div class="title-with-badge">
                <h2 class="view-title">账号中心</h2>
              </div>
              <p class="view-desc">管理用于 App Store 正版授权通信与 IPA 下载的 Apple ID 身份凭证</p>
            </div>
          </div>

          <div class="two-columns-layout">
            <!-- Account Status Card -->
            <AppCard class="clean-card">
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
                <AppButton
                  type="error"
                  secondary
                  :disabled="!isLoggedIn"
                  @click="handleRevoke"
                  :loading="isRevoking"
                  class="height-aligned-btn"
                >
                  退出登录
                </AppButton>
                <AppButton
                  secondary
                  @click="handleClearKeychain"
                  :loading="isClearing"
                  title="清理本地钥匙串缓存解决校验异常"
                  class="height-aligned-btn"
                >
                  清理钥匙串缓存
                </AppButton>
                <AppButton
                  secondary
                  @click="() => refreshAccount(false)"
                  :loading="isAccountLoading"
                  class="height-aligned-btn"
                >
                  刷新状态
                </AppButton>
              </div>
            </AppCard>

            <!-- Login Form Card -->
            <AppCard class="clean-card">
              <div class="card-headline">
                <span class="headline-title">登录 Apple ID</span>
              </div>

              <div class="form-item-clean">
                <label class="clean-label">Apple ID 账户邮箱</label>
                <AppInput
                  v-model="loginForm.email"
                  class="clean-input height-aligned"
                  placeholder="例如: your_apple_id@icloud.com"
                />
              </div>

              <div class="form-item-clean mt-3">
                <label class="clean-label">Apple ID 密码</label>
                <AppInput
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
                <AppButton
                  type="primary"
                  block
                  size="small"
                  class="height-aligned-btn"
                  :loading="isLoggingIn"
                  @click="handleLogin"
                >
                  登录
                </AppButton>
              </div>
            </AppCard>
          </div>
        </div>

        <!-- VIEW 7: 系统设置 (settings) -->
        <div v-show="activeTab === 'settings'" class="view-panel">
          <div class="view-header">
            <div>
              <div class="title-with-badge">
                <h2 class="view-title">系统设置</h2>
              </div>
              <p class="view-desc">全局参数、网络代理、下载路径与基础运行引擎配置</p>
            </div>
            <AppButton
              type="primary"
              size="small"
              class="height-aligned-btn settings-save-btn"
              @click="handleSaveSettings"
            >
              保存并应用设置
            </AppButton>
          </div>

          <AppCard class="clean-card settings-stack">
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
                  <AppInput
                    v-model="settings.keychainPassphrase"
                    type="password"
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
                  <AppSwitch v-model="settings.enableProxy" />
                </div>
              </div>

              <div class="settings-field-row">
                <div class="field-meta">
                  <div class="field-title">代理服务器地址</div>
                  <div class="field-desc">支持 Clash / v2rayN 等常见本地代理端口（如 http://127.0.0.1:10808）</div>
                </div>
                <div class="field-control">
                  <div class="setting-input-action-group">
                    <AppInput
                      v-model="settings.proxyUrl"
                      :disabled="!settings.enableProxy"
                      class="clean-input flex-1 height-aligned"
                      placeholder="http://127.0.0.1:10808"
                    />
                    <AppButton
                      variant="secondary"
                      size="small"
                      class="height-aligned-btn setting-fixed-btn"
                      :disabled="!settings.enableProxy"
                      :loading="isTestingProxy"
                      @click="handleTestProxy"
                    >
                      测试
                    </AppButton>
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
                    <AppInput
                      :model-value="settings.defaultDownloadDir"
                      readonly
                      disabled
                      class="clean-input flex-1 height-aligned"
                    />
                    <AppButton
                      variant="secondary"
                      size="small"
                      class="height-aligned-btn setting-fixed-btn"
                      @click="handleOpenDefaultDownloadDir"
                    >
                      打开
                    </AppButton>
                  </div>
                </div>
              </div>

              <div class="settings-field-row">
                <div class="field-meta">
                  <div class="field-title">默认目标平台</div>
                  <div class="field-desc">检索与下载默认针对的 Apple 硬件体系</div>
                </div>
                <div class="field-control">
                  <AppSelect
                    v-model="settings.defaultPlatform"
                    :options="platformOptions"
                    size="small"
                    class="height-aligned-select"
                    style="width: 100%;"
                  />
                </div>
              </div>
            </div>
          </AppCard>
        </div>

        <!-- VIEW 8: 关于软件 (about) -->
        <div v-show="activeTab === 'about'" class="view-panel">
          <div class="view-header">
            <div>
              <div class="title-with-badge">
                <h2 class="view-title">关于软件</h2>
              </div>
              <p class="view-desc">果仓助手 (AppleVault) 软件信息、项目开源地址与使用说明</p>
            </div>
          </div>

          <AppCard class="clean-card about-card-stack">
            <!-- App Banner -->
            <div class="about-hero">
              <div class="about-large-icon">
                <img class="about-logo-image" src="../assets/applevault-logo.webp" alt="AppleVault" />
              </div>
              <div class="about-hero-text">
                <div class="about-app-title">果仓助手 (AppleVault)</div>
                <div class="about-version-line">
                  <span class="about-version-badge">版本 v1.1.0</span>
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
                <AppInput
                  model-value="https://github.com/wnnz/AppleVault"
                  readonly
                  class="clean-input flex-1 height-aligned"
                />
                <AppButton
                  type="primary"
                  size="small"
                  class="height-aligned-btn"
                  @click="openGitHub"
                >
                  访问 GitHub
                </AppButton>
                <AppButton
                  secondary
                  size="small"
                  class="height-aligned-btn copy-address-btn"
                  @click="copyGitHubUrl"
                >
                  复制地址
                </AppButton>
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
          </AppCard>
        </div>

      </div>

      <!-- Bottom Minimal Status Bar -->
      <div class="bottom-status-bar">
        <div class="status-indicator-section">
          <span class="status-pulse" :class="{ 'pulse-busy': isAnyOperationRunning }"></span>
          <span class="status-summary-text text-ellipsis">{{ statusText }}</span>
        </div>

        <div class="status-bar-tools">
          <AppButton
            v-if="isAnyOperationRunning"
            size="tiny"
            type="error"
            secondary
            @click="handleCancel"
            class="mr-2"
          >
            终止操作
          </AppButton>

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
    <AppDialog
      v-model="show2FAModal"
      title="Apple ID 双重认证"
      style="width: 400px; border-radius: 14px;"
      :mask-closable="false"
    >
      <div class="modal-dialog-inner">
        <p class="dialog-desc">已向你的受信任 Apple 设备发送了验证码，请输入 6 位数字验证码：</p>
        <AppInput
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
          <AppButton secondary @click="cancel2FA" class="height-aligned-btn">取消</AppButton>
          <AppButton
            type="primary"
            :disabled="twoFACode.length !== 6"
            :loading="isLoggingIn"
            @click="confirm2FA"
            class="height-aligned-btn"
          >
            提交验证
          </AppButton>
        </div>
      </div>
    </AppDialog>

    <!-- Target Version Modal Dialog -->
    <AppDialog
      v-model="showTargetVersionModal"
      title="查找指定版本"
      style="width: 440px; border-radius: 14px;"
    >
      <div class="modal-dialog-inner">
        <p class="dialog-desc">
          请输入目标版本号（例如 <code>10.2.80</code> 或 <code>8.0.0</code>），程序将智能检索历史构建记录并快速定位匹配版本。
        </p>
        <AppInput
          ref="targetVersionInputRef"
          v-model="targetVersionInput"
          class="clean-input height-aligned"
          placeholder="例如: 10.2.80"
          autofocus
          @keydown.enter="confirmStartTargetQuery"
        />
        <div class="dialog-action-buttons mt-4">
          <AppButton secondary @click="showTargetVersionModal = false" class="height-aligned-btn">取消</AppButton>
          <AppButton
            type="primary"
            :disabled="!targetVersionInput.trim()"
            @click="confirmStartTargetQuery"
            class="height-aligned-btn"
          >
            开始查询
          </AppButton>
        </div>
      </div>
    </AppDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick, h, defineComponent } from 'vue'
import { useClipboard } from '@vueuse/core'
import AppButton from '../ui/components/AppButton.vue'
import AppCard from '../ui/components/AppCard.vue'
import AppDialog from '../ui/components/AppDialog.vue'
import AppInput from '../ui/components/AppInput.vue'
import AppPopover from '../ui/components/AppPopover.vue'
import AppProgress from '../ui/components/AppProgress.vue'
import AppSelect from '../ui/components/AppSelect.vue'
import AppSwitch from '../ui/components/AppSwitch.vue'
import AppTable from '../ui/components/AppTable.vue'
import { useAppDialog, useAppMessage } from '../ui/feedback'
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
import { type AppTheme, themeOptions } from '../ui/theme'
import standardLogo from '../assets/applevault-logo.webp'
import handdrawnSidebarArt from '../ui/themes/handdrawn/assets/sketch-sidebar-art.svg'
import handdrawnSidebarFloatingLeaf from '../ui/themes/handdrawn/assets/sketch-floating-leaf.svg'

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

const themeList = themeOptions

function selectTheme(themeKey: AppTheme) {
  emit('update:currentTheme', themeKey)
  showThemePopover.value = false
}

const message = useAppMessage()
const dialog = useAppDialog()
const { copy } = useClipboard({ legacy: true })

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
      return h('div', { class: 'app-button-group' }, [
        h(AppButton, { size: 'tiny', secondary: true, onClick: () => selectAppForVersions(row) }, () => '历史版本'),
        h(AppButton, { size: 'tiny', type: 'primary', onClick: () => downloadFromSearch(row) }, () => '下载'),
        h(AppButton, { size: 'tiny', secondary: true, onClick: () => handlePurchaseApp(row.bundleID) }, () => '获取许可')
      ])
    }
  }
]

const displayedSearchColumns = computed(() => {
  if (effectiveTheme.value !== 'handdrawn') return searchColumns

  const widths = [196, 163, 94, 87, 83, 217]
  return searchColumns.map((column, index) => ({
    ...column,
    minWidth: undefined,
    width: widths[index]
  }))
})

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
    versionItems.value = [...versionItems.value]
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
      return h('div', { class: 'app-button-group' }, [
        h(AppButton, { size: 'tiny', secondary: true, onClick: () => selectAppForVersions(row) }, () => '历史版本'),
        h(AppButton, { size: 'tiny', type: 'primary', onClick: () => downloadFromPurchased(row) }, () => '下载')
      ])
    }
  }
]

const displayedVersionColumns = computed(() => {
  if (effectiveTheme.value !== 'handdrawn') return versionColumns

  const widths = [314, 139, 101, 129, 155]
  return versionColumns.map((column, index) => ({
    ...column,
    minWidth: undefined,
    width: widths[index]
  }))
})

const displayedPurchasedColumns = computed(() => {
  if (effectiveTheme.value !== 'handdrawn') return purchasedColumns

  const widths = [227, 200, 104, 153, 156]
  return purchasedColumns.map((column, index) => ({
    ...column,
    minWidth: undefined,
    width: widths[index]
  }))
})

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
    versionItems.value = [...versionItems.value]
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
  purchasedLoadLoaded.value = 0
  purchasedLoadTotal.value = 0
  statusText.value = '正在获取已购应用总数...'

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
        statusText.value = `正在加载已购应用 (${purchasedLoadLoaded.value}/${firstTotal})...`
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
    statusText.value = `已购应用加载完成 (共 ${purchasedApps.value.length} 款)。`
  } catch (err: any) {
    message.error(`加载已购列表失败: ${err}`)
    statusText.value = '加载失败'
  } finally {
    isPurchasedLoading.value = false
    isAnyOperationRunning.value = false
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

async function copyGitHubUrl() {
  await copy('https://github.com/wnnz/AppleVault')
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
    account.value = { name: '果仓助手用户', email: 'applevault.user@icloud.com', success: true }
    const demoStatusByTab: Record<string, string> = {
      search: '搜索完成，找到 6 个应用。',
      versions: '共获取 8 个历史版本记录。',
      download: '当前有 1 个任务正在下载，速度 12.8 MB/s',
      purchased: '已购应用加载完成（共 6 款）。',
      installer: '已就绪，已检测到 iPhone 15 Pro Max',
      account: '已登录 果仓助手用户（applevault.user@icloud.com）',
      settings: '底层引擎就绪，配置已加载',
      about: '果仓助手 (AppleVault) v1.1.0'
    }
    statusText.value = demoStatusByTab[tabParam || 'search'] || '就绪'
    searchForm.value.term = '微信'
    settings.value.proxyUrl = 'http://127.0.0.1:7890'
    settings.value.defaultDownloadDir = 'D:\\Dev\\AppleVault\\data\\downloads\\applevault.user'
    searchResults.value = [
      { id: 414478124, bundleID: 'com.tencent.xin', name: '微信 (WeChat)', version: '8.0.50', price: 0, displayPrice: '免费' },
      { id: 333206289, bundleID: 'com.alipay.iphoneclient', name: '支付宝 - 生活好 支付宝', version: '10.5.88', price: 0, displayPrice: '免费' },
      { id: 736536022, bundleID: 'tv.danmaku.bilianime', name: '哔哩哔哩 (Bilibili)', version: '7.82.0', price: 0, displayPrice: '免费' },
      { id: 590338362, bundleID: 'com.netease.cloudmusic', name: '网易云音乐', version: '9.0.70', price: 0, displayPrice: '免费' },
      { id: 835599320, bundleID: 'com.zhiliaoapp.musically', name: 'TikTok - 精彩短视频与音乐', version: '35.8.0', price: 0, displayPrice: '免费' },
      { id: 444934666, bundleID: 'com.tencent.mqq', name: 'QQ - 轻松做自己', version: '9.0.65', price: 0, displayPrice: '免费' }
    ]
    versionForm.value.bundleId = 'com.tencent.xin'
    versionForm.value.appName = '微信'
    versionItems.value = [
      { versionId: '868192301', displayVersion: '8.0.50', fileSize: '286.4 MB', releaseDate: '2024-08-15' },
      { versionId: '867204918', displayVersion: '8.0.49', fileSize: '284.1 MB', releaseDate: '2024-07-20' },
      { versionId: '865819021', displayVersion: '8.0.48', fileSize: '279.8 MB', releaseDate: '2024-06-12' },
      { versionId: '864201990', displayVersion: '8.0.47', fileSize: '275.2 MB', releaseDate: '2024-05-08' },
      { versionId: '862901124', displayVersion: '8.0.46', fileSize: '270.5 MB', releaseDate: '2024-04-01' },
      { versionId: '861502391', displayVersion: '8.0.45', fileSize: '268.0 MB', releaseDate: '2024-03-05' },
      { versionId: '859810234', displayVersion: '8.0.44', fileSize: '263.8 MB', releaseDate: '2024-01-22' },
      { versionId: '858201992', displayVersion: '8.0.43', fileSize: '260.1 MB', releaseDate: '2023-12-18' }
    ]
    downloadTasks.value = [
      { id: '1', appName: '微信 (WeChat)', bundleID: 'com.tencent.xin', appId: 414478124, version: '8.0.50', versionId: '868192301', fileSize: '286.4 MB', totalBytes: 300312000, currBytes: 195202800, progress: 65, speed: '12.8 MB/s', status: 'downloading', outputPath: '', errorMessage: '', createdAt: '2024-08-20 15:30:12' },
      { id: '2', appName: '支付宝', bundleID: 'com.alipay.iphoneclient', appId: 333206289, version: '10.5.88', versionId: '865001129', fileSize: '142.0 MB', totalBytes: 148897000, currBytes: 148897000, progress: 100, speed: '已完成', status: 'completed', outputPath: 'data/downloads/user@icloud.com/alipay.ipa', errorMessage: '', createdAt: '2024-08-20 15:24:05' },
      { id: '3', appName: 'Infuse · 精彩影音播放器', bundleID: 'com.firecore.infuse', appId: 1136220934, version: '7.7.2', versionId: '863920191', fileSize: '118.5 MB', totalBytes: 124256256, currBytes: 124256256, progress: 100, speed: '已完成', status: 'completed', outputPath: 'data/downloads/user@icloud.com/infuse.ipa', errorMessage: '', createdAt: '2024-08-20 14:10:33' },
      { id: '4', appName: '网易云音乐', bundleID: 'com.netease.cloudmusic', appId: 590338362, version: '9.0.70', versionId: '864201991', fileSize: '215.3 MB', totalBytes: 225758413, currBytes: 0, progress: 0, speed: '等待中', status: 'pending', outputPath: '', errorMessage: '', createdAt: '2024-08-20 15:31:00' }
    ]
    purchasedApps.value = [
      { id: 414478124, bundleID: 'com.tencent.xin', name: '微信 (WeChat)', version: '8.0.50', price: 0, purchaseDate: '2024-01-15 10:20', displayPrice: '已购' },
      { id: 333206289, bundleID: 'com.alipay.iphoneclient', name: '支付宝', version: '10.5.88', price: 0, purchaseDate: '2024-02-08 14:12', displayPrice: '已购' },
      { id: 1136220934, bundleID: 'com.firecore.infuse', name: 'Infuse • 精彩影音播放器', version: '7.7.2', price: 0, purchaseDate: '2024-03-22 09:45', displayPrice: '已购' },
      { id: 916364737, bundleID: 'com.procreate.pocket', name: 'Procreate Pocket', version: '4.0.11', price: 0, purchaseDate: '2024-04-10 18:30', displayPrice: '已购' },
      { id: 1596487405, bundleID: 'com.taguirov.adam.WebDAV', name: 'WebDAV Manager', version: '2.1.0', price: 0, purchaseDate: '2024-05-19 16:30', displayPrice: '已购' },
      { id: 736536022, bundleID: 'tv.danmaku.bilianime', name: '哔哩哔哩 (Bilibili)', version: '7.82.0', price: 0, purchaseDate: '2024-06-01 21:05', displayPrice: '已购' }
    ]
    purchasedTotal.value = 6
    devices.value = [
      { udid: '00008130-001A49021E28001C', name: 'iPhone 15 Pro Max', productType: 'iPhone 15 Pro Max', productVersion: '17.5.1', connectionType: 'USB 3.0' }
    ]
    selectedDeviceUDID.value = '00008130-001A49021E28001C'
    selectedIPAPath.value = 'D:\\Dev\\AppleVault\\data\\downloads\\applevault.user@icloud.com\\WeChat_8.0.50.ipa'
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
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 10px;
  box-shadow: 0 4px 10px rgba(0, 113, 227, 0.25);
  overflow: hidden;
}

.brand-logo-image,
.about-logo-image {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: contain;
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
  height: 30px !important;
  box-sizing: border-box !important;
  line-height: 28px !important;
}

.height-aligned-btn {
  height: 30px !important;
  box-sizing: border-box !important;
  padding: 0 12px !important;
  font-size: 12px !important;
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
  height: 30px;
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

.clean-input-prefix-box:focus-within {
  border-color: #0071e3;
  box-shadow: 0 0 0 2px rgba(0, 113, 227, 0.12);
  background-color: #ffffff;
}

.dark-mode .clean-input-prefix-box:focus-within {
  border-color: #0284c7;
  box-shadow: 0 0 0 2px rgba(2, 132, 199, 0.2);
  background-color: #1e293b;
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
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 0 10px;
  height: 30px;
  width: 330px;
  max-width: 100%;
  transition: all 0.2s ease;
  box-sizing: border-box;
}

.dark-mode .purchased-search-input-box {
  background: #0f172a;
  border-color: #334155;
}

.purchased-search-input-box:focus-within {
  border-color: #0071e3;
  box-shadow: 0 0 0 2px rgba(0, 113, 227, 0.12);
  background: #ffffff;
}

.dark-mode .purchased-search-input-box:focus-within {
  border-color: #0284c7;
  box-shadow: 0 0 0 2px rgba(2, 132, 199, 0.2);
  background: #1e293b;
}

.purchased-search-input {
  border: none !important;
  background: transparent !important;
  box-shadow: none !important;
  outline: none !important;
  font-size: 12.5px;
  padding: 0 8px;
  color: #0f172a;
  width: 100%;
  height: 100%;
  box-sizing: border-box;
}

.purchased-search-input:focus {
  border: none !important;
  background: transparent !important;
  box-shadow: none !important;
  outline: none !important;
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

.purchased-table {
  flex: 1;
  min-height: 0;
  height: auto !important;
}

.purchased-bottom-pagination {
  display: flex;
  flex: 0 0 40px;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  padding: 6px 4px 0 12px;
  box-sizing: border-box;
}

.page-size-selector {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-right: 4px;
}

:deep(.height-aligned-select.app-select) {
  min-height: 30px;
  height: 30px;
}

:deep(.height-aligned-select.app-select__label) {
  height: 30px;
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
  justify-content: flex-end;
  gap: 10px;
  width: 370px;
  min-width: 280px;
}

.purchased-view-header.view-header > .purchased-header-actions {
  position: absolute;
  top: 0;
  right: 0;
}

.purchased-view-header {
  position: relative;
  display: block;
  min-height: 62px;
}

.purchased-header-copy {
  min-width: 0;
}

.purchased-header-search {
  flex: 1 1 270px;
  width: auto;
  min-width: 180px;
}

.purchased-footer-count {
  margin-right: auto;
  padding-left: 4px;
}

.purchased-loading-progress {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 170px;
  margin-top: 7px;
}

.purchased-loading-bar {
  flex: 1;
  min-width: 90px;
}

.purchased-loading-bar-indeterminate {
  position: relative;
  height: 6px;
  overflow: hidden;
  background: #e2e8f0;
  border-radius: 3px;
}

.purchased-loading-bar-indeterminate::after {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: 38%;
  border-radius: 3px;
  background: #0071e3;
  animation: purchased-loading-slide 1.1s ease-in-out infinite;
}

@keyframes purchased-loading-slide {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(270%);
  }
}

.purchased-loading-count {
  min-width: 54px;
  color: #64748b;
  font-size: 12px;
  text-align: right;
  white-space: nowrap;
}

.dark-mode .purchased-loading-count {
  color: #94a3b8;
}

.dark-mode .purchased-loading-bar-indeterminate {
  background: #334155;
}

.dark-mode .purchased-loading-bar-indeterminate::after {
  background: #38bdf8;
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
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 6px 16px rgba(0, 113, 227, 0.25);
  flex-shrink: 0;
  margin-left: 6px;
  overflow: hidden;
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

.title-with-badge {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.table-empty-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 36px 20px;
  text-align: center;
}

.empty-state-title {
  font-size: 14px;
  font-weight: 700;
  margin-bottom: 6px;
  color: #1e3a5f;
}

.dark-mode .empty-state-title {
  color: #e2e8f0;
}

.empty-state-desc {
  font-size: 12px;
  color: #627d98;
  max-width: 420px;
  line-height: 1.5;
}

.dark-mode .empty-state-desc {
  color: #94a3b8;
}
</style>
