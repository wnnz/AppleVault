import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import UnoCSS from 'unocss/vite'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue(), UnoCSS()],
  build: {
    // Keep hand-drawn SVG textures as files. Inline data URLs can prevent
    // SVG filter references (for example feTurbulence) from rendering in
    // the desktop WebView, which makes textured controls look flat.
    assetsInlineLimit: 0
  }
})
