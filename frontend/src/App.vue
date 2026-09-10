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
const initialTheme: AppTheme = storedTheme === 'minimal-dark' ? 'minimal-dark' : 'minimal-light'

const currentTheme = ref<AppTheme>(initialTheme)

if (storedTheme !== initialTheme) {
  localStorage.setItem(THEME_STORAGE_KEY, initialTheme)
}

const isDark = computed(() => currentTheme.value === 'minimal-dark')

function setTheme(theme: AppTheme) {
  currentTheme.value = theme
  localStorage.setItem(THEME_STORAGE_KEY, theme)
}

function toggleTheme() {
  setTheme(currentTheme.value === 'minimal-light' ? 'minimal-dark' : 'minimal-light')
}

watch(currentTheme, (newTheme) => {
  localStorage.setItem(THEME_STORAGE_KEY, newTheme)
})

const themeOverrides = computed<GlobalThemeOverrides>(() => ({
  common: {
    primaryColor: '#0071e3',
    primaryColorHover: '#0077ed',
    primaryColorPressed: '#005bb5',
    primaryColorSuppl: '#0071e3',
    borderRadius: '8px',
    fontFamily: '-apple-system, BlinkMacSystemFont, "SF Pro Display", "SF Pro Text", "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif'
  },
  Card: {
    borderRadius: '12px',
    paddingSmall: '14px'
  },
  Button: {
    borderRadiusMedium: '8px',
    borderRadiusSmall: '6px',
    borderRadiusTiny: '5px',
    fontWeight: '500'
  },
  Input: {
    borderRadius: '8px'
  },
  DataTable: {
    borderRadius: '8px'
  }
}))
</script>
