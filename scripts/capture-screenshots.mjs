import fs from 'node:fs'
import path from 'node:path'

const TABS = [
  { tab: 'search', file: '01_search.png' },
  { tab: 'versions', file: '02_versions.png' },
  { tab: 'download', file: '03_download.png' },
  { tab: 'purchased', file: '04_purchased.png' },
  { tab: 'installer', file: '05_installer.png' },
  { tab: 'account', file: '06_account.png' },
  { tab: 'settings', file: '07_settings.png' },
  { tab: 'about', file: '08_about.png' }
]

const VIEWPORT_WIDTH = 1104
const VIEWPORT_HEIGHT = 760
const OUTPUT_DIR = path.resolve('docs/images')

async function createCdpClient() {
  const versionRes = await fetch('http://127.0.0.1:9222/json/version')
  const versionData = await versionRes.json()
  
  // 创建新页面
  const newTabRes = await fetch('http://127.0.0.1:9222/json/new?about:blank', { method: 'PUT' })
  const newTabData = await newTabRes.json()
  const wsUrl = newTabData.webSocketDebuggerUrl

  const ws = new WebSocket(wsUrl)
  let id = 1
  const pending = new Map()

  await new Promise((resolve, reject) => {
    ws.onopen = resolve
    ws.onerror = reject
  })

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data)
    if (data.id && pending.has(data.id)) {
      const { resolve, reject } = pending.get(data.id)
      pending.delete(data.id)
      if (data.error) {
        reject(data.error)
      } else {
        resolve(data.result)
      }
    }
  }

  function send(method, params = {}) {
    return new Promise((resolve, reject) => {
      const msgId = id++
      pending.set(msgId, { resolve, reject })
      ws.send(JSON.stringify({ id: msgId, method, params }))
    })
  }

  async function close() {
    ws.close()
    await fetch(`http://127.0.0.1:9222/json/close/${newTabData.id}`)
  }

  return { send, close }
}

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

async function run() {
  console.log('Connecting to Chrome CDP...')
  const client = await createCdpClient()

  try {
    await client.send('Page.enable')
    await client.send('DOM.enable')
    await client.send('Network.enable')
    await client.send('Emulation.setDeviceMetricsOverride', {
      width: VIEWPORT_WIDTH,
      height: VIEWPORT_HEIGHT,
      deviceScaleFactor: 1,
      mobile: false
    })

    if (!fs.existsSync(OUTPUT_DIR)) {
      fs.mkdirSync(OUTPUT_DIR, { recursive: true })
    }

    for (const item of TABS) {
      const targetUrl = `http://127.0.0.1:4173/?demo=1&theme=handdrawn&tab=${item.tab}`
      console.log(`Navigating to ${targetUrl}...`)
      await client.send('Page.navigate', { url: targetUrl })

      // 等待页面加载完成
      await sleep(1200)

      // 确保设置主题为 handdrawn
      await client.send('Runtime.evaluate', {
        expression: `
          (() => {
            localStorage.setItem('apple_vault_theme', 'handdrawn');
            document.documentElement.dataset.theme = 'handdrawn';
            document.body.classList.remove('theme-minimal-light', 'theme-minimal-dark');
            document.body.classList.add('theme-handdrawn');
          })()
        `
      })

      // 等待字体及样式应用完毕
      await sleep(800)

      console.log(`Capturing screenshot for ${item.file}...`)
      const result = await client.send('Page.captureScreenshot', {
        format: 'png',
        captureBeyondViewport: false
      })

      const buffer = Buffer.from(result.data, 'base64')
      const outputPath = path.join(OUTPUT_DIR, item.file)
      fs.writeFileSync(outputPath, buffer)
      console.log(`Saved: ${outputPath} (${buffer.length} bytes)`)
    }

    console.log('All screenshots captured successfully!')
  } finally {
    await client.close()
  }
}

run().catch((err) => {
  console.error('Failed to capture screenshots:', err)
  process.exit(1)
})
