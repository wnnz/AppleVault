<template>
  <MainView
    :current-theme="currentTheme"
    :is-dark="isDark"
    @update:current-theme="setTheme"
  />
  <AppDialogHost />
  <AppToastHost />
</template>

<script setup lang="ts">
import MainView from './components/MainView.vue'
import AppDialogHost from './ui/components/AppDialogHost.vue'
import AppToastHost from './ui/components/AppToastHost.vue'
import { useAppTheme } from './ui/theme'

const { currentTheme, isDark, setTheme } = useAppTheme()
</script>
