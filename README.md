# 果仓助手 (AppleVault) `v1.0`

> 仓库地址：[https://github.com/wnnz/AppleVault](https://github.com/wnnz/AppleVault)

基于 **Go + Wails v2 + Vue 3 + Naive UI** 开发的现代化、高性能 Apple App Store 正版应用与历史版本 IPA 下载管理与真机直装工具。

---

## 📜 第三方开源项目引用与致谢

本项目作为上层图形化客户端与集成工具，底层深度依赖并引用了社区中优秀的开源项目。对以下开源项目作者及所有贡献者致以崇高的敬意：

| 项目 / 依赖 | 协议 | 仓库链接 | 在本项目中的定位与用途说明 |
| :--- | :--- | :--- | :--- |
| **ipatool** | MIT | [majd/ipatool](https://github.com/majd/ipatool) | **核心 CLI 引擎**。负责处理与 Apple App Store 服务器的通信：包括 Apple ID 账号密码登录与 2FA 双重认证、应用关键字检索、已购记录查询、历史版本构建 ID 枚举、版本元数据解析以及官方正版 IPA 包的购买与下载。 |
| **go-ios** | MIT | [danielpaulus/go-ios](https://github.com/danielpaulus/go-ios) | **设备与安装引擎**。纯 Go 实现的跨平台 iOS 设备通信工具。负责 USB/网络连接的 iOS 设备枚举识别 (`ios list --details`) 以及将有效签名的 IPA 安装到指定设备 (`ios install`)。 |
| **Wails v2** | MIT | [wailsapp/wails](https://github.com/wailsapp/wails) | **应用桌面开发框架**。负责 Go 原生能力与 Web 前端界面的双向高性能 RPC 绑定、本地原生窗口调度、无感文件拖放和系统级事件总线。 |
| **Naive UI** | MIT | [tusen-ai/naive-ui](https://github.com/tusen-ai/naive-ui) | **前端组件库**。提供优雅流畅的深色/浅色主题适配、超大容量历史版本的虚拟滚动表格 (`n-data-table`)、模态交互弹窗与通知反馈系统。 |
| **Vue 3** | MIT | [vuejs/core](https://github.com/vuejs/core) | **前端核心框架**。提供响应式状态驱动与组件化界面开发能力。 |

### 外部二进制解耦说明
本项目对 `ipatool.exe` 和 `ios.exe` 采用**免安装外部解耦设计**，而非源码侵入式合并或二进制打包篡改：
- 主程序运行时通过动态查找同目录、工作目录或系统 PATH 来调用对应的命令行工具。
- 使得外部工具链可独立随官方更新替换升级，保持运行环境的高透明性与轻量性。

### 许可证副本与免责声明
- 第三方组件的版权、商标与开源许可均归其各自原作者所有。相关开源许可证完整文本见本仓库 [`third_party`](third_party/) 目录（如 [`ipatool-LICENSE.txt`](third_party/ipatool-LICENSE.txt) 与 [`go-ios-LICENSE.txt`](third_party/go-ios-LICENSE.txt)）。
- 本项目仅为第三方独立开源图形界面工具，与 Apple Inc. 无任何隶属、赞助或官方合作关系。使用本项目请遵守 Apple 开发者协议及相关法律法规。

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
  - **🎯 自动查询指定版本**：输入目标版本号自动从新到旧逐个扫描并匹配，找到后即刻停下并高亮定位；
  - **智能互斥与随时中止**：批量查询与指定版本查询互斥防并发，进行时按钮动态变为「⏹️ 停止查询」，支持随时手动中断；
  - 选定版本一键发送到下载中心。
- **⬇️ 下载中心**：支持最新版或指定历史构建 ID 下载，支持自动获取免费购买凭据（`--purchase`）。
- **📦 已购应用**：分页查看当前 Apple ID 账号拥有的所有已购应用列表。
- **📲 IPA 安装**：支持点击选择或拖放 IPA，检测已连接的苹果设备，并将具有有效签名的 IPA 安装到指定设备。
- **🌐 网络代理支持**：支持配置 HTTP / SOCKS5 代理，底层自动注入环境变量，支持一键连通性测试。
- **⚙️ 无感密码注入**：自动管理 `--keychain-passphrase`，彻底解决 Windows 终端卡死与交互 Bug。
- **🖥️ 实时日志**：内置终端实时滚动显示执行命令与输出日志，支持中途取消。

---

## 🛠️ 编译与运行

### 环境要求
- Windows 10 / 11 (内置 WebView2)
- Go 1.23+ / 1.26+
- Node.js & pnpm / npm
- Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- `ipatool.exe` 与 `ios.exe` 放置在与免安装主程序同一目录下，或配置在系统环境变量 PATH 中
- Apple Mobile Device USB 驱动（可通过 Apple Devices 或 iTunes 安装）

设备首次连接时需要解锁并在设备上选择“信任此电脑”。iOS 17 及更高版本的部分设备通信场景可能需要先启动 go-ios tunnel，具体以底部日志提示为准。

### 便携部署与目录结构
本程序支持完全绿色便携运行，所有数据与依赖彼此隔离，免安装即开即用：
```text
build/bin/
├── AppleVault.exe            # 主程序
├── tools/                    # 外部依赖独立目录（推荐）
│   ├── ipatool.exe           # App Store 认证与下载命令行
│   └── ios.exe               # 苹果设备管理与安装命令行
└── data/                     # 运行时自动生成的全量数据目录
    ├── settings.json         # 软件偏好与代理配置
    ├── tasks.json            # 下载任务持久化列表（重启不丢失）
    └── .ipatool/             # 账号登录状态与本地安全钥匙串
```
> 注：依赖 exe 同时向前兼容放置于 `tools/`、`bin/`、程序同目录或系统 PATH 中。

### 开发模式
```bash
wails dev
```

### 生产发布编译（不清空 build/bin）
请**不要使用**带有 `-clean` 参数的构建命令（如 `wails build -clean`），否则会清除 `build/bin` 内已放置的依赖与数据。
使用以下增量编译命令或一键发布脚本：
```powershell
# 方式 1：执行根目录自带的发布脚本（自动检查前端与 tools 依赖，免清空）
.\build_release.ps1

# 方式 2：标准 Wails 构建（不带 -clean）
wails build
```
