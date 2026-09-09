export type AppTheme = 'minimal-light' | 'minimal-dark' | 'handdrawn-light' | 'handdrawn-dark'

export interface ThemeOption {
  key: AppTheme
  name: string
  category: 'minimal' | 'handdrawn'
  categoryName: string
  description: string
  preview: {
    bg: string
    card: string
    border: string
    accent: string
    text: string
  }
}

export const THEME_STORAGE_KEY = 'apple_vault_theme'
