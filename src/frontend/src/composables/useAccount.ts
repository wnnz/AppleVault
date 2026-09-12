import { ref, computed } from 'vue'
import { backend as main } from '../../wailsjs/go/models'
import {
  GetAccountInfo,
  Login,
  Revoke,
  ClearKeychainCache
} from '../../wailsjs/go/backend/App'
import { useAppDialog, useAppMessage } from '../ui/feedback'

export function useAccount(options?: {
  onSuccessLogin?: () => void
  onNeedNavigate?: (tab: string) => void
  onSettingsReload?: () => Promise<void>
  onBusyChange?: (busy: boolean, text?: string) => void
}) {
  const message = useAppMessage()
  const dialog = useAppDialog()

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

  const show2FAModal = ref(false)
  const twoFACode = ref('')

  async function refreshAccount(autoNavigate = false) {
    isAccountLoading.value = true
    options?.onBusyChange?.(false, '正在获取账号信息...')
    try {
      const res = await GetAccountInfo()
      account.value = res
      if (res.success && res.email) {
        options?.onBusyChange?.(false, `已登录: ${res.name} (${res.email})`)
        await options?.onSettingsReload?.()
        if (autoNavigate) {
          options?.onNeedNavigate?.('search')
        }
      } else {
        options?.onBusyChange?.(false, '未检测到已登录的 Apple ID')
        options?.onNeedNavigate?.('account')
      }
    } catch (err: any) {
      account.value = { name: '', email: '', success: false }
      options?.onBusyChange?.(false, '未登录')
      options?.onNeedNavigate?.('account')
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
    options?.onBusyChange?.(true, '正在验证 Apple ID 账号与密码...')

    try {
      const res = await Login(loginForm.value.email, loginForm.value.password, '')
      if (res.requires2FA) {
        twoFACode.value = ''
        show2FAModal.value = true
        options?.onBusyChange?.(true, '等待输入双重认证验证码...')
        return
      }

      if (res.success) {
        account.value = res.account
        message.success(`登录成功: ${res.account.name}`)
        loginForm.value.password = ''
        options?.onBusyChange?.(false, `登录成功: ${res.account.name}`)
        await options?.onSettingsReload?.()
        options?.onSuccessLogin?.()
      } else {
        message.error(`登录失败: ${res.errorMessage}`)
        options?.onBusyChange?.(false, `登录失败: ${res.errorMessage}`)
      }
    } catch (err: any) {
      message.error(`执行异常: ${err}`)
    } finally {
      isLoggingIn.value = false
      if (!show2FAModal.value) {
        options?.onBusyChange?.(false)
      }
    }
  }

  async function confirm2FA() {
    if (twoFACode.value.length !== 6) {
      message.warning('请输入正确的 6 位数字验证码！')
      return
    }

    isLoggingIn.value = true
    options?.onBusyChange?.(true, '正在提交验证码并完成登录...')

    try {
      const res = await Login(loginForm.value.email, loginForm.value.password, twoFACode.value)
      if (res.success) {
        show2FAModal.value = false
        account.value = res.account
        message.success(`登录成功: ${res.account.name}`)
        loginForm.value.password = ''
        twoFACode.value = ''
        options?.onBusyChange?.(false, `登录成功: ${res.account.name}`)
        await options?.onSettingsReload?.()
        options?.onSuccessLogin?.()
      } else {
        message.error(`验证失败: ${res.errorMessage}`)
        options?.onBusyChange?.(false, `验证失败: ${res.errorMessage}`)
      }
    } catch (err: any) {
      message.error(`验证异常: ${err}`)
    } finally {
      isLoggingIn.value = false
      options?.onBusyChange?.(false)
    }
  }

  function cancel2FA() {
    show2FAModal.value = false
    twoFACode.value = ''
    options?.onBusyChange?.(false, '用户取消了验证。')
  }

  async function handleRevoke() {
    dialog.warning({
      title: '确认退出登录',
      content: '确定要退出当前 Apple ID 账号并清除本地授权凭证吗？',
      positiveText: '确定退出',
      negativeText: '取消',
      onPositiveClick: async () => {
        isRevoking.value = true
        options?.onBusyChange?.(true)
        try {
          const ok = await Revoke()
          if (ok) {
            account.value = { name: '', email: '', success: false }
            message.success('已成功注销登录凭据')
            options?.onBusyChange?.(false, '已退出登录')
            await options?.onSettingsReload?.()
            options?.onNeedNavigate?.('account')
          }
        } catch (err: any) {
          message.error(`注销失败: ${err}`)
        } finally {
          isRevoking.value = false
          options?.onBusyChange?.(false)
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

  return {
    account,
    isLoggedIn,
    isAccountLoading,
    isRevoking,
    isClearing,
    isLoggingIn,
    loginForm,
    show2FAModal,
    twoFACode,
    refreshAccount,
    handleLogin,
    confirm2FA,
    cancel2FA,
    handleRevoke,
    handleClearKeychain
  }
}
