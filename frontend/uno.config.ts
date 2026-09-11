import { defineConfig, presetUno } from 'unocss'

export default defineConfig({
  presets: [presetUno()],
  shortcuts: {
    'ui-row': 'flex items-center',
    'ui-control-row': 'flex items-center gap-2',
    'ui-fill': 'w-full h-full',
    'ui-truncate': 'overflow-hidden whitespace-nowrap text-ellipsis'
  }
})
