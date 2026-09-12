import { ref, nextTick } from 'vue'

export function useConsoleLogs() {
  const showLogs = ref<boolean>(localStorage.getItem('apple_vault_show_logs') === 'true')
  const logLines = ref<string[]>([])
  const logContainerRef = ref<HTMLElement | null>(null)

  function toggleLogs() {
    showLogs.value = !showLogs.value
    localStorage.setItem('apple_vault_show_logs', String(showLogs.value))
  }

  function appendLog(line: string) {
    logLines.value.push(line)
    if (logLines.value.length > 600) {
      logLines.value.shift()
    }
    nextTick(() => {
      if (logContainerRef.value) {
        logContainerRef.value.scrollTop = logContainerRef.value.scrollHeight
      }
    })
  }

  function clearLogs() {
    logLines.value = []
  }

  return {
    showLogs,
    logLines,
    logContainerRef,
    toggleLogs,
    appendLog,
    clearLogs
  }
}
