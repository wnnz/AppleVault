<template>
  <div class="view-panel">
    <div class="view-header">
      <div>
        <div class="title-with-badge"><h2 class="view-title">账号中心</h2></div>
        <p class="view-desc">管理用于 App Store 正版授权通信与 IPA 下载的 Apple ID 身份凭证</p>
      </div>
    </div>

    <div class="two-columns-layout">
      <AppCard class="clean-card">
        <div class="card-headline">
          <span class="headline-title">当前登录凭证</span>
          <span class="account-pill" :class="isLoggedIn ? 'pill-success' : 'pill-gray'">{{ isLoggedIn ? '已授权' : '未登录' }}</span>
        </div>
        <div class="account-profile-box">
          <div class="large-avatar" :class="{ 'avatar-active': isLoggedIn }">{{ isLoggedIn ? (account.name ? account.name.charAt(0).toUpperCase() : '') : '' }}</div>
          <div class="large-profile-info">
            <div class="profile-name">{{ account.name || '尚未登录 Apple ID' }}</div>
            <div class="profile-email">{{ account.email || '请在右侧输入账号密码完成登录' }}</div>
          </div>
        </div>
        <div class="sub-alert-box mt-4"><span>官方直接认证：所有凭据直接向 Apple 官方接口请求并保存在本地钥匙串，不经过任何第三方服务器。</span></div>
        <div class="account-card-actions mt-4">
          <AppButton type="error" secondary :disabled="!isLoggedIn" :loading="isRevoking" @click="emit('revoke')">退出登录</AppButton>
          <AppButton secondary :loading="isClearing" title="清理本地钥匙串缓存解决校验异常" @click="emit('clearKeychain')">清理钥匙串缓存</AppButton>
          <AppButton secondary :loading="isAccountLoading" @click="emit('refresh')">刷新状态</AppButton>
        </div>
      </AppCard>

      <AppCard class="clean-card">
        <div class="card-headline"><span class="headline-title">登录 Apple ID</span></div>
        <div class="form-item-clean">
          <label class="clean-label">Apple ID 账户邮箱</label>
          <AppInput v-model="loginForm.email" size="small" placeholder="例如: your_apple_id@icloud.com" />
        </div>
        <div class="form-item-clean mt-3">
          <label class="clean-label">Apple ID 密码</label>
          <AppInput v-model="loginForm.password" type="password" size="small" placeholder="请输入 Apple ID 密码" @keydown.enter="emit('login')" />
        </div>
        <div class="sub-alert-box mt-3"><span>双重认证 (2FA)：若账号开启了 2FA，点击登录后将自动弹出 6 位验证码输入窗口。</span></div>
        <div class="mt-4">
          <AppButton type="primary" block size="medium" :loading="isLoggingIn" @click="emit('login')">登录</AppButton>
        </div>
      </AppCard>
    </div>

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
          size="large"
          class="text-center font-bold text-lg"
          placeholder="6 位验证码"
          maxlength="6"
          autofocus
          style="letter-spacing: 6px; font-size: 20px;"
          @keydown.enter="emit('confirm2FA')"
        />
        <div class="dialog-action-buttons mt-4">
          <AppButton secondary @click="emit('cancel2FA')">取消</AppButton>
          <AppButton type="primary" :disabled="twoFACode.length !== 6" :loading="isLoggingIn" @click="emit('confirm2FA')">提交验证</AppButton>
        </div>
      </div>
    </AppDialog>
  </div>
</template>

<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import AppButton from '../../ui/components/AppButton.vue'
import AppCard from '../../ui/components/AppCard.vue'
import AppDialog from '../../ui/components/AppDialog.vue'
import AppInput from '../../ui/components/AppInput.vue'

defineProps<{
  account: { name: string; email: string; success: boolean }
  isLoggedIn: boolean
  isAccountLoading: boolean
  isRevoking: boolean
  isClearing: boolean
  isLoggingIn: boolean
  loginForm: { email: string; password: string }
}>()

const show2FAModal = defineModel<boolean>('show2FAModal', { required: true })
const twoFACode = defineModel<string>('twoFACode', { required: true })
const twoFAInputRef = ref<InstanceType<typeof AppInput> | null>(null)

const emit = defineEmits<{
  revoke: []
  clearKeychain: []
  refresh: []
  login: []
  confirm2FA: []
  cancel2FA: []
}>()

watch(show2FAModal, async visible => {
  if (!visible) return
  await nextTick()
  twoFAInputRef.value?.focus()
})
</script>
