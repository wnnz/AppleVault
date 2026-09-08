<template>
  <div class="app-layout" :class="{ 'dark-mode': isDark }">
    <!-- 1. Top Header Bar -->
    <header class="top-header">
      <div class="header-brand">
        <span class="brand-emoji">🍎</span>
        <span class="brand-title">果仓助手</span>
        <span class="brand-badge">v1.0</span>
        <span class="brand-subtitle">— AppleVault 苹果 App Store 正版与历史版本下载</span>
      </div>

      <div class="header-tools">
        <!-- Proxy Pill Badge -->
        <div class="pill-badge" title="全局网络代理设置">
          <n-switch v-model:value="settings.enableProxy" size="small" @update:value="onProxyToggle" />
          <span class="pill-label">代理</span>
          <span class="pill-value text-ellipsis" :title="settings.proxyUrl">{{ settings.proxyUrl }}</span>
        </div>

        <!-- Account Pill Badge -->
        <div class="pill-badge">
          <span class="indicator-dot" :class="isLoggedIn ? 'dot-active' : 'dot-inactive'"></span>
          <span class="pill-value font-bold">{{ account.name || '未登录' }}</span>
          <span v-if="account.email" class="pill-sub">({{ account.email }})</span>
        </div>

        <!-- Refresh Account Button -->
        <n-button size="small" secondary @click="() => refreshAccount(false)" :loading="isAccountLoading">
          刷新状态
        </n-button>

        <!-- Toggle Console Logs -->
        <n-button size="small" secondary @click="toggleLogs" :title="showLogs ? '收起底部日志控制台' : '展开底部日志控制台'">
          📜 日志
        </n-button>

        <!-- Theme Switcher -->
        <n-button size="small" circle quaternary @click="emit('toggleTheme')" :title="isDark ? '切换到亮色模式' : '切换到深色模式'">
          <template #icon>
            <span style="font-size: 14px;">{{ isDark ? '☀️' : '🌙' }}</span>
          </template>
        </n-button>
      </div>
    </header>

    <!-- 2. Main Tab Content Area -->
    <main class="main-body">
      <n-tabs v-model:value="activeTab" type="line" animated class="custom-tabs">
        
        <!-- TAB 1: 账号管理 -->
        <n-tab-pane name="account" tab="👤 账号管理">
          <div class="tab-scroll-container">
            <div class="two-columns-layout">
              <!-- Left Card: Current Account Status -->
              <div class="fluent-card">
                <h3 class="card-title">当前登录状态</h3>
                
                <div class="key-value-row">
                  <span class="row-label">账号名称:</span>
                  <span class="row-value font-bold">{{ account.name || '未登录' }}</span>
                </div>

                <div class="key-value-row">
                  <span class="row-label">Apple ID:</span>
                  <span class="row-value">{{ account.email || '未登录' }}</span>
                </div>

                <div class="notice-box notice-warning">
                  💡 提示：IPATool 使用你自己的 Apple ID 官方凭据直接从苹果 App Store 服务器下载正版 IPA，下载的文件自带你的个人授权，装入设备不会闪退。
                </div>

                <div class="btn-group-row">
                  <n-button type="error" :disabled="!isLoggedIn" @click="handleRevoke" :loading="isRevoking">
                    退出登录 (Revoke)
                  </n-button>
                  <n-button secondary @click="handleClearKeychain" :loading="isClearing" title="删除 ~/.ipatool 目录，解决密码校验失败问题">
                    清空本地密钥库缓存
                  </n-button>
                </div>
              </div>

              <!-- Right Card: Login Form -->
              <div class="fluent-card">
                <h3 class="card-title">登录 Apple ID</h3>

                <div class="form-group">
                  <label class="field-label">Apple ID 邮箱:</label>
                  <n-input v-model:value="loginForm.email" placeholder="例如: your_apple_id@icloud.com" size="medium" />
                </div>

                <div class="form-group">
                  <label class="field-label">Apple ID 密码:</label>
                  <n-input v-model:value="loginForm.password" type="password" show-password-on="click" placeholder="请输入密码" size="medium" @keydown.enter="handleLogin" />
                </div>

                <div class="notice-box notice-gray">
                  🔒 验证说明：点击登录后，若你的 Apple ID 开启了双重认证（2FA），界面会自动弹出验证码输入弹窗，在手机上确认后输入 6 位验证码即可完成登录。
                </div>

                <n-button type="primary" block size="large" :loading="isLoggingIn" @click="handleLogin" style="height: 38px;">
                  登 录 Apple ID
                </n-button>
              </div>
            </div>
          </div>
        </n-tab-pane>

        <!-- TAB 2: 已购应用 -->
        <n-tab-pane name="purchased" tab="📦 已购应用">
          <div class="tab-table-container">
            <div class="fluent-card toolbar-card">
              <div class="purchased-toolbar">
                <n-button type="primary" size="medium" :loading="isPurchasedLoading" @click="loadPurchases">
                  刷新已购列表
                </n-button>

                <div class="pagination-area">
                  <n-button secondary size="small" :disabled="purchasedPage <= 1" @click="prevPurchasedPage">
                    上一页
                  </n-button>
                  <span class="pagination-info">第 {{ purchasedPage }} 页 (共 {{ purchasedTotal }} 个应用)</span>
                  <n-button secondary size="small" :disabled="purchasedPage * 20 >= purchasedTotal" @click="nextPurchasedPage">
                    下一页
                  </n-button>
                </div>
              </div>
            </div>

            <div class="fluent-card table-card">
              <n-data-table
                :columns="purchasedColumns"
                :data="purchasedApps"
                :loading="isPurchasedLoading"
                size="small"
                flex-height
                style="height: 100%;"
              />
            </div>
          </div>
        </n-tab-pane>

        <!-- TAB 3: 搜索应用 -->
        <n-tab-pane name="search" tab="🔍 搜索应用">
          <div class="tab-table-container">
            <!-- Search Toolbar Card -->
            <div class="fluent-card toolbar-card">
              <div class="search-toolbar">
                <n-input v-model:value="searchForm.term" placeholder="输入关键词搜索应用（如: 支付宝、微信、TikTok）" class="flex-1" size="medium" @keydown.enter="handleSearch" />
                
                <div class="select-wrapper">
                  <span class="label-inline">平台:</span>
                  <n-select v-model:value="searchForm.platform" :options="platformOptions" size="medium" style="width: 130px;" />
                </div>

                <div class="select-wrapper">
                  <span class="label-inline">数量:</span>
                  <n-select v-model:value="searchForm.limit" :options="limitOptions" size="medium" style="width: 85px;" />
                </div>

                <n-button type="primary" size="medium" :loading="isSearching" @click="handleSearch" style="width: 80px;">
                  搜 索
                </n-button>
              </div>
            </div>

            <!-- Search Results Table Card -->
            <div class="fluent-card table-card">
              <n-data-table
                :columns="searchColumns"
                :data="searchResults"
                :loading="isSearching"
                :pagination="{ pageSize: 10 }"
                size="small"
                flex-height
                style="height: 100%;"
              />
            </div>
          </div>
        </n-tab-pane>

        <!-- TAB 4: 历史版本 -->
        <n-tab-pane name="versions" tab="📜 历史版本">
          <div class="tab-table-container">
            <!-- Versions Toolbar Card -->
            <div class="fluent-card toolbar-card">
              <div class="version-toolbar-rows">
                <div class="toolbar-row">
                  <span class="label-inline">目标 Bundle ID:</span>
                  <n-input v-model:value="versionForm.bundleId" placeholder="例如: com.alipay.iphoneclient" style="width: 300px;" size="medium" @keydown.enter="handleListVersions" />
                  <n-button
                    type="primary"
                    size="medium"
                    :loading="isListingVersions"
                    :disabled="isListingVersions || isBatchQuerying || isTargetQuerying"
                    @click="handleListVersions"
                  >
                    获取历史版本列表
                  </n-button>
                  <n-button
                    :type="isBatchQuerying ? 'error' : 'default'"
                    :secondary="!isBatchQuerying"
                    size="medium"
                    :disabled="versionItems.length === 0 || isListingVersions || isTargetQuerying"
                    @click="handleBatchQueryClick"
                  >
                    <template #icon v-if="isBatchQuerying">
                      <n-spin size="small" />
                    </template>
                    {{ isBatchQuerying ? '⏹️ 停止查询' : '批量查询前 30 个版本号' }}
                  </n-button>
                  <n-button
                    :type="isTargetQuerying ? 'error' : 'default'"
                    :secondary="!isTargetQuerying"
                    size="medium"
                    :disabled="versionItems.length === 0 || isListingVersions || isBatchQuerying"
                    @click="handleTargetQueryClick"
                  >
                    <template #icon v-if="isTargetQuerying">
                      <n-spin size="small" />
                    </template>
                    {{ isTargetQuerying ? '⏹️ 停止查询' : '🎯 查询到指定版本' }}
                  </n-button>
                </div>

                <div class="toolbar-row mt-2">
                  <span class="label-inline">筛选版本 (输入版本号如 10.2.96、体积或构建 ID):</span>
                  <n-input v-model:value="versionForm.filter" placeholder="实时过滤筛选..." style="width: 280px;" size="small" clearable />
                  <span class="count-tag">共 {{ filteredVersions.length }} / {{ versionItems.length }} 个版本记录</span>
                </div>
              </div>
            </div>

            <!-- Versions DataGrid Card -->
            <div class="fluent-card table-card">
              <n-data-table
                :columns="versionColumns"
                :data="filteredVersions"
                :loading="isListingVersions"
                :virtual-scroll="true"
                flex-height
                style="height: 100%;"
                size="small"
              />
            </div>
          </div>
        </n-tab-pane>

        <!-- TAB 5: 下载中心 -->
        <n-tab-pane name="download" tab="⬇️ 下载中心">
          <div class="tab-table-container">
            <div class="fluent-card toolbar-card">
              <div class="purchased-toolbar">
                <div class="flex-align-center">
                  <span class="font-bold text-sm mr-2">下载任务列表</span>
                  <n-tag type="info" size="small" round>
                    共 {{ downloadTasks.length }} 个任务 ({{ activeTaskCount }} 个进行中，{{ completedTaskCount }} 个已完成)
                  </n-tag>
                </div>

                <div class="btn-group-row">
                  <n-button secondary size="small" :disabled="completedTaskCount === 0" @click="handleClearCompleted">
                    清空已完成
                  </n-button>
                  <n-button secondary size="small" @click="handleOpenDefaultDownloadDir">
                    📁 打开下载目录
                  </n-button>
                </div>
              </div>
            </div>

            <div class="fluent-card table-card" style="padding: 12px; overflow-y: auto;">
              <div v-if="downloadTasks.length === 0" class="empty-tasks-box">
                <n-empty description="暂无下载任务。请在「应用搜索」或「历史版本」中点击下载直接添加！">
                  <template #icon>
                    <span style="font-size: 36px;">⬇️</span>
                  </template>
                </n-empty>
              </div>

              <div v-else class="task-items-list">
                <div v-for="task in downloadTasks" :key="task.id" class="task-card">
                  <div class="task-main">
                    <div class="task-info">
                      <div class="task-title-row">
                        <span class="task-app-name">{{ task.appName }}</span>
                        <n-tag size="small" type="success" :bordered="false" round class="task-ver-tag">
                          {{ task.version }}
                        </n-tag>
                        <span class="task-bundle-id">{{ task.bundleID }}</span>
                        <span v-if="task.versionId" class="task-build-id">Build: {{ task.versionId }}</span>
                      </div>

                      <div class="task-progress-bar">
                        <n-progress
                          type="line"
                          :percentage="task.progress"
                          :status="getTaskProgressStatus(task.status)"
                          :show-indicator="false"
                          :height="8"
                          border-radius="4"
                        />
                      </div>

                      <div class="task-meta-row">
                        <div class="task-meta-left">
                          <span class="task-badge-status" :class="'badge-' + task.status">{{ getTaskStatusText(task.status) }}</span>
                          <span v-if="task.status === 'downloading'" class="task-speed font-semibold">{{ task.speed }}</span>
                          <span v-if="task.status === 'downloading' && task.totalBytes > 0" class="task-bytes">
                            {{ formatTaskBytes(task.currBytes) }} / {{ formatTaskBytes(task.totalBytes) }}
                          </span>
                          <span v-else-if="task.fileSize && task.fileSize !== '-'" class="task-size">大小: {{ task.fileSize }}</span>
                          <span v-if="task.status === 'completed' && task.outputPath" class="task-path" :title="task.outputPath">
                            保存至: {{ task.outputPath }}
                          </span>
                          <span v-if="task.status === 'error'" class="task-err-msg">
                            错误: {{ task.errorMessage }}
                          </span>
                        </div>
                        <div class="task-meta-right">
                          <span class="task-pct font-bold">{{ task.progress }}%</span>
                          <span class="task-time ml-2 text-gray-sub">{{ task.createdAt }}</span>
                        </div>
                      </div>
                    </div>

                    <div class="task-actions">
                      <n-button v-if="task.status === 'downloading'" size="tiny" type="error" secondary @click="handleCancelTask(task.id)">
                        取消
                      </n-button>
                      <n-button v-if="task.status === 'completed'" size="tiny" type="primary" @click="handleInstallFromTask(task.outputPath)">
                        📲 安装
                      </n-button>
                      <n-button v-if="task.status === 'completed'" size="tiny" secondary @click="handleOpenDir(task.outputPath)">
                        所在目录
                      </n-button>
                      <n-button v-if="task.status === 'error' || task.status === 'canceled'" size="tiny" secondary @click="handleRetryTask(task)">
                        重试
                      </n-button>
                      <n-button size="tiny" quaternary @click="handleDeleteTask(task.id)">
                        删除
                      </n-button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </n-tab-pane>

        <!-- TAB 6: IPA 安装 -->
        <n-tab-pane name="installer" tab="📲 IPA 安装">
          <div class="tab-scroll-container">
            <div class="installer-layout">
              <div class="fluent-card">
                <h3 class="card-title">选择 IPA 安装包</h3>
                <div
                  class="ipa-drop-zone"
                  :class="{ 'has-file': !!selectedIPAPath }"
                  @click="handleSelectIPA"
                >
                  <div class="drop-icon">{{ selectedIPAPath ? '✅' : '📦' }}</div>
                  <div class="drop-title">{{ selectedIPAPath ? selectedIPAFileName : '点击选择或拖放 IPA 文件到此处' }}</div>
                  <div class="drop-subtitle" :title="selectedIPAPath">
                    {{ selectedIPAPath || '仅支持 .ipa 文件' }}
                  </div>
                </div>
                <n-input-group class="mt-3">
                  <n-input :value="selectedIPAPath" readonly placeholder="尚未选择 IPA 文件" />
                  <n-button secondary @click="handleSelectIPA">浏览...</n-button>
                </n-input-group>
              </div>

              <div class="fluent-card">
                <div class="installer-card-header">
                  <h3 class="card-title">选择已连接的苹果设备</h3>
                  <n-button secondary size="small" :loading="isLoadingDevices" :disabled="isInstallingIPA" @click="loadConnectedDevices">
                    刷新设备
                  </n-button>
                </div>

                <n-select
                  v-model:value="selectedDeviceUDID"
                  :options="deviceOptions"
                  :loading="isLoadingDevices"
                  :disabled="isInstallingIPA"
                  placeholder="请选择设备"
                  size="large"
                />

                <div v-if="selectedDevice" class="device-detail-card">
                  <div class="key-value-row">
                    <span class="row-label">设备名称:</span>
                    <span class="row-value font-bold">{{ selectedDevice.name }}</span>
                  </div>
                  <div class="key-value-row">
                    <span class="row-label">设备型号:</span>
                    <span class="row-value">{{ selectedDevice.productType || '-' }}</span>
                  </div>
                  <div class="key-value-row">
                    <span class="row-label">系统版本:</span>
                    <span class="row-value">{{ selectedDevice.productVersion || '-' }}</span>
                  </div>
                  <div class="key-value-row">
                    <span class="row-label">连接方式:</span>
                    <span class="row-value">{{ selectedDevice.connectionType || '-' }}</span>
                  </div>
                  <div class="key-value-row">
                    <span class="row-label">UDID:</span>
                    <span class="row-value break-all">{{ selectedDevice.udid }}</span>
                  </div>
                </div>
                <n-empty v-else-if="!isLoadingDevices && devices.length === 0" description="未发现设备，请连接并解锁设备后刷新" class="device-empty" />
              </div>
            </div>

            <div class="fluent-card installer-action-card">
              <div class="notice-box notice-warning">
                安装前请确保设备已解锁并信任此电脑，Windows 已安装 Apple Mobile Device 驱动，且 IPA 具有适用于目标设备的有效签名。
              </div>
              <n-button
                type="primary"
                size="large"
                :disabled="!selectedIPAPath || !selectedDeviceUDID || isLoadingDevices"
                :loading="isInstallingIPA"
                @click="handleInstallIPA"
              >
                {{ isInstallingIPA ? '正在安装...' : '安装 IPA 到所选设备' }}
              </n-button>
            </div>
          </div>
        </n-tab-pane>

        <!-- TAB 7: 设置 -->
        <n-tab-pane name="settings" tab="⚙️ 全局设置">
          <div class="tab-scroll-container">
            <div class="fluent-card centered-card">
              <h3 class="card-title">全局参数与配置</h3>

              <!-- Engine Status Card -->
              <div class="form-group">
                <label class="field-label">内置 ipatool / go-ios 引擎状态:</label>
                <div class="status-box">
                  <div class="status-left">
                    <div class="flex-align-center mb-1">
                      <span class="indicator-dot dot-active"></span>
                      <span class="font-bold text-sm">已加载依赖引擎 (支持 tools/ 独立目录与主程序同目录)</span>
                    </div>
                    <div class="text-xs text-gray-sub">{{ settings.ipaToolPath }}</div>
                  </div>
                  <n-tag type="success" size="small" round :bordered="false">引擎就绪</n-tag>
                </div>
              </div>

              <!-- Passphrase -->
              <div class="form-group">
                <label class="field-label">本地密钥库解锁密码 (--keychain-passphrase):</label>
                <n-input v-model:value="settings.keychainPassphrase" placeholder="输入本地密钥解锁密码（如 123456）" size="medium" />
                <div class="notice-box notice-green mt-2">
                  ✅ 核心功能说明：设置此密码后，IPATool 将自动在所有操作（搜索、版本查询、下载等）中以命令行参数传入此密码，彻底规避 Windows 终端下无法输入密码和直接卡死报错的 Bug！
                </div>
              </div>

              <!-- Proxy Settings -->
              <div class="form-group">
                <label class="field-label">网络代理配置 (HTTP / SOCKS5):</label>
                <div class="sub-card">
                  <n-checkbox v-model:checked="settings.enableProxy" class="mb-3">
                    启用网络代理 (通过环境变量传递给 ipatool 进程)
                  </n-checkbox>

                  <div class="form-group mb-2">
                    <label class="field-label text-xs">代理地址 (例如 http://127.0.0.1:10808):</label>
                    <n-input-group>
                      <n-input v-model:value="settings.proxyUrl" :disabled="!settings.enableProxy" placeholder="http://127.0.0.1:10808" size="medium" />
                      <n-button secondary :disabled="!settings.enableProxy" :loading="isTestingProxy" @click="handleTestProxy" size="medium">
                        测试代理连接
                      </n-button>
                    </n-input-group>
                  </div>
                  <div class="field-tip">* 国内访问 Apple App Store 认证接口通常需要代理，支持 Clash / v2rayN 等常见本地代理端口。</div>
                </div>
              </div>

              <!-- Default Download Dir -->
              <div class="form-group">
                <label class="field-label">默认 IPA 下载保存目录:</label>
                <n-input-group>
                  <n-input v-model:value="settings.defaultDownloadDir" placeholder="默认保存目录" size="medium" />
                  <n-button secondary size="medium" @click="handleSelectDefaultDir">浏览...</n-button>
                </n-input-group>
              </div>

              <!-- Default Platform -->
              <div class="form-group">
                <label class="field-label">默认下载平台:</label>
                <n-select v-model:value="settings.defaultPlatform" :options="platformOptions" size="medium" />
              </div>

              <div class="mt-4 flex-align-center" style="justify-content: space-between;">
                <n-button type="primary" size="large" @click="handleSaveSettings" style="height: 38px; padding: 0 24px;">
                  💾 保存并应用设置
                </n-button>
                <span class="text-xs text-gray-sub">果仓助手 (AppleVault) v1.0 · Wails v2 + Go + Vue 3</span>
              </div>
            </div>
          </div>
        </n-tab-pane>

      </n-tabs>
    </main>

    <!-- 3. Bottom GridSplitter Divider -->
    <div v-if="showLogs" class="panel-divider"></div>

    <!-- 4. Bottom Real-time Logs Console Drawer -->
    <footer class="console-drawer" :class="{ 'collapsed': !showLogs }">
      <!-- Status Header -->
      <div class="console-header">
        <div class="console-status-left">
          <n-spin v-if="isAnyOperationRunning" size="small" class="mr-2" />
          <span class="status-prefix">状态:</span>
          <span class="status-val">{{ statusText }}</span>
        </div>
        <div class="console-actions-right">
          <n-button v-if="isAnyOperationRunning" size="tiny" type="error" @click="handleCancel" class="mr-2">
            取消当前操作
          </n-button>
          <n-button v-if="showLogs" size="tiny" secondary class="btn-clear-log mr-2" @click="clearLogs">
            清空日志
          </n-button>
          <n-button size="tiny" secondary class="btn-toggle-log" @click="toggleLogs">
            {{ showLogs ? '🔽 收起日志' : '📜 展开日志' }}
          </n-button>
        </div>
      </div>

      <!-- Log Output Box -->
      <div v-show="showLogs" ref="logContainerRef" class="console-screen">
        <div v-for="(log, idx) in logLines" :key="idx" class="console-line" :class="{ 'line-error': log.includes('[ERR]') }">
          {{ log }}
        </div>
      </div>
    </footer>

    <!-- 5. 2FA Modal Dialog -->
    <n-modal v-model:show="show2FAModal" preset="card" title="🔐 Apple ID 双重认证" style="width: 400px;" :mask-closable="false">
      <div class="twofa-dialog-body">
        <p class="twofa-desc">已向你的受信任 Apple 设备发送了验证码。请输入收到的 6 位验证码以继续登录：</p>
        <n-input
          ref="twoFAInputRef"
          v-model:value="twoFACode"
          placeholder="6 位验证码"
          maxlength="6"
          size="large"
          class="twofa-code-input"
          @keydown.enter="confirm2FA"
        />
        <div class="twofa-actions">
          <n-button secondary @click="cancel2FA">取消</n-button>
          <n-button type="primary" :disabled="twoFACode.length !== 6" :loading="isLoggingIn" @click="confirm2FA">
            提交验证
          </n-button>
        </div>
      </div>
    </n-modal>

    <!-- 6. Target Version Modal Dialog -->
    <n-modal v-model:show="showTargetVersionModal" preset="card" title="🎯 跨步加速查询指定版本" style="width: 460px;">
      <div class="target-version-dialog-body">
        <p style="margin-bottom: 12px; font-size: 13px; color: #666; line-height: 1.6;">
          请输入目标版本号（例如：<code>10.2.80</code> 或 <code>8.0.0</code>）。程序采用<b>跨步跳跃与回溯加速算法</b>，每隔若干版本探测一次，当版本小于目标时自动向回精确锁定，大幅缩短查询耗时。
        </p>
        <n-input
          ref="targetVersionInputRef"
          v-model:value="targetVersionInput"
          placeholder="例如: 10.2.80"
          size="medium"
          clearable
          autofocus
          @keydown.enter="confirmStartTargetQuery"
        />
        <div style="margin-top: 14px; display: flex; align-items: center; justify-content: space-between;">
          <span style="font-size: 13px; color: #555;">跳跃步长 (每隔几个版本探测):</span>
          <n-input-number v-model:value="targetQueryStep" :min="2" :max="30" size="small" style="width: 110px;" />
        </div>
        <div style="margin-top: 16px; display: flex; justify-content: flex-end; gap: 8px;">
          <n-button secondary @click="showTargetVersionModal = false">取消</n-button>
          <n-button type="primary" :disabled="!targetVersionInput.trim()" @click="confirmStartTargetQuery">
            确定开始查询
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
  NTag,
  NSpace,
  NTabs,
  NTabPane,
  NCard,
  NInput,
  NInputNumber,
  NInputGroup,
  NSelect,
  NSwitch,
  NModal,
  NAlert,
  NCheckbox,
  NSpin,
  NPagination,
  NDataTable,
  NGrid,
  NGridItem,
  NForm,
  NFormItem,
  NProgress,
  NEmpty
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
  Download,
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
import { EventsOn, OnFileDrop, OnFileDropOff } from '../../wailsjs/runtime/runtime'
import { main } from '../../wailsjs/go/models'

const props = defineProps<{ isDark: boolean }>()
const emit = defineEmits<{ (e: 'toggleTheme'): void }>()

const message = useMessage()
const dialog = useDialog()

const activeTab = ref('account')
const isAnyOperationRunning = ref(false)
const statusText = ref('就绪')
const tableMaxHeight = ref(320)

// Settings
const settings = ref<main.Settings>({
  keychainPassphrase: '123456',
  defaultDownloadDir: 'D:\\Downloads',
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
  term: '支付宝',
  limit: 10,
  platform: 'iphone'
})
const searchResults = ref<main.AppItem[]>([])

const searchColumns = [
  { title: '应用名称', key: 'name', width: 220, ellipsis: true },
  { title: 'Bundle Identifier', key: 'bundleID', width: 220, ellipsis: true },
  { title: 'App ID', key: 'id', width: 120 },
  { title: '最新版本', key: 'version', width: 90 },
  { title: '价格', key: 'displayPrice', width: 80 },
  {
    title: '操作',
    key: 'actions',
    width: 230,
    render(row: main.AppItem) {
      return h(NSpace, { size: 6 }, () => [
        h(NButton, { size: 'tiny', type: 'primary', onClick: () => selectAppForVersions(row) }, () => '历史版本'),
        h(NButton, { size: 'tiny', secondary: true, onClick: () => downloadFromSearch(row) }, () => '直接下载'),
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
const targetQueryStep = ref<number>(6)
const versionForm = ref({
  bundleId: 'com.alipay.iphoneclient',
  appId: 0,
  appName: '支付宝',
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
  statusText.value = `正在查询版本 ID ${row.versionId} 的版本详情...`

  try {
    const res = await GetVersionMetadata(versionForm.value.bundleId, row.versionId, versionForm.value.appId)
    row.displayVersion = res.displayVersion
    row.fileSize = res.displayFileSize
    row.releaseDate = res.releaseDate ? new Date(res.releaseDate).toLocaleDateString() : '-'
    statusText.value = `版本 ID ${row.versionId} 对应版本号: ${res.displayVersion} (体积: ${res.displayFileSize}, 发布日期: ${row.releaseDate})`
  } catch (err: any) {
    message.error(`查询版本详情失败: ${err}`)
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
    message.success(`已添加下载任务: ${appName} (${ver})`)
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
  { title: '构建 ID', key: 'versionId', width: 220 },
  {
    title: '对应版本号',
    key: 'displayVersion',
    width: 140,
    render(row: VersionItem) {
      if (row.displayVersion === '未查询') {
        return h('span', { style: 'color: #888; font-style: italic;' }, '未查询')
      }
      return h('span', { style: 'font-weight: 600; color: #107C41;' }, row.displayVersion)
    }
  },
  { title: '文件体积', key: 'fileSize', width: 130 },
  { title: '发布日期', key: 'releaseDate', width: 150 },
  {
    title: '操作',
    key: 'actions',
    width: 220,
    render(row: VersionItem) {
      return h(NSpace, { size: 6 }, () => [
        h(NButton, {
          size: 'tiny',
          secondary: true,
          disabled: isListingVersions.value || isBatchQuerying.value || isTargetQuerying.value,
          loading: row.isQuerying,
          onClick: () => querySingleVersionMetadata(row)
        }, () => '查询详情'),
        h(NButton, {
          size: 'tiny',
          type: 'primary',
          onClick: () => downloadFromVersions(row)
        }, () => '一键下载此版本')
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
const purchasedPage = ref(1)
const purchasedTotal = ref(0)
const purchasedApps = ref<main.AppItem[]>([])

// IPA Installer State
const selectedIPAPath = ref('')
const devices = ref<main.DeviceInfo[]>([])
const selectedDeviceUDID = ref<string | null>(null)
const isLoadingDevices = ref(false)
const isInstallingIPA = ref(false)
const selectedIPAFileName = computed(() => selectedIPAPath.value.split(/[\\/]/).pop() || selectedIPAPath.value)
const selectedDevice = computed(() => devices.value.find(device => device.udid === selectedDeviceUDID.value))
const deviceOptions = computed(() => devices.value.map(device => {
  const details = [device.productType, device.productVersion ? `iOS ${device.productVersion}` : '', device.connectionType]
    .filter(Boolean)
    .join(' · ')
  return {
    label: details ? `${device.name} (${details})` : device.name,
    value: device.udid
  }
}))

const purchasedColumns = [
  { title: '应用名称', key: 'name', width: 220, ellipsis: true },
  { title: 'Bundle Identifier', key: 'bundleID', width: 220, ellipsis: true },
  { title: 'App ID', key: 'id', width: 120 },
  { title: '获取时间', key: 'purchaseDate', width: 160 },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render(row: main.AppItem) {
      return h(NSpace, { size: 6 }, () => [
        h(NButton, { size: 'tiny', type: 'primary', onClick: () => selectAppForVersions(row) }, () => '历史版本'),
        h(NButton, { size: 'tiny', secondary: true, onClick: () => downloadFromPurchased(row) }, () => '直接下载')
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
  statusText.value = '正在提交 2FA 验证码并完成登录...'

  try {
    const res = await Login(loginForm.value.email, loginForm.value.password, twoFACode.value)
    if (res.success) {
      show2FAModal.value = false
      account.value = res.account
      message.success(`登录成功！用户: ${res.account.name}`)
      loginForm.value.password = ''
      twoFACode.value = ''
      statusText.value = `登录成功: ${res.account.name}`
      activeTab.value = 'search'
    } else {
      message.error(`2FA 验证失败: ${res.errorMessage}`)
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
  statusText.value = '用户取消了双重认证验证。'
}

async function handleRevoke() {
  dialog.warning({
    title: '确认退出',
    content: '确定要退出当前账号并清除登录凭据吗？',
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
    content: '此操作将删除本地用户目录下的 .ipatool 密钥文件夹 (%USERPROFILE%\\.ipatool)，重置所有本地缓存。确定要清理吗？',
    positiveText: '确定清理',
    negativeText: '取消',
    onPositiveClick: async () => {
      isClearing.value = true
      try {
        await ClearKeychainCache()
        message.success('本地密钥缓存已彻底清除！')
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
      message.info('你已经拥有该应用的许可，无需重复购买。')
    } else {
      message.success('获取免费许可成功！')
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
  statusText.value = `正在查询 ${versionForm.value.bundleId} 的历史版本构建 ID 列表...`

  try {
    const res = await ListVersions(versionForm.value.bundleId, versionForm.value.appId)
    versionItems.value = (res.externalVersionIdentifiers || []).map(id => ({
      versionId: id,
      displayVersion: '未查询',
      fileSize: '-',
      releaseDate: '-'
    }))
    statusText.value = `共获取到 ${versionItems.value.length} 个历史版本构建 ID。`
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
    item.releaseDate = res.releaseDate ? new Date(res.releaseDate).toLocaleDateString() : '-'
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
        statusText.value = '批量版本查询已手动停止。'
        message.info('批量查询已停止')
        break
      }

      const item = versionItems.value[i]
      if (item.displayVersion !== '未查询') continue

      statusText.value = `批量查询中 (${i + 1}/${count}): ${item.versionId}...`
      await fetchItemMetadata(i)

      if (shouldStopBatchQuery.value) break
      if (!await cancellableSleep(120, shouldStopBatchQuery)) break
    }

    if (!shouldStopBatchQuery.value) {
      statusText.value = '批量版本查询完成。'
      message.success('批量查询完成')
    }
  } catch (err: any) {
    statusText.value = `批量版本查询失败: ${err}`
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
    title: '批量查询确认',
    content: `即将批量查询前 ${count} 个版本的详细版本号（每秒约查询 2 个），是否继续？`,
    positiveText: '开始查询',
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
    message.warning('请先点击「获取历史版本列表」')
    return
  }
  showTargetVersionModal.value = true
  nextTick(() => {
    targetVersionInputRef.value?.focus()
  })
}

function stopTargetQuery() {
  shouldStopTargetQuery.value = true
  statusText.value = '正在停止指定版本查询...'
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
  statusText.value = `启动跨步加速查询目标版本: ${targetVersion}...`

  const total = versionItems.value.length
  const step = Math.max(2, targetQueryStep.value || 6)
  let found = false
  let foundItem: VersionItem | null = null

  try {
    // 步骤 1: 先查询第 0 个（最新版本）
    statusText.value = `[跨步探测] 查询最新版本 (1/${total})...`
    const v0 = await fetchItemMetadata(0)
    if (shouldStopTargetQuery.value) {
      statusText.value = `已停止查询指定版本 (${targetVersion})。`
      return
    }

    if (v0 && isVersionMatch(v0, targetVersion)) {
      found = true
      foundItem = versionItems.value[0]
    } else if (v0 && compareVersions(v0, targetVersion) < 0) {
      // 最新版本都已经小于目标版本，说明目标版本比最新版还要新，不存在
      statusText.value = `当前最新版本 (${v0}) 小于目标版本 ${targetVersion}，历史版本中不存在该版本。`
      message.warning(`当前最新版本为 ${v0}，未找到更高版本 ${targetVersion}`)
      return
    }

    if (!found) {
      let prevIndex = 0
      let currIndex = 0

      // 步骤 2: 按步长跨步跳跃向前探测 (列表从上到下单调递减)
      while (currIndex < total - 1 && !shouldStopTargetQuery.value) {
        prevIndex = currIndex
        currIndex = Math.min(currIndex + step, total - 1)

        statusText.value = `[跨步探测] 跳跃检查第 ${currIndex + 1}/${total} 个版本 (构建 ID: ${versionItems.value[currIndex].versionId})...`
        const vCurr = await fetchItemMetadata(currIndex)

        if (shouldStopTargetQuery.value) break
        if (!await cancellableSleep(120, shouldStopTargetQuery)) break

        if (vCurr && isVersionMatch(vCurr, targetVersion)) {
          found = true
          foundItem = versionItems.value[currIndex]
          break
        }

        // 核心判断：当前探测版本小于目标版本，说明已经跳过了目标所在区间！
        if (vCurr && compareVersions(vCurr, targetVersion) < 0) {
          statusText.value = `⚡ 当前版本 (${vCurr}) < 目标 (${targetVersion})，立即向回查找区间 [${prevIndex + 2} ~ ${currIndex}]...`

          // 步骤 3: 往回找（从 currIndex - 1 倒查到 prevIndex + 1）
          for (let b = currIndex - 1; b > prevIndex; b--) {
            if (shouldStopTargetQuery.value) break

            statusText.value = `[向回查找] 正在精确核对第 ${b + 1}/${total} 个版本 (构建 ID: ${versionItems.value[b].versionId})...`
            const vBack = await fetchItemMetadata(b)

            if (shouldStopTargetQuery.value) break
            if (!await cancellableSleep(120, shouldStopTargetQuery)) break

            if (vBack && isVersionMatch(vBack, targetVersion)) {
              found = true
              foundItem = versionItems.value[b]
              break
            }

            // 若往回找时已经找到了比 target 大的版本，说明该区间内版本已跨越且无完全匹配项
            if (vBack && compareVersions(vBack, targetVersion) > 0) {
              break
            }
          }
          // 回溯结束跳出主循环
          break
        }

        // 已到达列表末尾但还没小于目标
        if (currIndex === total - 1) {
          for (let b = currIndex - 1; b > prevIndex; b--) {
            if (shouldStopTargetQuery.value) break
            statusText.value = `[末段向回查找] 正在核对第 ${b + 1}/${total} 个版本...`
            const vBack = await fetchItemMetadata(b)
            if (shouldStopTargetQuery.value) break
            if (!await cancellableSleep(120, shouldStopTargetQuery)) break

            if (vBack && isVersionMatch(vBack, targetVersion)) {
              found = true
              foundItem = versionItems.value[b]
              break
            }
          }
          break
        }
      }
    }

    if (shouldStopTargetQuery.value) {
      statusText.value = `已停止查询指定版本 (${targetVersion})。`
      message.info('已停止查询指定版本')
    } else if (found && foundItem) {
      statusText.value = `🎯 已命中目标版本 ${foundItem.displayVersion} (构建 ID: ${foundItem.versionId})！`
      message.success(`已查询到指定版本 ${foundItem.displayVersion}！`)
      versionForm.value.filter = targetVersion
    } else {
      statusText.value = `查询完成，未在历史记录中找到指定版本: ${targetVersion}`
      message.warning(`已完成跨步与回溯检索，未找到版本号: ${targetVersion}`)
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
    message.success('已清空所有已完成任务记录')
  } catch (err: any) {
    message.error(`操作失败: ${err}`)
  }
}

function handleOpenDefaultDownloadDir() {
  OpenInExplorer(settings.value.defaultDownloadDir || 'D:\\Downloads')
}

async function handleCancelTask(id: string) {
  try {
    await CancelDownloadTask(id)
    message.info('正在取消下载任务...')
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

function handleOpenFile(path: string) {
  if (!path) return
  OpenInExplorer(path)
}

function handleOpenDir(path: string) {
  if (!path) return
  OpenInExplorer(path)
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
  statusText.value = `已选择 IPA: ${ipaPath}`
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
  statusText.value = `正在安装 ${selectedIPAFileName.value}...`
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
    const res = await ListPurchases(purchasedPage.value, 20)
    purchasedApps.value = res.apps || []
    purchasedTotal.value = res.totalCount || 0
    statusText.value = `已购应用加载完成 (第 ${purchasedPage.value} 页，共 ${res.totalCount} 个)。`
  } catch (err: any) {
    message.error(`加载已购列表失败: ${err}`)
    statusText.value = '加载失败'
  } finally {
    isPurchasedLoading.value = false
    isAnyOperationRunning.value = false
  }
}

function prevPurchasedPage() {
  if (purchasedPage.value > 1) {
    purchasedPage.value--
    loadPurchases()
  }
}

function nextPurchasedPage() {
  if (purchasedPage.value * 20 < purchasedTotal.value) {
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

function handleCancel() {
  shouldStopBatchQuery.value = true
  shouldStopTargetQuery.value = true
  CancelRunningCommand()
  isAnyOperationRunning.value = false
}

// Lifecycle
onMounted(() => {
  loadSettings()
  refreshAccount(true)
  loadDownloadTasks()

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
.app-layout {
  display: flex;
  flex-direction: column;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background-color: #f3f4f6;
  color: #1f2937;
  font-family: Segoe UI, "Microsoft YaHei UI", sans-serif;
}

.dark-mode {
  background-color: #18181b;
  color: #e4e4e7;
}

/* 1. Header */
.top-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 20px;
  background-color: #ffffff;
  border-bottom: 1px solid #e5e7eb;
  flex-shrink: 0;
}

.dark-mode .top-header {
  background-color: #1f1f23;
  border-bottom-color: #2e2e32;
}

.header-brand {
  display: flex;
  align-items: center;
}

.brand-emoji {
  font-size: 20px;
  margin-right: 6px;
}

.brand-title {
  font-size: 18px;
  font-weight: bold;
  color: #0078d4;
  margin-right: 8px;
}

.brand-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 6px;
  background-color: rgba(0, 120, 212, 0.12);
  color: #0078d4;
  margin-right: 2px;
  line-height: 1.4;
  vertical-align: middle;
}

.dark-mode .brand-badge {
  background-color: rgba(0, 120, 212, 0.25);
  color: #4daafc;
}

.brand-subtitle {
  font-size: 13px;
  color: #6b7280;
  margin-left: 10px;
}

.dark-mode .brand-subtitle {
  color: #9ca3af;
}

.header-tools {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pill-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background-color: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  padding: 4px 12px;
  font-size: 12px;
}

.dark-mode .pill-badge {
  background-color: #27272a;
  border-color: #3f3f46;
}

.pill-label {
  font-weight: 600;
  color: #4b5563;
}

.dark-mode .pill-label {
  color: #a1a1aa;
}

.pill-value {
  color: #111827;
}

.dark-mode .pill-value {
  color: #f4f4f5;
}

.pill-sub {
  color: #6b7280;
  font-size: 11px;
}

.indicator-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}

.dot-active {
  background-color: #10b981;
}

.dot-inactive {
  background-color: #ef4444;
}

/* 2. Main Body Tabs */
.main-body {
  flex: 1;
  overflow: hidden;
  padding: 8px 18px 4px 18px;
  display: flex;
  flex-direction: column;
}

:deep(.custom-tabs) {
  display: flex;
  flex-direction: column;
  height: 100%;
}

:deep(.custom-tabs > .n-tabs-nav) {
  padding-bottom: 4px;
  flex-shrink: 0;
}

:deep(.custom-tabs > .n-tabs-pane-wrapper) {
  flex: 1;
  overflow: hidden;
}

:deep(.custom-tabs > .n-tabs-pane-wrapper > .n-tab-pane) {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 0;
}

.tab-scroll-container {
  flex: 1;
  height: 100%;
  overflow-y: auto;
  padding: 8px 2px 14px 2px;
}

.tab-table-container {
  flex: 1;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 8px 2px 4px 2px;
}

/* Cards */
.fluent-card {
  background-color: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 18px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}

.dark-mode .fluent-card {
  background-color: #202023;
  border-color: #333338;
}

.toolbar-card {
  padding: 12px;
  margin-bottom: 10px;
  flex-shrink: 0;
}

.table-card {
  flex: 1;
  padding: 0;
  overflow: hidden;
}

.card-title {
  margin: 0 0 14px 0;
  font-size: 16px;
  font-weight: bold;
}

.two-columns-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.centered-card {
  max-width: 700px;
  margin: 0 auto;
}

.installer-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}

.installer-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.installer-card-header .card-title {
  margin-bottom: 0;
}

.ipa-drop-zone {
  --wails-drop-target: drop;
  min-height: 180px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 24px;
  border: 2px dashed #93c5fd;
  border-radius: 8px;
  background-color: #eff6ff;
  cursor: pointer;
  text-align: center;
  transition: border-color 0.2s ease, background-color 0.2s ease;
}

.ipa-drop-zone:hover,
.ipa-drop-zone.has-file {
  border-color: #0078d4;
  background-color: #dbeafe;
}

.dark-mode .ipa-drop-zone {
  border-color: #2563eb;
  background-color: #172554;
}

.dark-mode .ipa-drop-zone:hover,
.dark-mode .ipa-drop-zone.has-file {
  border-color: #60a5fa;
  background-color: #1e3a8a;
}

.drop-icon {
  font-size: 42px;
  margin-bottom: 10px;
}

.drop-title {
  font-size: 15px;
  font-weight: 700;
}

.drop-subtitle {
  max-width: 100%;
  margin-top: 8px;
  overflow: hidden;
  color: #6b7280;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dark-mode .drop-subtitle {
  color: #bfdbfe;
}

.device-detail-card {
  margin-top: 14px;
  padding: 12px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  background-color: #f9fafb;
}

.dark-mode .device-detail-card {
  border-color: #3f3f46;
  background-color: #27272a;
}

.device-detail-card .key-value-row:last-child {
  margin-bottom: 0;
}

.device-empty {
  padding: 24px 0;
}

.installer-action-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.installer-action-card .notice-box {
  flex: 1;
  margin-bottom: 0;
}

@media (max-width: 1000px) {
  .installer-layout {
    grid-template-columns: 1fr;
  }
}

/* Form Styles */
.form-group {
  margin-bottom: 12px;
}

.field-label {
  display: block;
  font-size: 13px;
  color: #6b7280;
  margin-bottom: 4px;
}

.dark-mode .field-label {
  color: #9ca3af;
}

.field-tip {
  font-size: 11px;
  color: #6b7280;
  margin-top: 4px;
}

.key-value-row {
  display: flex;
  margin-bottom: 10px;
  font-size: 13px;
}

.row-label {
  width: 90px;
  color: #6b7280;
}

.dark-mode .row-label {
  color: #9ca3af;
}

.row-value {
  flex: 1;
}

/* Notice Box */
.notice-box {
  padding: 10px 12px;
  border-radius: 4px;
  font-size: 12px;
  line-height: 1.5;
  margin-bottom: 14px;
}

.notice-warning {
  background-color: #fef3c7;
  border: 1px solid #fde68a;
  color: #92400e;
}

.dark-mode .notice-warning {
  background-color: #451a03;
  border-color: #78350f;
  color: #fde68a;
}

.notice-gray {
  background-color: #f3f4f6;
  color: #4b5563;
}

.dark-mode .notice-gray {
  background-color: #27272a;
  color: #d4d4d8;
}

.notice-success {
  background-color: #ecfdf5;
  border: 1px solid #a7f3d0;
  color: #065f46;
}

.dark-mode .notice-success {
  background-color: #064e3b;
  border-color: #047857;
  color: #a7f3d0;
}

.notice-green {
  background-color: #ecfdf5;
  border: 1px solid #a7f3d0;
  color: #065f46;
}

.dark-mode .notice-green {
  background-color: #064e3b;
  border-color: #047857;
  color: #a7f3d0;
}

/* Toolbars */
.search-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
}

.version-toolbar-rows {
  display: flex;
  flex-direction: column;
}

.toolbar-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.purchased-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pagination-area {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pagination-info {
  font-size: 13px;
  color: #4b5563;
  font-weight: 500;
}

.label-inline {
  font-size: 13px;
  color: #6b7280;
  white-space: nowrap;
}

.select-wrapper {
  display: flex;
  align-items: center;
  gap: 4px;
}

.count-tag {
  font-size: 12px;
  color: #6b7280;
  margin-left: 6px;
}

.btn-group-row {
  display: flex;
  gap: 10px;
}

.sub-card {
  background-color: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  padding: 12px;
}

.dark-mode .sub-card {
  background-color: #1f1f23;
  border-color: #333338;
}

.status-box {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  padding: 10px 14px;
}

.dark-mode .status-box {
  background-color: #1f1f23;
  border-color: #333338;
}

.flex-align-center {
  display: flex;
  align-items: center;
}

.mr-2 { margin-right: 8px; }
.mt-2 { margin-top: 8px; }
.mt-3 { margin-top: 12px; }
.mt-4 { margin-top: 16px; }
.mb-1 { margin-bottom: 4px; }
.mb-2 { margin-bottom: 8px; }
.mb-3 { margin-bottom: 12px; }
.flex-1 { flex: 1; }
.font-bold { font-weight: bold; }
.font-semibold { font-weight: 600; }
.text-xs { font-size: 11px; }
.text-sm { font-size: 13px; }
.text-gray-sub { color: #6b7280; }
.break-all { word-break: break-all; }

.text-ellipsis {
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 3. Splitter Divider */
.panel-divider {
  height: 4px;
  background-color: #e5e7eb;
  cursor: ns-resize;
  flex-shrink: 0;
}

.dark-mode .panel-divider {
  background-color: #27272a;
}

/* 4. Console Drawer */
.console-drawer {
  height: 180px;
  display: flex;
  flex-direction: column;
  background-color: #1e1e1e;
  flex-shrink: 0;
  transition: height 0.2s ease;
}

.console-drawer.collapsed {
  height: 28px;
  background-color: #f9fafb;
}

.dark-mode .console-drawer.collapsed {
  background-color: #18181b;
}

.console-header {
  height: 28px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 12px;
  background-color: #f3f4f6;
  border-top: 1px solid #e5e7eb;
  flex-shrink: 0;
  font-size: 12px;
  transition: background-color 0.2s ease;
}

.dark-mode .console-header {
  background-color: #27272a;
  border-top-color: #3f3f46;
}

.console-drawer:not(.collapsed) .console-header {
  height: 32px;
  background-color: #252526;
  border-top-color: #333333;
  border-bottom: 1px solid #333333;
}

.console-status-left {
  display: flex;
  align-items: center;
  font-size: 12px;
}

.status-prefix {
  color: #6b7280;
  margin-right: 6px;
}

.dark-mode .status-prefix {
  color: #9ca3af;
}

.console-drawer:not(.collapsed) .status-prefix {
  color: #858585;
}

.status-val {
  color: #374151;
  font-weight: 600;
}

.dark-mode .status-val {
  color: #e4e4e7;
}

.console-drawer:not(.collapsed) .status-val {
  color: #cccccc;
}

.console-actions-right {
  display: flex;
  align-items: center;
}

.btn-toggle-log {
  font-size: 11px !important;
  height: 20px !important;
  padding: 0 6px !important;
}

.btn-clear-log {
  font-size: 11px !important;
  height: 20px !important;
  padding: 0 6px !important;
}

.console-drawer:not(.collapsed) .btn-clear-log,
.console-drawer:not(.collapsed) .btn-toggle-log {
  background-color: #333333 !important;
  color: #cccccc !important;
  border: none !important;
}

.console-screen {
  flex: 1;
  padding: 8px 12px;
  overflow-y: auto;
  font-family: Consolas, Cascadia Code, "Courier New", monospace;
  font-size: 12px;
  line-height: 1.5;
  color: #d4d4d4;
  background-color: #1e1e1e;
}

.console-line {
  word-break: break-all;
  white-space: pre-wrap;
}

.line-error {
  color: #f87171;
}

/* 2FA Modal */
.twofa-dialog-body {
  padding: 8px 0;
}

.twofa-desc {
  font-size: 13px;
  color: #6b7280;
  margin-bottom: 16px;
  line-height: 1.5;
}

.twofa-code-input {
  text-align: center;
  font-size: 24px;
  font-weight: bold;
  letter-spacing: 8px;
  margin-bottom: 20px;
}

.twofa-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

/* Task Manager Styles */
.empty-tasks-box {
  padding: 60px 0;
  display: flex;
  justify-content: center;
  align-items: center;
}

.task-items-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.task-card {
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  background-color: #ffffff;
  padding: 12px 16px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03);
}

.dark-mode .task-card {
  border-color: #333338;
  background-color: #26262a;
}

.task-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.task-info {
  flex: 1;
  min-width: 0;
}

.task-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.task-app-name {
  font-size: 14px;
  font-weight: bold;
}

.task-ver-tag {
  font-weight: 600;
  font-size: 11px;
}

.task-bundle-id {
  font-size: 11px;
  color: #6b7280;
}

.task-build-id {
  font-size: 11px;
  color: #9ca3af;
  background: #f3f4f6;
  padding: 1px 6px;
  border-radius: 3px;
}

.dark-mode .task-build-id {
  background: #18181b;
  color: #a1a1aa;
}

.task-progress-bar {
  margin-bottom: 6px;
}

.task-meta-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 11px;
  color: #6b7280;
}

.dark-mode .task-meta-row {
  color: #9ca3af;
}

.task-meta-left {
  display: flex;
  align-items: center;
  gap: 10px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-badge-status {
  padding: 1px 6px;
  border-radius: 3px;
  font-weight: 600;
  font-size: 11px;
}

.badge-pending {
  background: #fef3c7;
  color: #b45309;
}

.badge-downloading {
  background: #e0f2fe;
  color: #0369a1;
}

.badge-completed {
  background: #dcfce7;
  color: #15803d;
}

.badge-error {
  background: #fee2e2;
  color: #b91c1c;
}

.badge-canceled {
  background: #f3f4f6;
  color: #4b5563;
}

.task-speed {
  color: #0078d4;
}

.task-bytes {
  color: #4b5563;
}

.dark-mode .task-bytes {
  color: #a1a1aa;
}

.task-path {
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-err-msg {
  color: #ef4444;
}

.task-meta-right {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.task-pct {
  color: #0078d4;
  font-size: 12px;
}

.task-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}
</style>
