// 浏览器开发与预览环境下的 Wails 接口垫片
if (typeof window !== 'undefined') {
  if (!(window as any).runtime) {
    ;(window as any).runtime = {
      EventsOn: () => () => {},
      EventsOnMultiple: () => () => {},
      OnFileDrop: () => {},
      OnFileDropOff: () => {},
      BrowserOpenURL: (url: string) => window.open(url, '_blank'),
      LogPrint: () => {},
      LogTrace: () => {},
      LogDebug: () => {},
      LogInfo: () => {},
      LogWarning: () => {},
      LogError: () => {},
      LogFatal: () => {},
      WindowSetTitle: () => {},
      WindowFullscreen: () => {},
      WindowUnfullscreen: () => {},
      WindowSetSize: () => {},
      WindowGetSize: () => Promise.resolve({ w: 1024, h: 768 }),
      WindowSetPosition: () => {},
      WindowGetPosition: () => Promise.resolve({ x: 0, y: 0 }),
      WindowHide: () => {},
      WindowShow: () => {},
      WindowCenter: () => {},
      Quit: () => {}
    }
  }
  if (!(window as any).go) {
    ;(window as any).go = {
      backend: {
        App: new Proxy({}, {
          get: (_target, prop) => {
            return (..._args: any[]) => {
              console.log(`[Wails Mock] backend.App.${String(prop)} called`)
              return Promise.resolve({})
            }
          }
        })
      }
    }
  }
}

import { createApp } from 'vue'
import App from './App.vue'
import 'virtual:uno.css'
import './style.css'
import './ui/themes/index.css'

createApp(App).mount('#app')
