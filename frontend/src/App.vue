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

// 从 localStorage 读取存储的主题，默认 'minimal-light'
const savedTheme = (localStorage.getItem(THEME_STORAGE_KEY) as AppTheme) || 'minimal-light'
const validThemes: AppTheme[] = ['minimal-light', 'minimal-dark', 'handdrawn-light', 'handdrawn-dark']

const currentTheme = ref<AppTheme>(
  validThemes.includes(savedTheme) ? savedTheme : 'minimal-light'
)

const isDark = computed(() => currentTheme.value.endsWith('-dark'))

function setTheme(theme: AppTheme) {
  currentTheme.value = theme
  localStorage.setItem(THEME_STORAGE_KEY, theme)
}

function toggleTheme() {
  if (currentTheme.value === 'minimal-light') {
    setTheme('minimal-dark')
  } else if (currentTheme.value === 'minimal-dark') {
    setTheme('minimal-light')
  } else if (currentTheme.value === 'handdrawn-light') {
    setTheme('handdrawn-dark')
  } else if (currentTheme.value === 'handdrawn-dark') {
    setTheme('handdrawn-light')
  }
}

watch(currentTheme, (newTheme) => {
  localStorage.setItem(THEME_STORAGE_KEY, newTheme)
})

const themeOverrides = computed<GlobalThemeOverrides>(() => {
  // 手绘风格 - 浅色 (水彩手账插画风)
  if (currentTheme.value === 'handdrawn-light') {
    return {
      common: {
        primaryColor: '#3f8fce',
        primaryColorHover: '#62a9dc',
        primaryColorPressed: '#2f75ad',
        primaryColorSuppl: '#3f8fce',
        successColor: '#42b88a',
        warningColor: '#e8aa46',
        errorColor: '#ef786e',
        borderRadius: '8px',
        fontFamily: '"Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif'
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

  // 手绘风格 - 深色 (夜空水彩插画风)
  if (currentTheme.value === 'handdrawn-dark') {
    return {
      common: {
        primaryColor: '#6aaed8',
        primaryColorHover: '#8bc5e6',
        primaryColorPressed: '#4e91bc',
        primaryColorSuppl: '#6aaed8',
        successColor: '#58c49a',
        warningColor: '#e5b45c',
        errorColor: '#ef8279',
        borderRadius: '8px',
        fontFamily: '"Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif'
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

  // 简约风格 (minimal-light 和 minimal-dark)
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
