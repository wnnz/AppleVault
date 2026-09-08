# IPATool GUI

基于 **Go + Wails v2 + Vue 3 + Naive UI** 开发的现代化、高性能 Apple App Store 正版应用与历史版本 IPA 下载管理工具。

## 🙏 致谢与依赖

本项目使用了开源项目 [majd/ipatool](https://github.com/majd/ipatool) 提供的 App Store 账号认证、应用搜索、已购应用查询及 IPA 下载等底层能力，并在此基础上提供 Windows 图形界面、下载任务管理和交互功能。

感谢 `ipatool` 项目作者及所有贡献者。本项目是独立的第三方 GUI 项目，与 Apple Inc. 无隶属或官方关联；`ipatool` 的版权和许可遵循其原项目声明。

---

## 🌟 核心特性

- **⚡ 极致轻量与性能**：采用 Wails v2 + Go 构建，内存占用极低（~30MB），告别传统 GUI 的卡死与进程死锁问题。
- **🎨 现代优雅 UI**：基于 **Naive UI**，支持一键切换**深色模式（Dark Mode）**与明亮模式，排版精致。
- **👤 账号管理**：支持 Apple ID 登录、状态常驻展示、一键退出登录、一键清理本地损坏密钥缓存。
- **🔐 智能双重认证**：登录触发 Apple 2FA 时，自动弹出居中模态验证窗口，输入 6 位验证码秒速提交。
- **🔍 应用搜索**：支持关键字搜索 App Store，支持筛选平台（iPhone / iPad / Apple TV / VisionOS）。
- **📜 历史版本下载**：
  - 输入 Bundle ID 获取全部历史版本构建号；
  - **虚拟滚动表格**：面对 300+ 个历史版本极速渲染，滑动无比顺畅；
  - 支持按具体版本号（如 `10.2.96`）、体积或构建 ID 实时过滤筛选；
  - **毫秒级查询体积与版本详情**：无需下载完整包即可解析具体版本号、发布日期与精准文件体积（MB/GB）；
  - 选定版本一键发送到下载中心。
- **⬇️ 下载中心**：支持最新版或指定历史构建 ID 下载，支持自动获取免费购买凭据（`--purchase`）。
- **📦 已购应用**：分页查看当前 Apple ID 账号拥有的所有已购应用列表。
- **🌐 网络代理支持**：支持配置 HTTP / SOCKS5 代理，底层自动注入环境变量，支持一键连通性测试。
- **⚙️ 无感密码注入**：自动管理 `--keychain-passphrase`，彻底解决 Windows 终端卡死与交互 Bug。
- **🖥️ 实时日志**：内置终端实时滚动显示执行命令与输出日志，支持中途取消。

---

## 🛠️ 编译与运行

### 环境要求
- Windows 10 / 11 (内置 WebView2)
- Go 1.23+ / 1.26+
- Node.js & pnpm
- Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### 开发模式
```bash
wails dev
```

### 生产编译打包
```bash
wails build
```
编译产物位于 `build/bin/IPAToolGUI.exe`。
