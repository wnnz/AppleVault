<template>
  <div class="view-panel">
    <div class="view-header">
      <div>
        <div class="title-with-badge"><h2 class="view-title">关于软件</h2></div>
        <p class="view-desc">果仓助手 (AppleVault) 软件信息、项目开源地址与使用说明</p>
      </div>
    </div>

    <AppCard class="clean-card about-card-stack">
      <div class="about-hero">
        <div class="about-large-icon"><img class="about-logo-image" :src="logo" alt="AppleVault" /></div>
        <div class="about-hero-text">
          <div class="about-app-title">果仓助手 (AppleVault)</div>
          <div class="about-version-line"><span class="about-version-badge">版本 v1.1.0</span><span class="about-badge-sub">基于 Wails & Go 构建</span></div>
          <p class="about-intro">现代优雅的 Apple App Store 正版应用与历史版本下载管理工具，支持 iOS 设备一键直装。</p>
        </div>
      </div>

      <div class="about-section-box">
        <div class="about-section-label">GitHub 官方开源仓库</div>
        <div class="repo-link-bar">
          <AppInput model-value="https://github.com/wnnz/AppleVault" readonly size="small" class="flex-1" />
          <AppButton type="primary" size="small" @click="emit('openGitHub')">访问 GitHub</AppButton>
          <AppButton secondary size="small" class="copy-address-btn" @click="emit('copyGitHubUrl')">复制地址</AppButton>
        </div>
      </div>

      <div class="about-features-grid">
        <div class="about-feature-item"><div class="feature-title">官方正版授权</div><div class="feature-desc">使用自己的 Apple ID 账户直接对接 App Store，下载包含个人官方凭据的正版 IPA，装机稳定不闪退。</div></div>
        <div class="about-feature-item"><div class="feature-title">历史版本检索</div><div class="feature-desc">支持输入任意 Bundle ID 查询完整历史构建记录与版本详情，可指定目标版本快速查找并一键下载。</div></div>
        <div class="about-feature-item"><div class="feature-title">真机无缝安装</div><div class="feature-desc">连接 iPhone / iPad 数据线即可自动识别设备规格与系统版本，将有效签名的应用包直接安装到设备。</div></div>
        <div class="about-feature-item"><div class="feature-title">绿色纯净便携</div><div class="feature-desc">无需安装器与后台常驻，所有配置与下载文件按登录账号自动隔离归档在本地 data 目录中。</div></div>
      </div>

      <div class="about-footer-notice"><span>本项目仅为 App Store 官方接口的图形化辅助工具，与 Apple Inc. 无官方隶属关系。App Store 模块衍生自 MIT 许可的 ipatool，设备功能由 go-ios 提供支持。</span></div>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import AppButton from '../../ui/components/AppButton.vue'
import AppCard from '../../ui/components/AppCard.vue'
import AppInput from '../../ui/components/AppInput.vue'
import logo from '../../assets/applevault-logo.webp'

const emit = defineEmits<{
  openGitHub: []
  copyGitHubUrl: []
}>()
</script>
