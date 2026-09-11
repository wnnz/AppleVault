<template>
  <n-config-provider 
    :theme="isDark ? darkTheme : null" 
    :theme-overrides="themeOverrides"
    :locale="zhCN" 
    :date-locale="dateZhCN"
  >
    <n-dialog-provider>
      <n-notification-provider>
        <n-message-provider>
          <MainView 
            :current-theme="currentTheme" 
            :is-dark="isDark" 
            @update:current-theme="setTheme"
            @toggle-theme="toggleTheme" 
          />
        </n-message-provider>
      </n-notification-provider>
    </n-dialog-provider>
    <n-global-style />
  </n-config-provider>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  NConfigProvider,
  NMessageProvider,
  NDialogProvider,
  NNotificationProvider,
  NGlobalStyle,
  darkTheme,
  zhCN,
  dateZhCN,
  GlobalThemeOverrides
} from 'naive-ui'
import MainView from './components/MainView.vue'
import { AppTheme, THEME_STORAGE_KEY } from './types/theme'

const storedTheme = localStorage.getItem(THEME_STORAGE_KEY)
const initialTheme: AppTheme =
  storedTheme === 'minimal-dark' || storedTheme === 'handdrawn'
    ? storedTheme
    : 'minimal-light'

const currentTheme = ref<AppTheme>(initialTheme)

if (storedTheme !== initialTheme) {
  localStorage.setItem(THEME_STORAGE_KEY, initialTheme)
}

const isDark = computed(() => currentTheme.value === 'minimal-dark')
const isHanddrawn = computed(() => currentTheme.value === 'handdrawn')

function applyThemeClass(theme: AppTheme) {
  document.body.classList.remove('theme-minimal-light', 'theme-minimal-dark', 'theme-handdrawn')
  document.body.classList.add(`theme-${theme}`)
}

applyThemeClass(initialTheme)

function setTheme(theme: AppTheme) {
  currentTheme.value = theme
  localStorage.setItem(THEME_STORAGE_KEY, theme)
  applyThemeClass(theme)
}

function toggleTheme() {
  setTheme(currentTheme.value === 'minimal-light' ? 'minimal-dark' : 'minimal-light')
}

watch(currentTheme, (newTheme) => {
  localStorage.setItem(THEME_STORAGE_KEY, newTheme)
})

const themeOverrides = computed<GlobalThemeOverrides>(() => ({
  common: {
    primaryColor: isHanddrawn.value ? '#168ff0' : '#0071e3',
    primaryColorHover: isHanddrawn.value ? '#0f82de' : '#0077ed',
    primaryColorPressed: isHanddrawn.value ? '#0b70c4' : '#005bb5',
    primaryColorSuppl: isHanddrawn.value ? '#168ff0' : '#0071e3',
    borderRadius: isHanddrawn.value ? '10px' : '8px',
    fontFamily: isHanddrawn.value
      ? '"Comic Sans MS", "Kaiti SC", "STKaiti", "KaiTi", -apple-system, BlinkMacSystemFont, "PingFang SC", "Microsoft YaHei", sans-serif'
      : '-apple-system, BlinkMacSystemFont, "SF Pro Display", "SF Pro Text", "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif'
  },
  Card: {
    borderRadius: isHanddrawn.value ? '14px' : '12px',
    paddingSmall: '14px'
  },
  Button: {
    borderRadiusMedium: isHanddrawn.value ? '10px' : '8px',
    borderRadiusSmall: isHanddrawn.value ? '8px' : '6px',
    borderRadiusTiny: isHanddrawn.value ? '7px' : '5px',
    fontWeight: '500'
  },
  Input: {
    borderRadius: isHanddrawn.value ? '10px' : '8px'
  },
  DataTable: {
    borderRadius: isHanddrawn.value ? '12px' : '8px'
  }
}))
</script>
