<template>
  <div class="view-panel">
    <div class="view-header">
      <div>
        <div class="title-with-badge"><h2 class="view-title">账号中心</h2></div>
        <p class="view-desc">管理用于 App Store 正版授权通信与 IPA 下载的 Apple ID 身份凭证</p>
      </div>
    </div>

    <AppCard class="clean-card account-keychain-card mb-4">
      <div class="account-keychain-copy">
        <div class="card-headline"><span class="headline-title">钥匙串密码</span></div>
        <div class="field-desc">用于加密本地 Apple ID 凭据。已有账号依赖当前密码，请勿随意修改；已有配置中的密码会继续保留。</div>
      </div>
      <div class="account-keychain-control">
        <AppInput
          v-model="keychainPassphrase"
          type="password"
          size="small"
          placeholder="请设置钥匙串密码"
          @keydown.enter="handleSaveKeychainPassphrase"
        />
        <AppButton
          type="primary"
          size="small"
          :loading="isSavingKeychainPassphrase"
          @click="handleSaveKeychainPassphrase"
        >
          保存密码
        </AppButton>
      </div>
    </AppCard>

    <AppCard class="clean-card account-list-card">
      <div class="account-list-toolbar">
        <div class="card-headline account-list-heading">
          <span class="headline-title">已登录账号</span>
          <span class="account-pill" :class="isLoggedIn ? 'pill-success' : 'pill-gray'">
            {{ accounts.length }} 个账号
          </span>
        </div>
        <AppButton type="primary" size="small" @click="openLoginModal">
          添加账号
        </AppButton>
      </div>
      <div v-if="accounts.length" class="account-list">
        <div
          v-for="item in accounts"
          :key="item.id"
          class="account-list-item"
          :class="{ active: item.active }"
        >
          <button
            type="button"
            class="account-list-main"
            :disabled="isSwitching || isAccountActionRunning"
            @click="handleSwitchAccount(item)"
          >
            <span class="account-list-avatar">{{ item.name ? item.name.charAt(0).toUpperCase() : '' }}</span>
            <span class="account-list-profile">
              <span class="account-list-name">{{ item.name || 'Apple ID' }}</span>
              <span class="account-list-email">{{ item.email }}</span>
            </span>
            <span v-if="item.region" class="account-region-pill">{{ item.region }}</span>
            <span v-if="item.active" class="account-current-label">当前</span>
          </button>
          <div class="account-row-actions">
            <AppButton
              type="error"
              secondary
              size="tiny"
              :disabled="isAccountActionRunning"
              :loading="revokingAccountID === item.id"
              @click="handleRevoke(item)"
            >
              退出登录
            </AppButton>
            <AppButton
              secondary
              size="tiny"
              :disabled="isAccountActionRunning"
              :loading="clearingAccountID === item.id"
              title="清理该账号的本地钥匙串凭据与 Cookie"
              @click="handleClearKeychain(item)"
            >
              清理钥匙串缓存
            </AppButton>
            <AppButton
              secondary
              size="tiny"
              :disabled="isAccountActionRunning"
              :loading="refreshingAccountID === item.id"
              @click="handleRefreshAccount(item)"
            >
              刷新状态
            </AppButton>
          </div>
        </div>
      </div>
      <div v-else class="account-empty-state">
        尚未登录 Apple ID，点击上方“添加账号”开始登录。
      </div>
      <div class="sub-alert-box mt-4">
        <span>官方直接认证：所有凭据直接向 Apple 官方接口请求并保存在本地钥匙串，不经过任何第三方服务器。</span>
      </div>
    </AppCard>

    <!-- 添加账号弹窗 -->
    <AppDialog v-model="showLoginModal" title="添加 Apple ID" style="width: 440px; border-radius: 14px;">
      <div class="modal-dialog-inner">
        <div class="form-item-clean">
          <label class="clean-label">Apple ID 账户邮箱</label>
          <AppInput
            ref="loginEmailInputRef"
            v-model="loginForm.email"
            placeholder="例如: your_apple_id@icloud.com"
            autofocus
            @keydown.enter="handleLogin"
          />
        </div>
        <div class="form-item-clean mt-3">
          <label class="clean-label">密码</label>
          <AppInput
            v-model="loginForm.password"
            type="password"
            placeholder="输入 Apple ID 账户密码"
            @keydown.enter="handleLogin"
          />
        </div>
        <div class="sub-alert-box mt-3">
          <span>安全声明：本软件完全开源，账号与密码仅在登录鉴权时直接发送给苹果官方服务器，绝不在本地或外部以任何形式上传明文。</span>
        </div>
        <div class="dialog-action-buttons mt-4">
          <AppButton secondary :disabled="isLoggingIn" @click="showLoginModal = false">取消</AppButton>
          <AppButton type="primary" :loading="isLoggingIn" @click="handleLogin">登录并添加</AppButton>
        </div>
      </div>
    </AppDialog>

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
  GetAccounts,
  Login,
  RefreshAccount,
  RemoveAccount,
  SwitchAccount,
  GetSettings,
  SetKeychainPassphrase
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
const account = ref<main.AccountInfo>({ id: '', name: '', email: '', region: '', active: false, success: false })
const accounts = ref<main.AccountInfo[]>([])
const isLoggedIn = computed(() => account.value.success && !!account.value.email)
const isAccountLoading = ref(false)
const isLoggingIn = ref(false)
const isSwitching = ref(false)
const revokingAccountID = ref('')
const clearingAccountID = ref('')
const refreshingAccountID = ref('')
const isAccountActionRunning = computed(() =>
  isAccountLoading.value || !!revokingAccountID.value || !!clearingAccountID.value || !!refreshingAccountID.value
)
const keychainPassphrase = ref('')
const savedKeychainPassphrase = ref('')
const isSavingKeychainPassphrase = ref(false)

// 登录表单
const showLoginModal = ref(false)
const loginForm = ref({
  email: '',
  password: ''
})
const loginEmailInputRef = ref<InstanceType<typeof AppInput> | null>(null)

// 2FA 弹窗
const show2FAModal = ref(false)
const twoFACode = ref('')
const twoFAInputRef = ref<InstanceType<typeof AppInput> | null>(null)

watch(showLoginModal, async visible => {
  if (!visible) return
  await nextTick()
  loginEmailInputRef.value?.focus()
})

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
  await loadKeychainPassphrase()
  try {
    const res = await GetAccountInfo()
    accounts.value = await GetAccounts()
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
    account.value = { id: '', name: '', email: '', region: '', active: false, success: false }
    accounts.value = await GetAccounts().catch(() => [])
    emit('accountChanged', account.value)
    emit('busyChange', false, '未登录')
    if (autoNavigate) {
      emit('navigate', 'account')
    }
  } finally {
    isAccountLoading.value = false
  }
}

async function saveKeychainPassphrase() {
  isSavingKeychainPassphrase.value = true
  emit('busyChange', true, '正在保存钥匙串密码...')
  try {
    await SetKeychainPassphrase(keychainPassphrase.value)
    savedKeychainPassphrase.value = keychainPassphrase.value
    message.success('钥匙串密码已保存')
    emit('busyChange', false, '钥匙串密码已更新')
  } catch (err: any) {
    message.error(`保存钥匙串密码失败: ${err}`)
    emit('busyChange', false, '钥匙串密码保存失败')
  } finally {
    isSavingKeychainPassphrase.value = false
  }
}

async function loadKeychainPassphrase() {
  try {
    const settings = await GetSettings()
    keychainPassphrase.value = settings.keychainPassphrase || ''
    savedKeychainPassphrase.value = keychainPassphrase.value
  } catch (err) {
    console.error('加载钥匙串密码失败:', err)
  }
}

function handleSaveKeychainPassphrase() {
  if (!keychainPassphrase.value) {
    message.warning('钥匙串密码不能为空')
    return
  }
  if (accounts.value.length > 0 && keychainPassphrase.value !== savedKeychainPassphrase.value) {
    dialog.warning({
      title: '确认修改钥匙串密码',
      content: '现有账号凭据使用当前密码加密。修改后可能无法读取已有钥匙串，需要重新登录账号。确定继续吗？',
      positiveText: '确认修改',
      negativeText: '取消',
      onPositiveClick: saveKeychainPassphrase
    })
    return
  }
  void saveKeychainPassphrase()
}

async function handleSwitchAccount(item: main.AccountInfo) {
  if (item.active || isSwitching.value) return
  isSwitching.value = true
  emit('busyChange', true, `正在切换至 ${item.email}...`)
  try {
    const res = await SwitchAccount(item.id)
    account.value = res
    accounts.value = await GetAccounts()
    localStorage.setItem('apple_vault_logged_in', 'true')
    emit('accountChanged', res)
    emit('loginSuccess', res)
    message.success(`已切换至 ${res.name || res.email}`)
    emit('busyChange', false, `当前账号: ${res.email}`)
  } catch (err: any) {
    message.error(`切换账号失败: ${err}`)
    emit('busyChange', false, '切换账号失败')
  } finally {
    isSwitching.value = false
  }
}

function openLoginModal() {
  loginForm.value = { email: '', password: '' }
  showLoginModal.value = true
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
      showLoginModal.value = false
      show2FAModal.value = true
      emit('busyChange', true, '等待输入双重认证验证码...')
      return
    }

    if (res.success) {
      localStorage.setItem('apple_vault_logged_in', 'true')
      showLoginModal.value = false
      account.value = res.account
      accounts.value = await GetAccounts()
      emit('accountChanged', res.account)
      emit('loginSuccess', res.account)
      message.success(`登录成功: ${res.account.name}`)
      loginForm.value.password = ''
      emit('busyChange', false, `登录成功: ${res.account.name}`)
      emit('navigate', 'search')
    } else {
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
      accounts.value = await GetAccounts()
      emit('accountChanged', res.account)
      emit('loginSuccess', res.account)
      message.success(`登录成功: ${res.account.name}`)
      loginForm.value.password = ''
      twoFACode.value = ''
      emit('busyChange', false, `登录成功: ${res.account.name}`)
      emit('navigate', 'search')
    } else {
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
  showLoginModal.value = true
  emit('busyChange', false, '用户取消了验证。')
}

/**
 * 注销登录凭据
 */
async function applyRemovedAccount(item: main.AccountInfo, next: main.AccountInfo) {
  accounts.value = await GetAccounts()
  if (!item.active) return
  account.value = next
  localStorage.setItem('apple_vault_logged_in', next.success ? 'true' : 'false')
  emit('accountChanged', next)
  if (next.success) {
    emit('loginSuccess', next)
  } else {
    emit('navigate', 'account')
  }
}

async function handleRevoke(item: main.AccountInfo) {
  dialog.warning({
    title: '确认退出登录',
    content: `确定要退出 ${item.email} 并清除该账号的本地授权凭证吗？`,
    positiveText: '确定退出',
    negativeText: '取消',
    onPositiveClick: async () => {
      revokingAccountID.value = item.id
      emit('busyChange', true)
      try {
        const next = await RemoveAccount(item.id)
        await applyRemovedAccount(item, next)
        message.success(`已退出 ${item.email}`)
        emit('busyChange', false, next.success ? `当前账号: ${next.email}` : '已退出登录')
      } catch (err: any) {
        message.error(`注销失败: ${err}`)
      } finally {
        revokingAccountID.value = ''
        emit('busyChange', false)
      }
    }
  })
}

/**
 * 清空本地钥匙串缓存
 */
async function handleClearKeychain(item: main.AccountInfo) {
  dialog.warning({
    title: '清空本地密钥库缓存',
    content: `此操作将删除 ${item.email} 的本地凭据与 Cookie，并从账号列表中移除。确定清理吗？`,
    positiveText: '确定清理',
    negativeText: '取消',
    onPositiveClick: async () => {
      clearingAccountID.value = item.id
      try {
        const next = await RemoveAccount(item.id)
        await applyRemovedAccount(item, next)
        message.success(`已清除 ${item.email} 的本地密钥缓存`)
      } catch (err: any) {
        message.error(`清理失败: ${err}`)
      } finally {
        clearingAccountID.value = ''
      }
    }
  })
}

async function handleRefreshAccount(item: main.AccountInfo) {
  refreshingAccountID.value = item.id
  emit('busyChange', true, `正在刷新 ${item.email}...`)
  try {
    const updated = await RefreshAccount(item.id)
    accounts.value = await GetAccounts()
    if (updated.active) {
      account.value = updated
      emit('accountChanged', updated)
    }
    message.success(`已刷新 ${updated.email} 的登录状态`)
    emit('busyChange', false, `账号状态正常: ${updated.email}`)
  } catch (err: any) {
    message.error(`刷新失败: ${err}`)
    emit('busyChange', false, `账号状态异常: ${item.email}`)
  } finally {
    refreshingAccountID.value = ''
  }
}

defineExpose({
  account,
  accounts,
  isLoggedIn,
  keychainPassphrase,
  refreshAccount
})
</script>
