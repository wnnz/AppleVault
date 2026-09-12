<template>
  <div class="view-panel">
    <div class="view-header">
      <div>
        <div class="title-with-badge"><h2 class="view-title">账号中心</h2></div>
        <p class="view-desc">管理用于 App Store 正版授权通信与 IPA 下载的 Apple ID 身份凭证</p>
      </div>
    </div>

    <div class="two-columns-layout">
      <!-- 凭证卡片 -->
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
            :loading="isRevoking"
            @click="handleRevoke"
          >
            退出登录
          </AppButton>
          <AppButton
            secondary
            :loading="isClearing"
            title="清理本地钥匙串缓存解决校验异常"
            @click="handleClearKeychain"
          >
            清理钥匙串缓存
          </AppButton>
          <AppButton
            secondary
            :loading="isAccountLoading"
            @click="refreshAccount(false)"
          >
            刷新状态
          </AppButton>
        </div>
      </AppCard>

      <!-- 登录表单 -->
      <AppCard class="clean-card">
        <div class="card-headline"><span class="headline-title">登录 Apple ID</span></div>
        <div class="form-item-clean">
          <label class="clean-label">Apple ID 账户邮箱</label>
          <AppInput
            v-model="loginForm.email"
            size="small"
            placeholder="例如: your_apple_id@icloud.com"
            @keydown.enter="handleLogin"
          />
        </div>
        <div class="form-item-clean mt-3">
          <label class="clean-label">密码</label>
          <AppInput
            v-model="loginForm.password"
            type="password"
            size="small"
            placeholder="输入 Apple ID 账户密码"
            @keydown.enter="handleLogin"
          />
        </div>
        <div class="sub-alert-box mt-3">
          <span>安全声明：本软件完全开源，账号与密码仅在登录鉴权时直接发送给苹果官方服务器，绝不在本地或外部以任何形式上传明文。</span>
        </div>
        <div class="form-action-btn-row mt-4">
          <AppButton
            type="primary"
            size="large"
            :loading="isLoggingIn"
            class="full-width"
            @click="handleLogin"
          >
            登录并获取授权
          </AppButton>
        </div>
      </AppCard>
    </div>

    <!-- 2FA 验证码输入弹窗 -->
    <AppDialog v-model="show2FAModal" title="Apple ID 双重认证" style="width: 420px; border-radius: 14px;">
      <div class="modal-dialog-inner">
        <p class="dialog-desc">
          已向您的受信任受认证 Apple 设备发送了验证码，请输入 6 位验证码以完成登录。
        </p>
        <AppInput
          ref="twoFAInputRef"
          v-model="twoFACode"
          placeholder="6 位数字验证码"
          maxlength="6"
          autofocus
          class="dialog-code-input"
          @keydown.enter="confirm2FA"
        />
        <div class="dialog-action-buttons mt-4">
          <AppButton secondary @click="cancel2FA">取消</AppButton>
          <AppButton type="primary" :disabled="twoFACode.length !== 6" @click="confirm2FA">确认登录</AppButton>
        </div>
      </div>
    </AppDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { backend as main } from '../../../wailsjs/go/models'
import {
  GetAccountInfo,
  Login,
  Revoke,
  ClearKeychainCache
} from '../../../wailsjs/go/backend/App'
import AppButton from '../../ui/components/AppButton.vue'
import AppCard from '../../ui/components/AppCard.vue'
import AppDialog from '../../ui/components/AppDialog.vue'
import AppInput from '../../ui/components/AppInput.vue'
import { useAppDialog, useAppMessage } from '../../ui/feedback'

const emit = defineEmits<{
  (e: 'busyChange', busy: boolean, text?: string): void
  (e: 'loginSuccess', account: main.AccountInfo): void
  (e: 'accountChanged', account: main.AccountInfo): void
  (e: 'navigate', tab: string): void
}>()

const message = useAppMessage()
const dialog = useAppDialog()

// 账号状态
const account = ref<main.AccountInfo>({ name: '', email: '', success: false })
const isLoggedIn = computed(() => account.value.success && !!account.value.email)
const isAccountLoading = ref(false)
const isRevoking = ref(false)
const isClearing = ref(false)
const isLoggingIn = ref(false)

// 登录表单
const loginForm = ref({
  email: '',
  password: ''
})

// 2FA 弹窗
const show2FAModal = ref(false)
const twoFACode = ref('')
const twoFAInputRef = ref<InstanceType<typeof AppInput> | null>(null)

watch(show2FAModal, async visible => {
  if (!visible) return
  await nextTick()
  twoFAInputRef.value?.focus()
})

/**
 * 刷新当前账号登录状态
 * @param autoNavigate 是否根据登录状态自动跳转视图（已登录跳转到搜索页，未登录跳转到账号页）
 */
async function refreshAccount(autoNavigate = false) {
  isAccountLoading.value = true
  emit('busyChange', false, '正在获取账号信息...')
  try {
    const res = await GetAccountInfo()
    account.value = res
    emit('accountChanged', res)
    if (res.success && res.email) {
      localStorage.setItem('apple_vault_logged_in', 'true')
      emit('busyChange', false, `已登录: ${res.name} (${res.email})`)
      if (autoNavigate) {
        emit('navigate', 'search')
      }
    } else {
      localStorage.setItem('apple_vault_logged_in', 'false')
      emit('busyChange', false, '未检测到已登录的 Apple ID')
      if (autoNavigate) {
        emit('navigate', 'account')
      }
    }
  } catch {
    localStorage.setItem('apple_vault_logged_in', 'false')
    account.value = { name: '', email: '', success: false }
    emit('accountChanged', account.value)
    emit('busyChange', false, '未登录')
    if (autoNavigate) {
      emit('navigate', 'account')
    }
  } finally {
    isAccountLoading.value = false
  }
}

/**
 * 执行 Apple ID 账户与密码登录
 */
async function handleLogin() {
  if (!loginForm.value.email || !loginForm.value.password) {
    message.warning('请输入 Apple ID 邮箱和密码！')
    return
  }

  isLoggingIn.value = true
  emit('busyChange', true, '正在验证 Apple ID 账号与密码...')

  try {
    const res = await Login(loginForm.value.email, loginForm.value.password, '')
    if (res.requires2FA) {
      twoFACode.value = ''
      show2FAModal.value = true
      emit('busyChange', true, '等待输入双重认证验证码...')
      return
    }

    if (res.success) {
      localStorage.setItem('apple_vault_logged_in', 'true')
      account.value = res.account
      emit('accountChanged', res.account)
      emit('loginSuccess', res.account)
      message.success(`登录成功: ${res.account.name}`)
      loginForm.value.password = ''
      emit('busyChange', false, `登录成功: ${res.account.name}`)
      emit('navigate', 'search')
    } else {
      localStorage.setItem('apple_vault_logged_in', 'false')
      message.error(`登录失败: ${res.errorMessage}`)
      emit('busyChange', false, `登录失败: ${res.errorMessage}`)
    }
  } catch (err: any) {
    message.error(`执行异常: ${err}`)
  } finally {
    isLoggingIn.value = false
    if (!show2FAModal.value) {
      emit('busyChange', false)
    }
  }
}

/**
 * 确认提交 2FA 验证码完成登录
 */
async function confirm2FA() {
  if (twoFACode.value.length !== 6) {
    message.warning('请输入正确的 6 位数字验证码！')
    return
  }

  isLoggingIn.value = true
  emit('busyChange', true, '正在提交验证码并完成登录...')

  try {
    const res = await Login(loginForm.value.email, loginForm.value.password, twoFACode.value)
    if (res.success) {
      localStorage.setItem('apple_vault_logged_in', 'true')
      show2FAModal.value = false
      account.value = res.account
      emit('accountChanged', res.account)
      emit('loginSuccess', res.account)
      message.success(`登录成功: ${res.account.name}`)
      loginForm.value.password = ''
      twoFACode.value = ''
      emit('busyChange', false, `登录成功: ${res.account.name}`)
      emit('navigate', 'search')
    } else {
      localStorage.setItem('apple_vault_logged_in', 'false')
      message.error(`验证失败: ${res.errorMessage}`)
      emit('busyChange', false, `验证失败: ${res.errorMessage}`)
    }
  } catch (err: any) {
    message.error(`验证异常: ${err}`)
  } finally {
    isLoggingIn.value = false
    emit('busyChange', false)
  }
}

/**
 * 取消 2FA 验证
 */
function cancel2FA() {
  show2FAModal.value = false
  twoFACode.value = ''
  emit('busyChange', false, '用户取消了验证。')
}

/**
 * 注销登录凭据
 */
async function handleRevoke() {
  dialog.warning({
    title: '确认退出登录',
    content: '确定要退出当前 Apple ID 账号并清除本地授权凭证吗？',
    positiveText: '确定退出',
    negativeText: '取消',
    onPositiveClick: async () => {
      isRevoking.value = true
      emit('busyChange', true)
      try {
        const ok = await Revoke()
        if (ok) {
          localStorage.setItem('apple_vault_logged_in', 'false')
          account.value = { name: '', email: '', success: false }
          emit('accountChanged', account.value)
          message.success('已成功注销登录凭据')
          emit('busyChange', false, '已退出登录')
          emit('navigate', 'account')
        }
      } catch (err: any) {
        message.error(`注销失败: ${err}`)
      } finally {
        isRevoking.value = false
        emit('busyChange', false)
      }
    }
  })
}

/**
 * 清空本地钥匙串缓存
 */
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

defineExpose({
  account,
  isLoggedIn,
  refreshAccount
})
</script>
