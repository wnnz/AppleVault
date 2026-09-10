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
const initialTheme: AppTheme = (storedTheme === 'sketch-light' || storedTheme === 'minimal-dark')
  ? storedTheme
  : 'sketch-light'

const currentTheme = ref<AppTheme>(initialTheme)

if (storedTheme !== initialTheme) {
  localStorage.setItem(THEME_STORAGE_KEY, initialTheme)
}

const isDark = computed(() => currentTheme.value === 'minimal-dark')
const isSketch = computed(() => currentTheme.value === 'sketch-light')

function setTheme(theme: AppTheme) {
  currentTheme.value = theme
  localStorage.setItem(THEME_STORAGE_KEY, theme)
}

function toggleTheme() {
  if (currentTheme.value === 'sketch-light') {
    setTheme('minimal-dark')
  } else if (currentTheme.value === 'minimal-dark') {
    setTheme('minimal-light')
  } else {
    setTheme('sketch-light')
  }
}

watch(currentTheme, (newTheme) => {
  localStorage.setItem(THEME_STORAGE_KEY, newTheme)
})

const themeOverrides = computed<GlobalThemeOverrides>(() => {
  if (isSketch.value) {
    return {
      common: {
        primaryColor: '#0284c7',
        primaryColorHover: '#0369a1',
        primaryColorPressed: '#075985',
        primaryColorSuppl: '#0284c7',
        borderRadius: '10px',
        fontFamily: '"Chalkboard SE", "Comic Sans MS", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif'
      },
      Card: {
        borderRadius: '14px',
        paddingSmall: '14px'
      },
      Button: {
        borderRadiusMedium: '10px',
        borderRadiusSmall: '8px',
        borderRadiusTiny: '6px',
        fontWeight: '600'
      },
      Input: {
        borderRadius: '10px'
      },
      DataTable: {
        borderRadius: '12px'
      }
    }
  }

  return {
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
  }
})
</script>
