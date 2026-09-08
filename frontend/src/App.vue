<template>
  <n-config-provider :theme="isDark ? darkTheme : null" :locale="zhCN" :date-locale="dateZhCN">
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
import { ref } from 'vue'
import {
  NConfigProvider,
  NMessageProvider,
  NDialogProvider,
  NNotificationProvider,
  NGlobalStyle,
  darkTheme,
  zhCN,
  dateZhCN
} from 'naive-ui'
import MainView from './components/MainView.vue'

const isDark = ref(false)

function toggleTheme() {
  isDark.value = !isDark.value
}
</script>
