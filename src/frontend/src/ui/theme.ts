import { computed, ref, watch } from 'vue'

export type AppTheme = 'minimal-light' | 'minimal-dark' | 'handdrawn'
export const THEME_STORAGE_KEY = 'apple_vault_theme'

export interface ThemeOption {
  key: AppTheme
  name: string
  description: string
  recommended?: boolean
  preview: { bg: string; card: string; border: string; accent: string; text: string }
}

export const themeOptions: ThemeOption[] = [
  {
    key: 'minimal-light', name: '极简浅色', description: '明亮清爽，适合日常使用',
    preview: { bg: '#f8fafc', card: '#ffffff', border: '#e2e8f0', accent: '#0071e3', text: '#1e293b' }
  },
  {
    key: 'minimal-dark', name: '极简深色', description: '深色护眼，适合夜间使用',
    preview: { bg: '#0f172a', card: '#1e293b', border: '#334155', accent: '#0284c7', text: '#f1f5f9' }
  },
  {
    key: 'handdrawn', name: '纸间手绘', description: '温润纸感与手绘插画主题', recommended: true,
    preview: { bg: '#fffdf6', card: '#f4fbff', border: '#b8e1f8', accent: '#168ff0', text: '#173b6a' }
  }
]

function normalizeTheme(value: string | null): AppTheme {
  return value === 'minimal-dark' || value === 'handdrawn' ? value : 'minimal-light'
}

export function applyTheme(theme: AppTheme) {
  document.documentElement.dataset.theme = theme
  document.body.classList.remove(...themeOptions.map(item => `theme-${item.key}`))
  document.body.classList.add(`theme-${theme}`)
}

export function useAppTheme() {
  const urlParams = typeof window !== 'undefined' ? new URLSearchParams(window.location.search) : null
  const urlTheme = urlParams?.get('theme')
  const initialTheme = normalizeTheme(urlTheme || localStorage.getItem(THEME_STORAGE_KEY))
  const currentTheme = ref<AppTheme>(initialTheme)
  const isDark = computed(() => currentTheme.value === 'minimal-dark')
  applyTheme(currentTheme.value)
  watch(currentTheme, theme => {
    localStorage.setItem(THEME_STORAGE_KEY, theme)
    applyTheme(theme)
  })
  const setTheme = (theme: AppTheme) => { currentTheme.value = theme }
  const toggleTheme = () => setTheme(currentTheme.value === 'minimal-light' ? 'minimal-dark' : 'minimal-light')
  return { currentTheme, isDark, setTheme, toggleTheme }
}
