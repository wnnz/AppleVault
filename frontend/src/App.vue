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
          <MainView :is-dark="isDark" @toggle-theme="toggleTheme" />
        </n-message-provider>
      </n-notification-provider>
    </n-dialog-provider>
    <n-global-style />
  </n-config-provider>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
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

const isDark = ref(false)

function toggleTheme() {
  isDark.value = !isDark.value
}

const themeOverrides = computed<GlobalThemeOverrides>(() => ({
  common: {
    primaryColor: '#0078D4',
    primaryColorHover: '#106EBE',
    primaryColorPressed: '#005A9E',
    primaryColorSuppl: '#0078D4',
    borderRadius: '4px',
    fontFamily: 'Segoe UI, "Microsoft YaHei UI", sans-serif'
  },
  Card: {
    borderRadius: '8px',
    paddingSmall: '16px'
  },
  Button: {
    borderRadiusMedium: '4px',
    borderRadiusSmall: '4px',
    borderRadiusTiny: '4px',
    fontWeight: '600'
  },
  Input: {
    borderRadius: '4px'
  }
}))
</script>
