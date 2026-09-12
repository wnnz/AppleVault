<template>
  <div class="view-panel">
    <div class="view-header">
      <div>
        <div class="title-with-badge"><h2 class="view-title">系统设置</h2></div>
        <p class="view-desc">全局参数、网络代理、下载路径与基础运行引擎配置</p>
      </div>
      <AppButton type="primary" size="large" class="settings-save-btn" @click="emit('save')">保存并应用设置</AppButton>
    </div>

    <AppCard class="clean-card settings-stack">
      <div class="settings-group">
        <div class="settings-group-title">底层引擎与环境</div>
        <div class="settings-field-row">
          <div class="field-meta">
            <div class="field-title">CLI 引擎状态</div>
            <div class="field-desc">App Store 服务已内置；自动检测 tools/ 目录与系统环境变量中的 go-ios</div>
          </div>
          <div class="field-control"><div class="engine-badge-box"><span class="engine-dot"></span><span class="engine-text">引擎就绪</span></div></div>
        </div>
      </div>

      <div class="settings-group">
        <div class="settings-group-title">安全与钥匙串</div>
        <div class="settings-field-row">
          <div class="field-meta">
            <div class="field-title">密钥库解锁密码 (--keychain-passphrase)</div>
            <div class="field-desc">自动注入命令行参数，彻底杜绝 Windows 终端弹窗与死锁卡死</div>
          </div>
          <div class="field-control">
            <AppInput v-model="settings.keychainPassphrase" type="password" size="small" placeholder="输入密钥库密码（默认 123456）" />
          </div>
        </div>
      </div>

      <div class="settings-group">
        <div class="settings-group-title">网络与代理</div>
        <div class="settings-field-row">
          <div class="field-meta">
            <div class="field-title">启用网络代理</div>
            <div class="field-desc">App Store 接口认证时通过环境变量透传 HTTP / SOCKS5 代理</div>
          </div>
          <div class="field-control"><AppSwitch v-model="settings.enableProxy" /></div>
        </div>
        <div class="settings-field-row">
          <div class="field-meta">
            <div class="field-title">代理服务器地址</div>
            <div class="field-desc">支持 Clash / v2rayN 等常见本地代理端口（如 http://127.0.0.1:10808）</div>
          </div>
          <div class="field-control">
            <div class="setting-input-action-group">
              <AppInput v-model="settings.proxyUrl" :disabled="!settings.enableProxy" size="small" class="flex-1" placeholder="http://127.0.0.1:10808" />
              <AppButton variant="secondary" size="small" class="setting-fixed-btn" :disabled="!settings.enableProxy" :loading="isTestingProxy" @click="emit('testProxy')">测试</AppButton>
            </div>
          </div>
        </div>
      </div>

      <div class="settings-group">
        <div class="settings-group-title">存储与下载</div>
        <div class="settings-field-row">
          <div class="field-meta">
            <div class="field-title">IPA 存储路径</div>
            <div class="field-desc">按登录账号动态归档至 data/downloads/账号标识 目录</div>
          </div>
          <div class="field-control">
            <div class="setting-input-action-group">
              <AppInput :model-value="settings.defaultDownloadDir" readonly disabled size="small" class="flex-1" />
              <AppButton variant="secondary" size="small" class="setting-fixed-btn" @click="emit('openDownloadDir')">打开</AppButton>
            </div>
          </div>
        </div>
        <div class="settings-field-row">
          <div class="field-meta">
            <div class="field-title">默认目标平台</div>
            <div class="field-desc">检索与下载默认针对的 Apple 硬件体系</div>
          </div>
          <div class="field-control">
            <AppSelect v-model="settings.defaultPlatform" :options="platformOptions" size="small" style="width: 100%;" />
          </div>
        </div>
      </div>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import AppButton from '../../ui/components/AppButton.vue'
import AppCard from '../../ui/components/AppCard.vue'
import AppInput from '../../ui/components/AppInput.vue'
import AppSelect, { type AppSelectOption } from '../../ui/components/AppSelect.vue'
import AppSwitch from '../../ui/components/AppSwitch.vue'

defineProps<{
  settings: {
    keychainPassphrase: string
    defaultDownloadDir: string
    defaultPlatform: string
    enableProxy: boolean
    proxyUrl: string
    ipaToolPath: string
  }
  platformOptions: AppSelectOption[]
  isTestingProxy: boolean
}>()

const emit = defineEmits<{
  save: []
  testProxy: []
  openDownloadDir: []
}>()
</script>
